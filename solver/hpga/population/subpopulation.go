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

import (
	"github.com/brennie/spaghetti/solver/heuristics"
	"github.com/brennie/spaghetti/tt"
)

// A Sub-Population
type SubPopulation struct {
	pop     []*individual // The slice into the larger population.
	length  int           // The length of the sub-population
	minSize int           // The minimum size
	maxSize int           // the maxium size
}

// Generate minPop individuals randomly.
func (p *SubPopulation) Generate(inst *tt.Instance) (bestSoln *tt.Solution, bestValue tt.Value) {
	bestSoln = nil
	bestValue = tt.WorstValue()

	for p.length < p.minSize {
		soln := heuristics.RandomAssignment(inst.NewSolution())
		value := soln.Value()

		if value.Less(bestValue) {
			bestValue = value
			bestSoln = soln
		}

		p.Insert(soln, value)
	}

	return
}

// Insert a solution into the sub-population. Value is an optional parameter
// and will be calculated if it is not provided.
func (p *SubPopulation) Insert(soln *tt.Solution, value ...tt.Value) {
	if p.length == p.maxSize {
		panic("SubPopulation.Insert: full SubPopulation")
	}

	switch len(value) {
	case 0:
		p.pop[p.length] = newIndividual(soln, soln.Value())

	case 1:
		p.pop[p.length] = newIndividual(soln, value[0])

	default:
		panic("SubPopulation.Insert: Multiple values supplied")
	}

	p.length++
}

// Determine if the sub-population is full.
func (p *SubPopulation) IsFull() bool {
	return p.length == p.maxSize
}

// Get the length of the sub-population
func (p *SubPopulation) Len() int {
	return p.length
}
