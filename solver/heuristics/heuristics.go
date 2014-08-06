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

	"github.com/brennie/spaghetti/tt"
)

// Randomly assign a solution by using a random variable ordering and picking
// random domain entries in (non-empty) domains to assign to them.
func RandomAssignment(soln *tt.Solution) *tt.Solution {
	for _, event := range rand.Perm(soln.NEvents()) {
		rat := soln.Domains[event][rand.Intn(len(soln.Domains[event]))]
		soln.Assign(event, rat)
	}
	return soln
}

// Follow the given variable ordering
func RandomAssignmentWithOrdering(soln *tt.Solution, ordering []int) *tt.Solution {
	domains := soln.MakeShrinkableDomains()
	unassigned := make(map[int]bool)

	for _, event := range ordering {
		domainSize := len(domains[event])
		if domainSize == 0 {
			unassigned[event] = true
		} else {
			var rat tt.Rat
			offset := rand.Intn(len(domains[event]))
			i := 0

			for rat = range domains[event] {
				if i == offset {
					break
				}
				i++
			}

			soln.AssignAndShrink(event, rat, domains)
		}
	}

	for event := range unassigned {
		ratIndex := rand.Intn(len(soln.Domains[event]))
		soln.Assign(event, soln.Domains[event][ratIndex])
	}

	return soln
}

func OrderedWeightedAssignment(soln *tt.Solution, varOrdering []int, valWeights []map[tt.Rat]int) *tt.Solution {
	unassigned := make(map[int]bool)
	domains := soln.MakeShrinkableDomains()

	for _, event := range varOrdering {
		if len(domains[event]) == 0 {
			unassigned[event] = true
			continue
		}

		totalWeight := 0
		var chosenRat tt.Rat
		// We do a weighted reservoir sampling of size 1.
		for rat, weight := range valWeights[event] {
			if _, inDomain := domains[event][rat]; inDomain {
				totalWeight += weight
				if rand.Intn(totalWeight) < weight {
					chosenRat = rat
				}
			}
		}

		soln.AssignAndShrink(event, chosenRat, domains)
	}

	for event := range unassigned {
		ratIndex := rand.Intn(len(soln.Domains[event]))
		soln.Assign(event, soln.Domains[event][ratIndex])
	}

	return soln
}
