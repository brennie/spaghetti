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

	"github.com/brennie/spaghetti/options"
	"github.com/brennie/spaghetti/solver/hpga/population"
	"github.com/brennie/spaghetti/tt"
)

// A slave is a child of an island.
type slave struct {
	child
	island   int                    // The island the slave belongs to
	inst     *tt.Instance           // The timetabling instance.
	topValue tt.Value               // The best seen value thus far.
	pop      *population.Population // The slave's population of soluations
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
		tt.WorstValue(),
		nil,
	}

	go s.run(opts.MinPop, opts.MaxPop)

	return fromParent
}

func (s *slave) handleMessage(msg messageContent) (shouldExit bool) {
	shouldExit = false

	switch msg.messageType() {
	case valueMessageType:
		if value := msg.(valueMessage).value; value.Less(s.topValue) {
			s.topValue = value
		}

	case gmRequestMessageType:
		s.sendToParent(gmReplyMessage{s.pop.PickSolution()})

	case individualRequestMessageType:
		id := msg.(individualRequestMessage).id
		s.sendToParent(individualReplyMessage{id, s.pop.PickIndividual()})

	case stopMessageType:
		shouldExit = true
		s.fin()
	}

	return
}

// Run the slave.
func (s *slave) run(minPop, maxPop int) {
	topValue := tt.WorstValue()

	// Generate the population and signal the island that population generation has finished.
	s.pop = population.New(s.inst, minPop, maxPop)
	(<-s.fromParent).content.(waitMessage).wg.Done()

	if best, value := s.pop.Best(); value.Less(topValue) {
		topValue = value
		s.sendToParent(solutionMessage{best.Assignments(), topValue})
	}

	for {
		received := true
		for received {
			select {
			case msg := <-s.fromParent:
				if shouldExit := s.handleMessage(msg.content); shouldExit {
					return
				}

			default:
				received = false
			}
		}

		prob := rand.Intn(99) + 1 // [1, 100]

		if prob < pLocal+pMutate {
			var individual *tt.Solution
			var value tt.Value

			if prob < pMutate {
				individual = s.pop.RemoveOne()
				mutate(individual, rand.Intn(s.inst.NEvents()))
				value = individual.Value()
			} else {
				mother := s.pop.PickIndividual()
				father := s.pop.PickIndividual()

				individual = s.inst.NewSolution()

				value = population.Crossover(mother, father, individual)
			}

			s.pop.Insert(individual)

			if value.Less(topValue) {
				topValue = value
				s.sendToParent(solutionMessage{individual.Assignments(), topValue})
			}

		} else {
			s.sendToParent(crossoverRequestMessage{s.pop.PickIndividual()})

			// We wait for a solutionMessageType message and process messages in the
			// mean time. We do this as try to not overload the island with
			// messages.
			msg := <-s.fromParent
			for msg.messageType() != solutionMessageType {
				if s.handleMessage(msg.content) {
					return
				}
				msg = <-s.fromParent
			}
			soln := msg.content.(solutionMessage).soln
			value := msg.content.(solutionMessage).value

			s.pop.Insert(s.inst.SolutionFromRats(soln))

			if value.Less(topValue) {
				topValue = value
			}

		}
	}
}
