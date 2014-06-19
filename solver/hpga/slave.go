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

	"github.com/brennie/spaghetti/solver/hpga/population"
	"github.com/brennie/spaghetti/tt"
)

// A slave is just a child of an island.
type slave struct {
	child
	inst *tt.Instance
	rng  *rand.Rand
}

// Create a new slave with the given id. The given channel is the channel the
// island should use to communicate with the controller. The channel returned
// is the channel the controller should use to communicate with the island.
func newSlave(id int, inst *tt.Instance, seed int64, toParent chan<- message) chan<- message {
	fromParent := make(chan message)
	slave := &slave{
		child{
			id,
			fromParent,
			toParent,
		},
		inst,
		rand.New(rand.NewSource(seed)),
	}

	go slave.run()

	return fromParent
}

// Run the slave.
func (slave *slave) run() {
	// Receive the topValue-valued solution message from the island.
	topValue := (<-slave.fromParent).(valueMessage).value

	pop := population.New(slave.rng, slave.inst)
	if best, value := pop.Best(); value.Less(topValue) {
		topValue = value
		slave.sendToParent(solnMsg, *best.Clone(), value)
	}

	for {
		select {
		case msg := <-slave.fromParent:
			switch msg.MsgType() {

			// Update the globally known topValue value.
			case valueMsg:
				if value := msg.(valueMessage).value; value.Less(topValue) {
					topValue = value
				}

			// StopValue the slave.
			case stopMsg:
				slave.fin()
				return
			}
		}
	}

	// XXX: This is here to temporarily squelch a compiler warning that topValue is
	// declared and not used.
	if &topValue == nil || &pop == nil {
	}
}
