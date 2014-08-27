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

// The package describing genetic algorithm populations.
package population

import "fmt"

type population []*individual

type Population struct {
	pop     population // The total population
	minSize int        // The minimum size of a sub-population
	maxSize int        // The maximum size of a sub-population
	count   int        // The number of sub-populations

	subPops []*SubPopulation // The sub-populations

	temp [][]*individual // Slices for doing selection

}

// Create a new population of size coun
func New(minSize, maxSize, count int) *Population {
	if maxSize <= minSize {
		panic(fmt.Sprintf("population.New: maxSize (%d) <= minSize (%d)", maxSize, minSize))
	}

	p := &Population{
		make([]*individual, maxSize*count),
		minSize,
		maxSize,
		count,
		make([]*SubPopulation, count),
		make([][]*individual, count),
	}

	for i := range p.subPops {
		p.subPops[i] = &SubPopulation{
			p.pop[i*maxSize : (i+1)*maxSize],
			0,
			minSize,
			maxSize,
		}

		p.temp[i] = make([]*individual, minSize)
	}

	return p
}

// Get the sub-population at the given index.
func (p *Population) SubPopulation(index int) *SubPopulation {
	if index > p.count {
		panic("Population.SubPopulation: index out of range")
	}

	return p.subPops[index]
}

func (p *Population) IsSubPopulationFull(index int) bool {
	if index > p.count {
		panic("Population.SubPopulation: index out of range")
	}

	return p.subPops[index].IsFull()
}

// Return the length of the population.
//
// NB: This will always return the same number because we are not appending to
// or slicing the population.
func (p population) Len() int {
	return len(p)
}

// Compare two members of the population.
func (p population) Less(i, j int) bool {
	return p[i].value.Less(p[j].value)
}

// Swap two members of the population.
func (p population) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
