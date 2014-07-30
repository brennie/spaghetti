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
	"math/rand"

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

// Determine if the sub-population is full.
func (p *SubPopulation) Full() bool {
	return p.length == p.maxSize
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

// Return an individual so that it may be crossed over.
func (p *SubPopulation) PickIndividual() *Individual {
	if p.length == 0 {
		panic("SubPopulation.PickIndividual: empty SubPopulation")
	}
	picked := rand.Intn(p.length)
	return p.pop[picked].export()
}

// Pick a member of the population randomly.
func (p *SubPopulation) PickSolution() []tt.Rat {
	if p.length == 0 {
		panic("SubPopulation.PickSolution: empty SubPopulation")
	}
	picked := rand.Intn(p.length)
	return p.pop[picked].soln.Assignments()
}

// Remove one solution from th population, chosen at random.
func (p *SubPopulation) RemoveOne() (soln *tt.Solution) {
	if p.length == 0 {
		panic("SubPopulation.RemoveOne: empty SubPopulation")
	} else if p.length == 1 {
		soln = p.pop[0].soln
		p.pop[0] = nil
	} else {
		picked := rand.Intn(p.length)
		soln = p.pop[picked].soln
		p.length--
		p.pop[picked] = p.pop[p.length]
		p.pop[p.length] = nil
	}

	return
}
