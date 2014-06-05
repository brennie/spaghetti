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

// A controller is just a parent; its children are the islands.
type controller struct {
	parent
}

// Create a new controller. There will be nIslands islands, each with nSlaves
// slaves.
func newController(nIslands, nSlaves int) *controller {
	comm := make(chan message)

	controller := &controller{
		parent{
			comm,
			make([]chan<- message, nIslands),
		},
	}

	for i := 0; i < nIslands; i++ {
		controller.toChildren[i] = newIsland(i, nSlaves, comm)
	}

	return controller
}

// Run the controller; currently this just stops the HPGA.
func (c *controller) run() {
	c.stop()
}
