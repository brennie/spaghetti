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

	"github.com/brennie/spaghetti/options"
	"github.com/brennie/spaghetti/solver/hpga/population"
	"github.com/brennie/spaghetti/tt"
)

// A slave is a child of an island.
type slave struct {
	child
	island   int                    // The island the slave belongs to
	inst     *tt.Instance           // The timetabling instance.
	verbose  bool                   // Determines if events should be logged.
	topValue tt.Value               // The best seen value thus far.
	pop      *population.Population // The slave's population of soluations
}

// Optionally log a message if the verbose flag is set.
func (s *slave) log(format string, args ...interface{}) {
	if s.verbose {
		msg := fmt.Sprintf(format, args...)
		log.Printf("slave(%d.%d): %s\n", s.island, s.id, msg)
	}
}

// Create a new slave with the given id. The given channel is the channel the
// island should use to communicate with the controller. The channel returned
// is the channel the controller should use to communicate with the island.
func newSlave(island int, id int, inst *tt.Instance, toParent chan<- message, opts options.SolveOptions) chan<- message {
	fromParent := make(chan message, 5)
	s := &slave{
		child{
			id,
			fromParent,
			toParent,
		},
		island,
		inst,
		opts.Verbose,
		tt.Value{0, 0},
		nil,
	}

	go s.run(opts.MinPop, opts.MaxPop)

	return fromParent
}

func (s *slave) handleMessage(msg message) (shouldExit bool) {
	shouldExit = false

	switch msg.messageType() {
	case valueMessageType:
		if value := msg.content.(valueMessage).value; value.Less(s.topValue) {
			s.topValue = value
			s.log("received valueMessageType: value was better")
		} else {
			s.log("received valueMessageType: value was worse")
		}

	case gmRequestMessageType:
		s.sendToParent(gmReplyMessage{s.pop.PickSolution()})
		s.log("received gmRequestMessageType; replied with solution")

	case individualRequestMessageType:
		id := msg.content.(individualRequestMessage).id
		s.sendToParent(individualReplyMessage{id, s.pop.PickIndividual()})
		s.log("received individualRequestMessageType; replied with solution")

	case stopMessageType:
		s.log("received stopMessageType; exiting")
		shouldExit = true
		s.fin()
	}

	return
}

// Run the slave.
func (s *slave) run(minPop, maxPop int) {
	// Receive the topValue-valued solution message from the island.
	topValue := (<-s.fromParent).content.(valueMessage).value

	s.log("generating population...")

	s.pop = population.New(s.inst, minPop, maxPop)

	s.log("finished generating population")

	if best, value := s.pop.Best(); value.Less(topValue) {
		topValue = value
		s.sendToParent(solutionMessage{best.Assignments(), value})

		s.log("found new best-valued solution: (%s)", value)
	}

	for {
		select {
		case msg := <-s.fromParent:
			if s.handleMessage(msg) {
				return
			}

		default:
		}

		prob := rand.Intn(99) + 1 // [1, 100]

		if prob < pLocal+pMutate {
			var individual *tt.Solution
			var value tt.Value

			if prob < pMutate {
				individual = s.pop.RemoveOne()
				nAssigned := individual.NAssigned()

				// If all the events have been assigned, we mutate one of them
				// at random. Otherwise we pick the Nth unassigned event and
				// walk through the events until we find its index.
				if nAssigned == s.inst.NEvents() {
					mutate(individual, rand.Intn(s.inst.NEvents()))
				} else {
					chromosome := rand.Intn(s.inst.NEvents() - nAssigned)
					for i := 0; i < s.inst.NEvents(); i++ {
						if individual.Assigned(i) {
							chromosome++
						} else if i == chromosome {
							break
						}
					}
					mutate(individual, chromosome)
					value = individual.Value()
				}

				s.log("performed a mutation")
			} else {
				mother := s.pop.PickIndividual()
				father := s.pop.PickIndividual()

				individual = s.inst.NewSolution()

				value = population.Crossover(mother, father, individual)

			}

			s.pop.Insert(individual)

			if value.Less(topValue) {
				topValue = value

				s.log("found new best-valued solution: (%d,%d)", value.Distance, value.Fitness)

				s.sendToParent(solutionMessage{individual.Assignments(), value})
			}

		} else {
			s.sendToParent(crossoverRequestMessage{s.pop.PickIndividual()})
			s.log("sent crossover request to island(%d); awaiting reply", s.island)

			// We wait for a solutionMessageType message and process messages in the
			// mean time. We do this as try to not overload the island with
			// messages.
			msg := <-s.fromParent
			for msg.messageType() != solutionMessageType {
				if s.handleMessage(msg) {
					return
				}
				msg = <-s.fromParent
			}

			s.log("received solutionMessageType from island(%d); inserting into population", s.island)
			soln := msg.content.(solutionMessage).soln
			value := msg.content.(solutionMessage).value

			s.pop.Insert(s.inst.SolutionFromRats(soln))

			if value.Less(topValue) {
				topValue = value
				s.log("result of crossover is a best-valued solution: (%s)", value)
			}

		}
	}
}
