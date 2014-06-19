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

	"github.com/brennie/spaghetti/tt"
)

// An island is both a parent (slaves run under it) and a child (it runs under
// the controller).
type island struct {
	parent
	child
	rng     *rand.Rand // The random number generator for the island.
	verbose bool       // Determines if events should be logged.
}

// Create a new island with the given id and number of slaves. The given
// channel is the channel the island should use to communicate with the
// controller. The channel returned is the channel the controller should use to
// communicate with the island.
func newIsland(id, nSlaves int, inst *tt.Instance, seed int64, toParent chan<- message, verbose bool) chan<- message {
	fromParent := make(chan message)
	comm := make(chan message)

	island := &island{
		parent{
			comm,
			make([]chan<- message, nSlaves),
		},
		child{
			id,
			fromParent,
			toParent,
		},
		rand.New(rand.NewSource(seed)),
		verbose,
	}

	for i := 0; i < nSlaves; i++ {
		island.toChildren[i] = newSlave(id, i, inst, rand.Int63(), comm, verbose)
	}

	go island.run()

	return fromParent
}

func (island *island) log(format string, args ...interface{}) {
	if island.verbose {
		msg := fmt.Sprintf(format, args...)
		log.Printf("island(%d): %s\n", island.id, msg)
	}
}

// Run the island.
func (island *island) run() {
	topValue := (<-island.fromParent).(valueMessage).value

	for child := range island.toChildren {
		island.sendToChild(child, valueMsg, topValue)
	}

	for {
		select {
		case msg := <-island.fromParent:

			switch msg.MsgType() {
			case stopMsg:
				island.log("received stopMsg; sending stop to slaves")
				island.stop()
				island.log("received finMsg from all slaves; exiting")
				island.fin()
				return

			case valueMsg:
				if value := msg.(valueMessage).value; value.Less(topValue) {
					island.log("received valueMsg: value was better")

					topValue = value

					for child := range island.toChildren {
						island.sendToChild(child, valueMsg, topValue)
					}
				} else {
					island.log("received valueMsg: value was worse")
				}
			}

		case msg := <-island.fromChildren:
			switch msg.MsgType() {
			case solnMsg:
				best, value := msg.(solnMessage).soln, msg.(solnMessage).value

				if value.Less(topValue) {
					island.log("received solnMsg: value was better")

					island.sendToParent(solnMsg, best, value)
					topValue = value

					for child := range island.toChildren {
						if child != msg.Source() {
							island.sendToChild(child, valueMsg, topValue)
						}
					}
				} else {
					island.log("received solnMsg: value was worse")
				}
			}
		}
	}
}
