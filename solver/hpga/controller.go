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

	"github.com/brennie/spaghetti/solver/heuristics"
	"github.com/brennie/spaghetti/tt"
)

// A controller is just a parent; its children are the islands.
type controller struct {
	parent
	inst *tt.Instance
}

// Create a new controller. There will be nIslands islands, each with nSlaves
// slaves.
func newController(nIslands, nSlaves int, inst *tt.Instance) *controller {
	comm := make(chan message)

	controller := &controller{
		parent{
			comm,
			make([]chan<- message, nIslands),
		},
		inst,
	}

	for i := 0; i < nIslands; i++ {
		controller.parent.toChildren[i] = newIsland(i, nSlaves, inst, rand.Int63(), comm)
	}

	return controller
}

// Run the controller.
func (controller *controller) run() *tt.Solution {
	top := controller.inst.NewSolution() // The top-valued solution over the whole HPGA.

	heuristics.MostConstrainedOrdering(top)
	topValue := top.Value() // The value of the top-valued solution over the whole HPGA.

	for child := range controller.toChildren {
		controller.send(child, valueMsg, topValue)
	}

	controller.stop()

	return top
}
