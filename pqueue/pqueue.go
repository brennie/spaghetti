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

// A priority queue data structure that sorts by domain size.
package pqueue

import (
	"container/heap"
	"fmt"

	"github.com/brennie/spaghetti/tt"
)

// A priority queue for domains. This allows us to use the most constrained
// variable first heuristic. This data structure should be interfaced with the
// container/heap package through e.g. heap.Pop(pq)
//
// Since heaps must use interface{} types, the result of heap.Pop(pq) should be
// casted to int; this int is the event's index in the given solution. In the
// case of an empty queue, heap.Pop will result in a panic from an out-of-range index.
type PQueue struct {
	domains []tt.Domain
	sizes   []int
	heap    []int
}

// Create a new priority queue.
func New(domains []tt.Domain) (pq *PQueue) {
	size := len(domains)

	pq = &PQueue{
		domains,
		make([]int, size),
		make([]int, size),
	}

	for event := range pq.heap {
		pq.heap[event] = event
	}

	pq.Update()

	return
}

// Update the priority queue to restore the heap-ordering principle.
func (pq *PQueue) Update() {
	for i := range pq.domains {
		pq.sizes[i] = len(pq.domains[i].Entries)
	}

	heap.Init(pq)
}

// Push a new element onto the heap. This will panic if element is not an int.
func (pq *PQueue) Push(element interface{}) {
	index, ok := element.(int)

	if !ok {
		panic(fmt.Sprintf("PQueue.Push() expected an int; got %v instead",
			element))
	}

	pq.heap = append(pq.heap, index)
}

// Pop an element from the heap.
func (pq *PQueue) Pop() (element interface{}) {
	size := pq.Len()

	if size == 0 {
		element = nil
	} else {
		element = pq.heap[size-1]
		pq.heap = pq.heap[0 : size-1]
	}

	return
}

// Get the length of the priority queue.
func (pq *PQueue) Len() int {
	return len(pq.heap)
}

// Determine which event has a smaller domain.
func (pq *PQueue) Less(i, j int) bool {
	return pq.sizes[pq.heap[i]] < pq.sizes[pq.heap[j]]
}

// Swap two events in the priority queue.
func (pq *PQueue) Swap(i, j int) {
	pq.heap[i], pq.heap[j] = pq.heap[j], pq.heap[i]
}
