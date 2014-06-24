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
	rng     *rand.Rand   // The random number generator for the i.
	inst    *tt.Instance // The timetabling instance.
	verbose bool         // Determines if events should be logged.
}

type crossoverRequest struct {
	origin int          // The slave that requested the crossover
	mother *tt.Solution // The given solutions
}

// Create a new island with the given id and number of slaves. The given
// channel is the channel the island should use to communicate with the
// controller. The channel returned is the channel the controller should use to
// communicate with the i.
func newIsland(id, nSlaves int, inst *tt.Instance, seed int64, toParent chan<- message, verbose bool) chan<- message {
	fromParent := make(chan message, 5)
	comm := make(chan message, 5)

	i := &island{
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
		inst,
		verbose,
	}

	for child := 0; child < nSlaves; child++ {
		i.toChildren[child] = newSlave(id, child, inst, rand.Int63(), comm, verbose)
	}

	go i.run()

	return fromParent
}

// Optionally log a message if the verbose flag is set.
func (i *island) log(format string, args ...interface{}) {
	if i.verbose {
		msg := fmt.Sprintf(format, args...)
		log.Printf("island(%d): %s\n", i.id, msg)
	}
}

// Run the island.
func (i *island) run() {
	topValue := (<-i.fromParent).(valueMessage).value

	crossovers := make(map[int]crossoverRequest)

	for child := range i.toChildren {
		i.sendToChild(child, valueMsgType, topValue)
	}

	for {
		select {
		case msg := <-i.fromParent:

			switch msg.MsgType() {
			case stopMsgType:
				i.log("received stopMsgType; sending stop to slaves")
				i.stop()
				i.log("received finMsgType from all slaves; exiting")
				i.fin()
				return

			case valueMsgType:
				if value := msg.(valueMessage).value; value.Less(topValue) {
					i.log("received valueMsgType: value was better")

					topValue = value

					for child := range i.toChildren {
						i.sendToChild(child, valueMsgType, topValue)
					}
				} else {
					i.log("received valueMsgType: value was worse")
				}
			}

		case msg := <-i.fromChildren:
			switch msg.MsgType() {
			case solnMsgType:
				best, value := msg.(solnMessage).soln, msg.(solnMessage).value

				if value.Less(topValue) {
					i.log("received solnMsgType: value was better")

					i.sendToParent(solnMsgType, best, value)
					topValue = value

					for child := range i.toChildren {
						if child != msg.Source() {
							i.sendToChild(child, valueMsgType, topValue)
						}
					}
				} else {
					i.log("received solnMsgType: value was worse")
				}

			case xoverReqMsgType:
				request := msg.(xoverReqMessage)

				// Generate a currently unused crossover id
				id := i.rng.Int()
				for _, used := crossovers[id]; used; {
					id = i.rng.Int()
				}

				i.log("received a crossover request from slave(%d,%d); assigned id %d", i.id, request.Source(), id)

				crossovers[id] = crossoverRequest{request.Source(), &request.soln}

				// We generate a random number in [0, N-1) as there are N-1 other
				// slaves under the i. We can then map all n >= nSource to
				// n+1 to get a uniform probability that any slave that is not the
				// source is picked.
				other := i.rng.Intn(len(i.toChildren) - 1)
				if other >= msg.Source() {
					other++
				}

				i.sendToChild(other, solnReqMsgType, id)
				i.log("sent solution request %d to slave(%d,%d)", id, i.id, other)

			case solnReplyMsgType:
				id := msg.(solnReplyMessage).id

				if _, used := crossovers[id]; used {
					reply := msg.(solnReplyMessage)
					child := i.inst.NewSolution()
					chromosome := i.rng.Intn(i.inst.NEvents())

					i.log("received a solnReplyMsgType with id %d from slave(%d,%d); doing crossover", id, i.id, msg.Source())

					crossover(crossovers[id].mother, &reply.soln, child, chromosome)
					value := child.Value()
					i.sendToChild(crossovers[id].origin, solnMsgType, child.Value(), *child)
					i.log("sent crossover result of crossover %d to slave(%d,%d)", id, i.id, crossovers[id].origin)

					if value.Less(topValue) {
						topValue = value
						for child := range i.toChildren {
							if child != crossovers[id].origin {
								i.sendToChild(child, valueMsgType, value)
							}
						}
					}

					delete(crossovers, id)
				} else {
					i.log("received a crossover id (%d) that is not currently in use from slave(%d,%d)", id, i.id, msg.Source())
				}
			}
		}
	}
}
