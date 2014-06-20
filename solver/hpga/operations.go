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
	"github.com/brennie/spaghetti/solver/heuristics"
	"github.com/brennie/spaghetti/tt"
)

// Perform a crossover between the mother and father, yielding the child. The
// child gets the rats from events [0, chromosome) from the mother and rats
// from events [chromosome, nEvents) from the father (assuming that those do
// not have any conflicts).
func crossover(mother, father, child *tt.Solution, chromosome int) {
	for event := 0; event < chromosome; event++ {
		if mother.Assigned(event) {
			child.Assign(event, mother.RatAt(event))
		}
	}

	for event := chromosome; event < len(child.Domains); event++ {
		if father.Assigned(event) {
			if rat := father.RatAt(event); !child.Domains[event].HasConflict(rat) {
				child.Assign(event, rat)
			}
		}
	}
}

// Mutate a solution at the given chromosome, giving that chromosome the
// optimal value in the domain. This is followed by most-constrained variable
// first search to fill up the rest of the domain.
func mutate(mutant *tt.Solution, chromosome int) {
	mutant.Unassign(chromosome)
	mutant.RemoveConflicts(chromosome)
	mutant.Best(chromosome)

	heuristics.MostConstrainedOrdering(mutant)
}