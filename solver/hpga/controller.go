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
		tt.WorstValue(),
		inst.NewSolution(),
	}

	for i := 0; i < opts.NIslands; i++ {
		c.parent.toChildren[i] = newIsland(i, inst, fromChildren, opts)
	}

	return c
}

// Run the controller.
func (c *controller) run(timeoutInterval int) (*tt.Solution, tt.Value) {
	// Wait for islands to signal that their children have finished generating populations
	c.wait()

	log.Println("Population generation finished")

	timeout := time.After(time.Duration(timeoutInterval) * time.Minute)
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)

msgLoop:
	for {
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

					if c.topValue.IsIdeal() {
						log.Println("Found ideal solution. Stopping...")
						c.stopChildren()
						break msgLoop
					}

					for child := range c.toChildren {
						if child != msg.source {
							c.sendToChild(child, valueMessage{c.topValue})
						}
					}
				}
			}
		case <-timeout:
			log.Println("Timeout: stopping...")
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
