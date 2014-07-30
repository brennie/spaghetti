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
	"sort"

	"github.com/brennie/spaghetti/tt"
)

type parentMask bool

const (
	useMother parentMask = false // Mask value signalling to use the mother
	useFather parentMask = true  // Mask value signalling to use the father
	maxMutate float64    = 0.2   // The maximum percentage of an individual to mutate
)

func (p *Population) Crossover(motherPop, fatherPop int, inst *tt.Instance) (child *tt.Solution, value tt.Value) {
	if motherPop > p.count || fatherPop > p.count {
		panic("Population.Crossover: population out of bounds")
	}

	mother := p.subPops[motherPop].pop[rand.Intn(p.subPops[motherPop].length)]
	father := p.subPops[fatherPop].pop[rand.Intn(p.subPops[fatherPop].length)]

	child, value = crossover(mother, father, inst)
	p.subPops[motherPop].Insert(child, value)

	return
}

func (p *SubPopulation) Crossover(inst *tt.Instance) (*tt.Solution, tt.Value) {
	mIndex := rand.Intn(p.length)
	fIndex := rand.Intn(p.length - 1)
	if fIndex >= mIndex {
		fIndex++
	}

	return crossover(p.pop[mIndex], p.pop[fIndex], inst)
}

func crossover(mother, father *individual, inst *tt.Instance) (child *tt.Solution, value tt.Value) {
	pMother := float64(0.5 + (mother.success.ratio()-father.success.ratio())*0.5)

	child = inst.NewSolution()
	for event := 0; event < inst.NEvents(); event++ {
		parent := mother
		if mask(mother, father, event, pMother) == useFather {
			parent = father
		}

		if rat := parent.soln.RatAt(event); rat.Assigned() {
			child.Assign(event, rat)
		}
	}
	value = child.Value()

	mother.didCrossover(value)
	father.didCrossover(value)

	return
}

// Generate the crossover mask for the specific event in the two individuals.
func mask(mother, father *individual, event int, pMother float64) parentMask {
	mQual := mother.soln.AssignmentQuality(event)
	fQual := father.soln.AssignmentQuality(event)

	if mQual.Less(fQual) {
		return useMother
	} else if fQual.Less(mQual) {
		return useFather
	} else if rand.Float64() < pMother {
		return useMother
	} else {
		return useFather
	}
}

// Mutate one member of a given population and return the solution and the value
func (p *SubPopulation) MutateOne() (mutant *tt.Solution, value tt.Value) {
	picked := rand.Intn(p.length)
	mutant = p.pop[picked].soln.Clone()

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

	value = mutant.Value()

	return
}

// Perform selection.
func (pop *Population) Select() {
	sort.Sort(pop.pop)

	p := 0
	i := 0
	direction := +1

	// We do snake/wraparound picking to try to make the sub-populations as
	// well-balanced as possible. We copy into the temp arrays so that we can
	// distribute without overwriting.
	for pick := 0; pick < pop.count*pop.minSize; pick++ {
		pop.temp[p][i] = pop.pop[pick]

		if (p == 0 && direction == -1) || (p == pop.count-1 && direction == +1) {
			// Switch direction at the boundaries
			direction = -direction
			i++
		} else {
			p += direction
		}
	}

	for i := pop.count * pop.minSize; i < pop.count*pop.maxSize; i++ {
		pop.pop[i].soln.Free()
		pop.pop[i] = nil
	}

	// Distribute the individuals back to the sub-populations, then clear the
	// remaining entries and reset the sub-populations lengths.
	for p := range pop.temp {
		for i := range pop.temp[p] {
			pop.subPops[p].pop[i] = pop.temp[p][i]
			pop.temp[p][i] = nil
		}

		for i := pop.minSize; i < pop.maxSize; i++ {
			pop.subPops[p].pop[i] = nil
		}

		pop.subPops[p].length = pop.minSize
	}
}
