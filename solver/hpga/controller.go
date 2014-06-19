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
	comm := make(chan message)

	controller := &controller{
		parent{
			comm,
			make([]chan<- message, nIslands),
		},
		inst,
		verbose,
	}

	for i := 0; i < nIslands; i++ {
		controller.parent.toChildren[i] = newIsland(i, nSlaves, inst, rand.Int63(), comm, verbose)
	}

	return controller
}

func (controller *controller) log(format string, args ...interface{}) {
	if controller.verbose {
		msg := fmt.Sprintf(format, args...)
		log.Printf("controller: %s\n", msg)
	}
}

// Run the controller.
func (controller *controller) run() *tt.Solution {
	top := controller.inst.NewSolution() // The top-valued solution over the whole HPGA.

	heuristics.MostConstrainedOrdering(top)
	topValue := top.Value() // The value of the top-valued solution over the whole HPGA.

	for child := range controller.toChildren {
		controller.sendToChild(child, valueMsg, topValue)
	}

msgLoop:
	for {
		timeout := time.After(2 * time.Minute)
		select {
		case msg := <-controller.fromChildren:
			switch msg.MsgType() {
			case solnMsg:
				best, value := msg.(solnMessage).soln, msg.(solnMessage).value

				if value.Less(topValue) {
					controller.log("received solnMsg: value was better")
					topValue = value
					top = &best

					for child := range controller.toChildren {
						if child != msg.Source() {
							controller.sendToChild(child, valueMsg, topValue)
						}
					}
				} else {
					controller.log("received solnMsg: value was worse")
				}
			}
		case <-timeout:
			controller.log("timeout: sending stopMsg to all islands")
			controller.stop()
			controller.log("received finMsg from all islands: exiting")
			break msgLoop
		}
	}

	return top
}
