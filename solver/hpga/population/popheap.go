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

package population

// The actual population heap. This is unexported so that the Push and Pop
// methods cannot be used except by this package.
type popHeap []*individual

// Push an element onto the heap. Use Insert instead.
func (heap *popHeap) Push(element interface{}) {
	if len(*heap) == cap(*heap) {
		panic("popHeap.Push() on a full heap")
	}

	// We don't have to check this type assertion because Push can only be
	// called through Population.Insert (as popHeap.Push isn't exported) which
	// is guaranteed to call this function with an individual.
	*heap = append(*heap, element.(*individual))
}

// Remove an element from the population.
func (heap *popHeap) Pop() (element interface{}) {
	if len(*heap) == 0 {
		panic("Popping from empty Population")
	}
	newLen := len(*heap) - 1
	element = (*heap)[newLen].soln
	*heap = (*heap)[0:newLen]

	return
}

// Get the size of the population.
func (heap popHeap) Len() int {
	return len(heap)
}

// Determine if one solution has a lesser valuation than another.
func (heap popHeap) Less(i, j int) bool {
	return heap[i].value.Less(heap[j].value)
}

// Swap to members of the population.
func (heap popHeap) Swap(i, j int) {
	heap[i], heap[j] = heap[j], heap[i]
}
