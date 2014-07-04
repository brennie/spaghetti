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
	"math/rand"
	"time"

	"github.com/brennie/spaghetti/solver/heuristics"
	"github.com/brennie/spaghetti/tt"
)

// A controller is just a parent; its children are the islands.
type controller struct {
	parent
	inst    *tt.Instance // The timetabling instance
	verbose bool         // Determines if events should be logged.
}

// Create a new controller. There will be nIslands islands, each with nSlaves
// slaves.
func newController(nIslands, nSlaves int, inst *tt.Instance, verbose bool) *controller {
	comm := make(chan message, 5)

	c := &controller{
		parent{
			comm,
			make([]chan<- message, nIslands),
		},
		inst,
		verbose,
	}

	for i := 0; i < nIslands; i++ {
		c.parent.toChildren[i] = newIsland(i, nSlaves, inst, rand.Int63(), comm, verbose)
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
func (c *controller) run(timeout int) *tt.Solution {
	top := c.inst.NewSolution() // The top-valued solution over the whole HPGA.

	heuristics.MostConstrainedOrdering(top)
	topValue := top.Value() // The value of the top-valued solution over the whole HPGA.

	for child := range c.toChildren {
		c.sendToChild(child, valueMsgType, topValue)
	}

msgLoop:
	for {
		timeout := time.After(time.Duration(timeout) * time.Minute)
		select {
		case msg := <-c.fromChildren:
			switch msg.MsgType() {
			case solnMsgType:
				best, value := msg.(solnMessage).soln, msg.(solnMessage).value

				if value.Less(topValue) {
					c.log("received solnMsgType: value was better")
					topValue = value
					top.Free()
					top = c.inst.SolutionFromRats(best)

					for child := range c.toChildren {
						if child != msg.Source() {
							c.sendToChild(child, valueMsgType, topValue)
						}
					}
				} else {
					c.log("received solnMsgType: value was worse")
				}
			}
		case <-timeout:
			c.log("timeout: sending stopMsgType to all islands")
			c.stop()
			c.log("received finMsgType from all islands: exiting")
			break msgLoop
		}
	}

	return top
}
