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

	"github.com/brennie/spaghetti/tt"
)

// An island is both a parent (slaves run under it) and a child (it runs under
// the controller).
type island struct {
	parent
	child
	rng *rand.Rand // The random number generator for the island.
}

// Create a new island with the given id and number of slaves. The given
// channel is the channel the island should use to communicate with the
// controller. The channel returned is the channel the controller should use to
// communicate with the island.
func newIsland(id, nSlaves int, inst *tt.Instance, seed int64, toParent chan<- message) chan<- message {
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
	}

	for i := 0; i < nSlaves; i++ {
		island.toChildren[i] = newSlave(i, inst, rand.Int63(), comm)
	}

	go island.run()

	return fromParent
}

// Run the island.
func (island *island) run() {
	top := (<-island.fromParent).(valueMessage).value
	for child := range island.toChildren {
		island.sendToChild(child, valueMsg, top)
	}

	for {
		select {
		case msg := <-island.fromParent:

			switch msg.MsgType() {
			case stopMsg:
				island.stop()
				island.fin()
				return
			}

		case <-island.fromChildren:
			break
		}
	}
}
