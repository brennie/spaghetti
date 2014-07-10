// spaghetti: Applying Hierarchical Parallel Genetic Algorithms to solve the
// University Timetabling Problem.
// Copyright (C) 2014  Barret Rennie
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package hpga

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/brennie/spaghetti/options"
	"github.com/brennie/spaghetti/solver/heuristics"
	"github.com/brennie/spaghetti/tt"
)

// A controller is just a parent; its children are the islands.
type controller struct {
	parent
	inst     *tt.Instance // The timetabling instance
	topValue tt.Value     // The value of the top-valued solution.
	top      *tt.Solution // The top-valued solution
	verbose  bool         // Determines if events should be logged.
}

// Create a new controller. There will be nIslands islands, each with nSlaves
// slaves.
func newController(inst *tt.Instance, opts options.SolveOptions) *controller {
	fromChildren := make(chan message, 5)

	c := &controller{
		parent{
			fromChildren,
			make([]chan<- message, opts.NIslands),
		},
		inst,
		tt.Value{-1, -1},
		nil,
		opts.Verbose,
	}

	for i := 0; i < opts.NIslands; i++ {
		c.parent.toChildren[i] = newIsland(i, inst, fromChildren, opts)
	}

	return c
}

// Optionally log a message if the verbose flag is set.
func (c *controller) log(format string, args ...interface{}) {
	if c.verbose {
		msg := fmt.Sprintf(format, args...)
		log.Printf("controller: %s\n", msg)
	}
}

// Run the controller.
func (c *controller) run(timeout int) (*tt.Solution, tt.Value) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	c.top = c.inst.NewSolution() // The top-valued solution over the whole HPGA.

	heuristics.MostConstrainedOrdering(c.top)
	c.topValue = c.top.Value() // The value of the top-valued solution over the whole HPGA.

	for child := range c.toChildren {
		c.sendToChild(child, valueMsgType, c.topValue)
	}

msgLoop:
	for {
		timeout := time.After(time.Duration(timeout) * time.Minute)
		select {
		case msg := <-c.fromChildren:
			switch msg.MsgType() {
			case solnMsgType:
				best, value := msg.(solnMessage).soln, msg.(solnMessage).value

				if value.Less(c.topValue) {
					c.log("received solnMsgType: value was better")
					c.topValue = value
					c.top.Free()
					c.top = c.inst.SolutionFromRats(best)

					for child := range c.toChildren {
						if child != msg.Source() {
							c.sendToChild(child, valueMsgType, c.topValue)
						}
					}
				} else {
					c.log("received solnMsgType: value was worse")
				}
			}
		case <-timeout:
			c.log("timeout: sending stopMsgType to all islands")
			c.stopChildren()
			c.log("received finMsgType from all islands: exiting")
			break msgLoop

		case <-signals:
			log.Println("caught interrupt")
			break msgLoop
		}
	}

	return c.top, c.topValue
}

// Send a stopMsgType message to all islands under the controller and wait for
// a finMsgType message from each of them. If a solnMsgType message arrives,
// it will be processed as normal (i.e., updating c.top and c.topValue if
// better than the current solution).
func (c *controller) stopChildren() {
	for child := range c.toChildren {
		c.sendToChild(child, stopMsgType)
	}

	finished := make(map[int]bool)

	for len(finished) != len(c.toChildren) {
		msg := <-c.fromChildren

		switch msg.MsgType() {
		case finMsgType:
			finished[msg.Source()] = true

		case solnMsgType:
			value := msg.(solnMessage).value

			if value.Less(c.topValue) {
				soln := msg.(solnMessage).soln
				c.topValue = value
				c.top.Free()
				c.top = c.inst.SolutionFromRats(soln)
			}
		}
	}
}
