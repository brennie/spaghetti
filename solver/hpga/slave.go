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

import "math/rand"

// A slave is just a child of an island.
type slave struct {
	child
}

// Create a new slave with the given id. The given channel is the channel the
// island should use to communicate with the controller. The channel returned
// is the channel the controller should use to communicate with the island.
func newSlave(id int, toParent chan<- message) chan<- message {
	fromParent := make(chan message)
	slave := &slave{
		child{
			id,
			fromParent,
			toParent,
		},
	}

	go slave.run()

	return fromParent
}

// Run the slave.
func (slave *slave) run() {
	var rng *rand.Rand

	for {
		msg := <-slave.fromParent

		switch msg.MsgType() {
		case stopMsg:
			slave.fin()
			return

		case seedMsg:
			rng = rand.New(rand.NewSource(msg.(seedMessage).seed))

		default:
			break
		}
	}

	// XXX: This is here to temporarily squelch a compiler warning that rng is
	// declared and not used.
	rng.Int31()
}
