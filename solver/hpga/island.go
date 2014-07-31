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
	"math/rand"
	"time"

	"github.com/brennie/spaghetti/options"
	"github.com/brennie/spaghetti/solver/hpga/population"
	"github.com/brennie/spaghetti/tt"
)

const gmInterval = 5 * time.Minute

// An island is both a parent (slaves run under it) and a child (it runs under
// the controller).
type island struct {
	parent
	child
	inst      *tt.Instance           // The timetabling instance.
	pop       *population.Population // The island's population
	topValue  tt.Value               // The best value seen thus far.
	ordering  []int                  // The variable ordering for the GM operator.
	gmRecv    <-chan bool            // Determine when the GM operator has finished.
	gmSend    chan<- bool            // Channel to send GM requests.
	generated []tt.Pair              // The list of individuals generated by the GM operatoer
}

// Create a new island with the given id and number of slaves. The given
// channel is the channel the island should use to communicate with the
// controller. The channel returned is the channel the controller should use to
// communicate with the i.
func newIsland(id int, inst *tt.Instance, toParent chan<- message, opts options.SolveOptions) chan<- message {
	fromParent := make(chan message, 5)
	fromChildren := make(chan message, 5)
	gmRecv := make(chan bool)
	gmSend := make(chan bool)

	toGenerate := int(0.05*float64(opts.MinPop)) * opts.NSlaves

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
		population.New(opts.MinPop, opts.MaxPop, opts.NSlaves),
		tt.WorstValue(),
		nil,
		gmRecv,
		gmSend,
		make([]tt.Pair, toGenerate),
	}

	for child := 0; child < opts.NSlaves; child++ {
		i.toChildren[child] = newSlave(id, child, inst, i.pop.SubPopulation(child), fromChildren, opts)
	}

	go i.run()
	go i.runGM(toGenerate, gmSend, gmRecv)

	return fromParent
}

// Perform selection and notify the children that they can continue.
func (i *island) doSelection() {
	if i.ordering == nil {
		i.pop.Select(nil)
	} else {
		// Wait for the GM to generate some individuals.
		<-i.gmRecv
		i.pop.Select(i.generated)
		for individual := range i.generated {
			i.generated[individual].Soln = nil
		}
		i.gmSend <- true
	}

	for child := range i.toChildren {
		i.sendToChild(child, continueMessage{})
	}
}

// Run the island.
func (i *island) run() {
	crossovers := make(map[int]int) // Map crossover id -> child id

	i.wait()
	(<-i.fromParent).content.(waitMessage).wg.Done()

	for {
		select {
		case msg := <-i.fromParent:
			switch msg.messageType() {
			case stopMessageType:
				i.stopChildren()
				i.fin()
				return

			case valueMessageType:
				if value := msg.content.(valueMessage).value; value.Less(i.topValue) {

					i.topValue = value

					for child := range i.toChildren {
						i.sendToChild(child, valueMessage{i.topValue})
					}
				}

			case orderingMessageType:
				i.ordering = msg.content.(orderingMessage).order
				i.gmSend <- true
			}

		case msg := <-i.fromChildren:
			switch msg.messageType() {
			case solutionMessageType:
				best, value := msg.content.(solutionMessage).soln, msg.content.(solutionMessage).value

				if value.Less(i.topValue) {
					i.sendToParent(solutionMessage{best, value})
					i.topValue = value

					for child := range i.toChildren {
						if child != msg.source {
							i.sendToChild(child, valueMessage{i.topValue})
						}
					}
				}

			case crossoverMessageType:
				request := msg.content.(crossoverMessage)

				if request.id == newRequest {

					// Generate a currently unused crossover id
					id := rand.Int()
					for _, used := crossovers[id]; used; {
						id = rand.Int()
					}

					crossovers[id] = msg.source

					// We generate a random number in [0, N-1) as there are N-1 other
					// slaves under the i. We can then map all n >= nSource to
					// n+1 to get a uniform probability that any slave that is not the
					// source is picked.
					other := rand.Intn(len(i.toChildren) - 1)
					if other >= msg.source {
						other++
					}

					i.sendToChild(other, crossoverMessage{id})
				} else {
					if _, used := crossovers[request.id]; used {
						origin := crossovers[request.id]
						child, value := i.pop.Crossover(origin, msg.source, i.inst)

						if value.Less(i.topValue) {
							i.topValue = value

							for child := range i.toChildren {
								i.sendToChild(child, valueMessage{value})
							}
							i.sendToParent(solutionMessage{child.Assignments(), value})
						}

						if i.pop.IsSubPopulationFull(origin) {
							if i.pop.ShouldSelect() {
								i.doSelection()
							}
						} else {
							i.sendToChild(origin, continueMessage{})
						}

						i.sendToChild(msg.source, continueMessage{})

						delete(crossovers, request.id)
					}
				}

			case fullMessageType:
				if i.pop.ShouldSelect() {
					i.doSelection()
				}
			}
		}
	}
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
