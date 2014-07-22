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
		inst.NewSolution(),
	}

	for i := 0; i < opts.NIslands; i++ {
		c.parent.toChildren[i] = newIsland(i, inst, fromChildren, opts)
	}

	return c
}

// Run the controller.
func (c *controller) run(timeout int) (*tt.Solution, tt.Value) {
	// Wait for islands to signal that their children have finished generating populations
	c.wait()

	log.Println("Population generation finished")

	// Use most-constrained variable ordering to find an upper bound for the
	// HPGA to work towards. This way we will only try to update the best-
	// known solution when this one is beat.
	heuristics.MostConstrainedOrdering(c.top)
	c.topValue = c.top.Value()

	log.Printf("Found new best solution: %s\n", c.topValue)

	for child := range c.toChildren {
		c.sendToChild(child, valueMessage{c.topValue})
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)
msgLoop:
	for {
		timeout := time.After(time.Duration(timeout) * time.Minute)
		select {
		case msg := <-c.fromChildren:
			switch msg.messageType() {
			case solutionMessageType:
				soln := msg.content.(solutionMessage).soln
				value := msg.content.(solutionMessage).value

				if value.Less(c.topValue) {
					c.topValue = value
					c.top.Free()
					c.top = c.inst.SolutionFromRats(soln)
					log.Printf("Found new best solution: %s\n", c.topValue)

					for child := range c.toChildren {
						if child != msg.source {
							c.sendToChild(child, valueMessage{c.topValue})
						}
					}
				}
			}
		case <-timeout:
			log.Println("Timeout: sending stopMessageType to all islands...")
			c.stopChildren()
			break msgLoop

		case <-signals:
			log.Println("Caught interrupt")
			break msgLoop
		}
	}

	return c.top, c.topValue
}

// Send a stopMessageType message to all islands under the controller and wait for
// a finMessageType message from each of them. If a solutionMessageType message arrives,
// it will be processed as normal (i.e., updating c.top and c.topValue if
// better than the current solution).
func (c *controller) stopChildren() {
	for child := range c.toChildren {
		c.sendToChild(child, stopMessage{})
	}

	finished := make(map[int]bool)

	for len(finished) != len(c.toChildren) {
		msg := <-c.fromChildren

		switch msg.messageType() {
		case finMessageType:
			finished[msg.source] = true

		case solutionMessageType:
			value := msg.content.(solutionMessage).value

			if value.Less(c.topValue) {
				soln := msg.content.(solutionMessage).soln
				c.topValue = value
				c.top.Free()
				c.top = c.inst.SolutionFromRats(soln)
				log.Printf("Found new best solution: %s\n", c.topValue)
			}
		}
	}
}
