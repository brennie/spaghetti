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

	"github.com/brennie/spaghetti/solver/hpga/population"
	"github.com/brennie/spaghetti/tt"
)

const (
	pMutate = 5      // The probability of a mutation is 5%
	pLocal  = 5 + 75 // The probability of a local crossover is 75%
)

// A slave is a child of an island.
type slave struct {
	child
	island   int                    // The island the slave belongs to
	inst     *tt.Instance           // The timetabling instance.
	rng      *rand.Rand             // The random number generator for the s.
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
func newSlave(island int, id int, inst *tt.Instance, seed int64, toParent chan<- message, verbose bool) chan<- message {
	fromParent := make(chan message, 5)
	s := &slave{
		child{
			id,
			fromParent,
			toParent,
		},
		island,
		inst,
		rand.New(rand.NewSource(seed)),
		verbose,
		tt.Value{0, 0},
		nil,
	}

	go s.run()

	return fromParent
}

func (s *slave) handleMessage(msg message) (shouldExit bool) {
	shouldExit = false

	switch msg.MsgType() {
	case valueMsgType:
		if value := msg.(valueMessage).value; value.Less(s.topValue) {
			s.topValue = value
			s.log("received valueMsgType: value was better")
		} else {
			s.log("received valueMsgType: value was worse")
		}

	case solnReqMsgType:
		id := msg.(solnReqMessage).id
		s.sendToParent(solnReplyMsgType, id, s.pop.Pick(s.rng).Clone())

	case stopMsgType:
		s.fin()
		s.log("received stopMsgType; exiting")
		shouldExit = true
	}

	return
}

// Run the slave.
func (s *slave) run() {
	// Receive the topValue-valued solution message from the island.
	topValue := (<-s.fromParent).(valueMessage).value

	s.log("generating population...")

	s.pop = population.New(s.rng, s.inst)

	s.log("finished generating population")

	if best, value := s.pop.Best(); value.Less(topValue) {
		topValue = value
		s.sendToParent(solnMsgType, value, best.Clone())

		s.log("found new best-valued solution: (%d,%d)", value.Distance, value.Fitness)
	}

	for {
		select {
		case msg := <-s.fromParent:
			if s.handleMessage(msg) {
				return
			}

		default:
		}

		prob := s.rng.Intn(99) + 1 // [1, 100]

		if prob < pLocal {
			var individual *tt.Solution

			if prob < pMutate {
				individual = s.pop.RemoveOne(s.rng)
				chromosome := s.rng.Intn(s.inst.NEvents())

				mutate(individual, chromosome)

				s.log("performed a mutation")
			} else {
				mother := s.pop.Pick(s.rng)
				father := s.pop.Pick(s.rng)

				individual = s.inst.NewSolution()

				chromosome := s.rng.Intn(s.inst.NEvents())

				crossover(mother, father, individual, chromosome)

				s.log("performed a local crossover")
			}

			s.pop.Insert(individual)
			value := individual.Value()

			if value.Less(topValue) {
				topValue = value

				s.log("found new best-valued solution: (%d,%d)", value.Distance, value.Fitness)

				s.sendToParent(solnMsgType, individual.Clone(), value)
			}

		} else {
			individual := s.pop.Pick(s.rng).Clone()

			s.sendToParent(xoverReqMsgType, individual)
			s.log("sent crossover request to island(%d); awaiting reply", s.island)

			// We wait for a solnMsgType message and process messages in the
			// mean time. We do this as try to not overload the island with
			// messages.
			msg := <-s.fromParent
			for msg.MsgType() != solnMsgType {
				if s.handleMessage(msg) {
					return
				}
				msg = <-s.fromParent
			}

			s.log("received solnMsgType from island(%d); inserting into population", s.island)
			soln := msg.(solnMessage).soln
			value := msg.(solnMessage).value

			s.pop.Insert(soln)

			if value.Less(topValue) {
				topValue = value
				s.log("result of crossover is a best-valued solution: (%d, %d)", value.Distance, value.Fitness)
			}

		}
	}
}
