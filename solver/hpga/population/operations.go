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

	"github.com/brennie/spaghetti/tt"
)

type parentMask bool

const (
	useMother parentMask = false // Mask value signalling to use the mother
	useFather parentMask = true  // Mask value signalling to use the father
	maxMutate float64    = 0.2   // The maximum percentage of an individual to mutate
)

// Do a crossover between the mother and the father into the (empty) child
// solution and return its value.
func Crossover(mother, father *Individual, child *tt.Solution) (childValue tt.Value) {
	pMother := float64(0.5 + (mother.Success.Ratio()-father.Success.Ratio())*0.5)
	for event := range mother.Assignments {
		parent := mother
		if mask(mother, father, event, pMother) == useFather {
			parent = father
		}

		if rat := parent.Assignments[event]; rat.Assigned() {
			child.Assign(event, rat)
		}
	}
	childValue = child.Value()

	mother.didCrossover(childValue)
	father.didCrossover(childValue)

	return childValue
}

// Generate the crossover mask for the specific event in the two individuals.
func mask(mother, father *Individual, event int, pMother float64) parentMask {
	if mother.Quality[event].Less(father.Quality[event]) {
		return useMother
	} else if father.Quality[event].Less(mother.Quality[event]) {
		return useFather
	} else if rand.Float64() < pMother {
		return useMother
	} else {
		return useFather
	}

}

// Mutate a solution.
func Mutate(mutant *tt.Solution) (value tt.Value) {
	nEvents := mutant.NEvents()
	max := int(maxMutate * float64(nEvents))
	nMutations := rand.Intn(max) + 1 // nMutations is in the range [1, max]

	toMutate := make(map[int]bool)
	for len(toMutate) < nMutations {
		chromosome := rand.Intn(nEvents)
		toMutate[chromosome] = true
	}

	for event := range toMutate {
		rat := mutant.Domains[event][rand.Intn(len(mutant.Domains[event]))]
		mutant.Assign(event, rat)
	}

	return mutant.Value()
}

// Mutate one member of a given population and return the solution and the value
func (p *Population) MutateOne() (mutant *tt.Solution, value tt.Value) {
	mutant = p.RemoveOne()
	value = Mutate(mutant)
	return
}
