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

// An island is both a parent (slaves run under it) and a child (it runs under
// the controller).
type island struct {
	parent
	child
}

// Create a new island with the given id and number of slaves. The given
// channel is the channel the island should use to communicate with the
// controller. The channel returned is the channel the controller should use to
// communicate with the island.
func newIsland(id, nSlaves int, toParent chan<- message) chan<- message {
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
	}

	for i := 0; i < nSlaves; i++ {
		island.toChildren[i] = newSlave(i, comm)
	}

	go island.run()

	return fromParent
}

// Run the island.
func (island *island) run() {
	for {
		msg := <-island.fromParent

		switch msg.MsgType() {
		case stopMsg:
			island.stop()
			island.toParent <- baseMessage{island.id, finMsg}
			return

		case seedMsg:
			island.seedChildren(msg.(seedMessage).seed)

		default:
			break
		}
	}

}

// Seed the children by creating a new RNG with the given seed.
func (island *island) seedChildren(baseSeed int64) {
	rng := rand.New(rand.NewSource(baseSeed))
	for child := range island.toChildren {
		island.send(child, seedMsg, rng.Int63())
	}
}
