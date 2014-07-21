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

// The package describing genetic algorithm populations.
package population

import (
	"container/heap"
	"fmt"
	"math/rand"

	"github.com/brennie/spaghetti/solver/heuristics"
	"github.com/brennie/spaghetti/tt"
)

// A population.
type Population struct {
	heap    popHeap // The actual heap with population members.
	minSize int     // The minimum number of individuals in the population.
	maxSize int     // The maximum number of individuals in the population.
}

// Generate a population of size minSize using the random variable order
// heuristic.
func New(inst *tt.Instance, minSize, maxSize int) (p *Population) {
	if maxSize <= minSize {
		panic(fmt.Sprintf("population.New: maxSize (%d) <= minSize (%d)", maxSize, minSize))
	}

	p = &Population{
		make(popHeap, minSize, maxSize),
		minSize,
		maxSize,
	}

	for i := 0; i < minSize; i++ {
		p.heap[i] = newIndividual(heuristics.RandomVariableOrdering(inst.NewSolution()))
	}

	heap.Init(&p.heap)

	return
}

// Determine the size of the population.
func (p *Population) Size() int {
	return len(p.heap)
}

// Do selection so that the population has at most MinSize members.
func (p *Population) Select() {
	if p.Size() <= p.minSize {
		return
	}

	// We do in-place heap sort so that elements [MaxSize - Minsize, MaxSize)
	// are sorted in increasing order. We copy the underlying slice as heap.Pop
	// will shrink it and we need access to the whole thing (so that we can
	// reverse it).
	oldLen := p.Size()
	copy := p.heap[:]
	for i, j := 0, oldLen-1; i < p.minSize; i, j = i+1, j-1 {
		// heap.Pop will only return the *tt.Solution part of the underlying
		// individual. Hence we copy the value before popping so that we can
		// save it and put it in the appropriate place without re-calculating
		// the value of the solution.
		value := p.heap[0]
		heap.Pop(&p.heap)
		copy[j] = value
	}

	// Now we restore the heap. Elmements [MaxSize - Minsize, MaxSize) are
	// currently sorted in decreasing order so we do swaps such that
	// [0, MinSize) is sorted in increasing order. We note that this restores
	// the heap order property so we do not have to do heap.Init()
	p.heap = copy
	for i, j := 0, oldLen-1; i < j; i, j = i+1, j-1 {
		p.heap.Swap(i, j)
	}

	// Nil all pointers that are used in the no-longer-needed solutions.
	for i := p.minSize; i < oldLen; i++ {
		p.heap[i].soln.Free()
		p.heap[i].soln = nil
		p.heap[i].success = nil
		p.heap[i] = nil
	}

	// We can drop all the rest of the elements so that we only have a heap of
	// MinSize elements.
	p.heap = p.heap[0:p.minSize]
}

// Insert a member into the population. If the population is full, do selection
// first.
func (p *Population) Insert(soln *tt.Solution) {
	if p.Size() == p.maxSize {
		p.Select()
	}

	heap.Push(&p.heap, newIndividual(soln))
}

// Determine the best member of the population. A solution picked this way must
// not be modified.
func (p *Population) Best() (*tt.Solution, tt.Value) {
	return p.heap[0].soln, p.heap[0].value
}

// Pick a member randomly using the given random number generator. A solution
// picked this way must not be modified. To pick a member randomly and modify
// it, use RemoveOne followed by Insert.
func (p *Population) PickSolution() []tt.Rat {
	return p.heap[rand.Intn(p.Size())].soln.Assignments()
}

// Remove one solution from the population, chosen at random.
func (p *Population) RemoveOne() (soln *tt.Solution) {
	index := rand.Intn(p.Size())
	soln = p.heap[index].soln

	heap.Remove(&p.heap, index)

	return
}
