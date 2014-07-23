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

	"github.com/brennie/spaghetti/tt"
)

const (
	pMutate = 5  // The probability of a mutation is 5%
	pLocal  = 75 // The probability of a local crossover is 75%
)

// Mutate a solution at the given chromosome, giving that chromosome the
// optimal value in the domain (such that it does not violate a hard
// constraint).
func mutate(mutant *tt.Solution, chromosome int) {
	rat := mutant.Domains[chromosome][rand.Intn(len(mutant.Domains[chromosome]))]
	mutant.Assign(chromosome, rat)
}

// Perform genetic mutation on the given solution, which is a more extreme
// form of the mutate operator. This operater does unassignemnt and then re-
// assigns to find the best overall value for the chromosome and then re-fill
// the domain by using the most constrained ordering heuristic.
//
// XXX: This should actually do something.
func gm(mutant *tt.Solution, chromosome int) {
	mutate(mutant, chromosome)
}
