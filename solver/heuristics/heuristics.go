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

package heuristics

import (
	"container/heap"
	"math/rand"

	"github.com/brennie/spaghetti/tt"
	"github.com/brennie/spaghetti/tt/pqueue"
)

// Do most constrained variable first search to filla s much of the domains of
// the solution as possible.
func MostConstrainedOrdering(soln *tt.Solution) *tt.Solution {
	pq := pqueue.New(soln.Domains)

	for pq.Len() > 0 {
		event := heap.Pop(pq).(int)
		soln.Best(event)
		pq.Update()
	}

	return soln
}

// Use random variable ordering to fill as much of the domains of the solution
// as possible. We do not try to find the best element in the domain to assign;
// we only try to fill up the domain as fast as possible.
func RandomVariableOrdering(soln *tt.Solution) *tt.Solution {
	for _, event := range rand.Perm(len(soln.Domains)) {
		soln.Best(event)
	}

	return soln
}

// Randomly assign a solution by using a random variable ordering and picking
// random domain entries in (non-empty) domains to assign to them.
func RandomAssignment(soln *tt.Solution) *tt.Solution {
	for _, event := range rand.Perm(len(soln.Domains)) {
		if soln.Domains[event].Entries.Size() > 0 {
			el := soln.Domains[event].Entries.First()
			for offset := rand.Intn(soln.Domains[event].Entries.Size()); offset > 0; offset-- {
				el = el.Next()
			}
			soln.Assign(event, el.Value().(tt.Rat))
		}
	}
	return soln
}
