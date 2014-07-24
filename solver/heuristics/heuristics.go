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
	"math/rand"

	"github.com/brennie/spaghetti/set"
	"github.com/brennie/spaghetti/tt"
)

// Randomly assign a solution by using a random variable ordering and picking
// random domain entries in (non-empty) domains to assign to them.
func RandomAssignment(soln *tt.Solution) *tt.Solution {
	for _, event := range rand.Perm(len(soln.Domains)) {
		rat := soln.Domains[event][rand.Intn(len(soln.Domains[event]))]
		soln.Assign(event, rat)
	}
	return soln
}

// Follow the given variable ordering
func RandomAssignmentWithOrdering(soln *tt.Solution, ordering []int) *tt.Solution {
	domains := make([]set.Set, len(soln.Domains))
	for event := range domains {
		domains[event] = set.FromList(tt.RatCompare, soln.Domains[event])
	}

	unassigned := make(map[int]bool)

	for _, event := range ordering {
		if domains[event].Size() == 0 {
			unassigned[event] = true
		} else {
			el := domains[event].First()

			for offset := rand.Intn(domains[event].Size()); offset > 0; offset-- {
				el = el.Next()
			}

			soln.AssignAndShrink(event, el.Value().(tt.Rat), domains)
		}
	}

	for event := range unassigned {
		ratIndex := rand.Intn(len(soln.Domains[event]))
		soln.Assign(event, soln.Domains[event][ratIndex])
	}

	for event := range domains {
		domains[event].Clear()
	}

	return soln
}
