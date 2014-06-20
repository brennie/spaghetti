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

// The package exporting the Population type
package population

import (
	"container/heap"
	"fmt"
	"math/rand"

	"github.com/brennie/spaghetti/solver/heuristics"
	"github.com/brennie/spaghetti/tt"
)

const (
	MaxSize = 150 // Maximum number of individuals in a population
	MinSize = 100 // Minimum number of individuals in a population
)

// An individual in the population
type individual struct {
	soln  *tt.Solution // The corresponding solution
	value tt.Value     // The corresponding value
}

// The actual population heap. This is unexported so that the Push and Pop
// methods cannot be used except by this package.
type popHeap []individual

// A population.
type Population struct {
	heap popHeap // The actual heap with population members.
}

// Generate a population of size MinSize using the random variable order
// heuristic.
func New(rng *rand.Rand, inst *tt.Instance) (p *Population) {
	p = &Population{make(popHeap, MinSize, MaxSize)}

	for i := 0; i < MinSize; i++ {
		p.heap[i].soln = inst.NewSolution()
		heuristics.RandomVariableOrdering(p.heap[i].soln, rng)
		p.heap[i].value = p.heap[i].soln.Value()
	}

	heap.Init(p.heap)

	return
}

// Push an element onto the heap. Use Insert instead.
func (heap popHeap) Push(element interface{}) {
	soln, ok := element.(*tt.Solution)

	if !ok {
		panic(fmt.Sprintf("popHeap.Push() expected *tt.Solution; got %v instead", element))
	}

	if len(heap) == cap(heap) {
		panic("popHeap.Push() on a full heap")
	}

	heap = append(heap, individual{soln, soln.Value()})
}

// Remove an element from the population.
func (heap popHeap) Pop() (element interface{}) {
	if len(heap) == 0 {
		panic("Popping from empty Population")
	}
	newLen := len(heap) - 1
	element = heap[newLen].soln
	heap = heap[0:newLen]

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

// Determine the size of the population.
func (p *Population) Size() int {
	return len(p.heap)
}

// Do selection so that the population has at most MinSize members.
func (p *Population) Select() {
	if p.Size() <= MinSize {
		return
	}

	// We do in-place heap sort so that elements [MaxSize - Minsize, MaxSize)
	// are sorted in increasing order. We copy the underlying slice as heap.Pop
	// will shrink it and we need access to the whole thing (so that we can
	// reverse it).
	oldLen := p.Size()
	copy := p.heap[:]
	for i, j := 0, oldLen-1; i < MinSize; i, j = i+1, j-1 {
		copy[j] = heap.Pop(p.heap).(individual)
	}

	// Now we restore the heap. Elmements [MaxSize - Minsize, MaxSize) are
	// currently sorted in decreasing order so we do swaps such that
	// [0, MinSize) is sorted in increasing order. We note that this restores
	// the heap order property so we do not have to do heap.Init()
	p.heap = copy
	for i, j := 0, oldLen-1; i < j; i, j = i+1, j-1 {
		p.heap.Swap(i, j)
	}

	// We can drop all the rest of the elements so that we only have a heap of
	// MinSize elements.
	p.heap = p.heap[0:MinSize]
}

// Insert a member into the population. If the population is full, do selection
// first.
func (p *Population) Insert(soln *tt.Solution) {
	if p.Size() == MaxSize {
		p.Select()
	}

	heap.Push(p.heap, individual{soln, soln.Value()})
}

// Determine the best member of the population. A solution picked this way must
// not be modified.
func (p *Population) Best() (*tt.Solution, tt.Value) {
	return p.heap[0].soln, p.heap[0].value
}

// Pick a member randomly using the given random number generator. A solution
// picked this way must not be modified. To pick a member randomly and modify
// it, use RemoveOne followed by Insert.
func (p *Population) Pick(rng *rand.Rand) *tt.Solution {
	return p.heap[rng.Intn(p.Size())].soln
}

// Remove one solution from the population, chosen at random.
func (p *Population) RemoveOne(rng *rand.Rand) (soln *tt.Solution) {
	index := rng.Intn(p.Size())
	soln = p.heap[index].soln

	heap.Remove(p.heap, index)

	return
}
