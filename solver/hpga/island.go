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

	"github.com/brennie/spaghetti/options"
	"github.com/brennie/spaghetti/tt"
)

const gmInterval = 5 * time.Minute

// An island is both a parent (slaves run under it) and a child (it runs under
// the controller).
type island struct {
	parent
	child
	inst     *tt.Instance // The timetabling instance.
	verbose  bool         // Determines if events should be logged.
	topValue tt.Value     // The best value seen thus far.
	gmTimer  *time.Timer  // A timer for the GM operation.
}

type crossoverRequest struct {
	origin int      // The slave that requested the crossover.
	mother []tt.Rat // The first parent to crossover with.
}

// Create a new island with the given id and number of slaves. The given
// channel is the channel the island should use to communicate with the
// controller. The channel returned is the channel the controller should use to
// communicate with the i.
func newIsland(id int, inst *tt.Instance, toParent chan<- message, opts options.SolveOptions) chan<- message {
	fromParent := make(chan message, 5)
	fromChildren := make(chan message, 5)

	i := &island{
		parent{
			fromChildren,
			make([]chan<- message, opts.NSlaves),
		},
		child{
			id,
			fromParent,
			toParent,
		},
		inst,
		opts.Verbose,
		tt.Value{-1, -1},
		nil,
	}

	for child := 0; child < opts.NSlaves; child++ {
		i.toChildren[child] = newSlave(id, child, inst, fromChildren, opts)
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
	i.topValue = (<-i.fromParent).content.(valueMessage).value

	crossovers := make(map[int]crossoverRequest)

	for child := range i.toChildren {
		i.sendToChild(child, valueMessage{i.topValue})
	}

	i.gmTimer = time.NewTimer(gmInterval)

	for {
		select {
		case msg := <-i.fromParent:

			switch msg.messageType() {
			case stopMessageType:
				i.log("received stopMessageType; sending stop to slaves")
				i.stopChildren()
				i.log("received finMessageType from all slaves; exiting")
				i.fin()
				return

			case valueMessageType:
				if value := msg.content.(valueMessage).value; value.Less(i.topValue) {
					i.log("received valueMessageType: value was better")

					i.topValue = value

					for child := range i.toChildren {
						i.sendToChild(child, valueMessage{i.topValue})
					}
				} else {
					i.log("received valueMessageType: value was worse")
				}
			}

		case msg := <-i.fromChildren:
			switch msg.messageType() {
			case gmReplyMessageType:
				individual := msg.content.(gmReplyMessage).soln
				i.log("received gmReplyMessageType from slave(%d.%d)", i.id, msg.source)
				go i.runGM(individual, msg.source)

			case solutionMessageType:
				best, value := msg.content.(solutionMessage).soln, msg.content.(solutionMessage).value

				if value.Less(i.topValue) {
					i.log("received solutionMessageType: value was better")

					i.sendToParent(solutionMessage{best, value})
					i.topValue = value

					for child := range i.toChildren {
						if child != msg.source {
							i.sendToChild(child, valueMessage{i.topValue})
						}
					}
				} else {
					i.log("received solutionMessageType: value was worse")
				}

			case crossoverRequestMessageType:
				request := msg.content.(crossoverRequestMessage)

				// Generate a currently unused crossover id
				id := rand.Int()
				for _, used := crossovers[id]; used; {
					id = rand.Int()
				}

				i.log("received a crossover request from slave(%d.%d); assigned id %d", i.id, msg.source, id)

				crossovers[id] = crossoverRequest{msg.source, request.soln}

				// We generate a random number in [0, N-1) as there are N-1 other
				// slaves under the i. We can then map all n >= nSource to
				// n+1 to get a uniform probability that any slave that is not the
				// source is picked.
				other := rand.Intn(len(i.toChildren) - 1)
				if other >= msg.source {
					other++
				}

				i.sendToChild(other, solutionRequestMessage{id})
				i.log("sent solution request %d to slave(%d.%d)", id, i.id, other)

			case solutionReplyMessageType:
				id := msg.content.(solutionReplyMessage).id

				if _, used := crossovers[id]; used {
					reply := msg.content.(solutionReplyMessage)
					child := i.inst.NewSolution()
					chromosome := rand.Intn(i.inst.NEvents())

					i.log("received a solutionReplyMessageType with id %d from slave(%d.%d); doing crossover", id, i.id, msg.source)

					crossover(crossovers[id].mother, reply.soln, child, chromosome)
					value := child.Value()
					i.sendToChild(crossovers[id].origin, solutionMessage{child.Assignments(), child.Value()})
					i.log("sent crossover result of crossover %d to slave(%d.%d)", id, i.id, crossovers[id].origin)

					if value.Less(i.topValue) {
						i.topValue = value
						for child := range i.toChildren {
							if child != crossovers[id].origin {
								i.sendToChild(child, valueMessage{value})
							}
						}
					}

					delete(crossovers, id)
				} else {
					i.log("received a crossover id (%d) that is not currently in use from slave(%d.%d)", id, i.id, msg.source)
				}
			}

		case <-i.gmTimer.C:
			child := rand.Intn(len(i.toChildren))

			i.sendToChild(child, gmRequestMessage{})
			i.log("requested solution for GM from slave(%d.%d)", i.id, child)
		}
	}
}

// Run the genetic modification operation on the given individual.
func (i *island) runGM(individual []tt.Rat, child int) {
	soln := i.inst.SolutionFromRats(individual)

	chromosome := rand.Intn(i.inst.NEvents())

	gm(soln, chromosome)

	i.sendToChild(child, solutionMessage{soln.Assignments(), soln.Value()})
	soln.Free()
	i.log("sent result of GM to slave(%d.%d)", i.id, child)

	i.gmTimer.Reset(gmInterval)
}

// Send a stopMessageType message to all slaves under the island and wait for a
// finMessageType message from each of them. If a solutionMessageType message arrives, it
// will be processed as normal (i.e., forwarded to the controller if the
// associated value is better than i.top).
func (i *island) stopChildren() {
	for child := range i.toChildren {
		i.sendToChild(child, stopMessage{})
	}

	finished := make(map[int]bool)

	for len(finished) != len(i.toChildren) {
		msg := <-i.fromChildren

		switch msg.messageType() {
		case finMessageType:
			finished[msg.source] = true

		case solutionMessageType:
			value := msg.content.(solutionMessage).value

			if value.Less(i.topValue) {
				soln := msg.content.(solutionMessage).soln
				i.topValue = value
				i.sendToParent(solutionMessage{soln, value})
			}
		}
	}
}
