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

import "github.com/brennie/spaghetti/tt"

// An individual in the population
type individual struct {
	soln    *tt.Solution // The corresponding solution.
	value   tt.Value     // The corresponding value.
	success *success     // The success ratio of the individual.
}

// Create a new individual from the given solution and optional value. If the
// value is not provided, it will be computed.
func newIndividual(soln *tt.Solution, value ...tt.Value) *individual {
	switch len(value) {
	case 0:
		return &individual{soln, soln.Value(), &success{}}

	case 1:
		return &individual{soln, value[0], &success{}}

	default:
		panic("population: newIndividual: len(value) > 1")
	}
}

// Report the result of a crossover.
func (i *individual) didCrossover(childValue tt.Value) {
	i.success.mutex.Lock()
	if childValue.Less(i.value) {
		i.success.successes++
	}
	i.success.crossovers++
	i.success.mutex.Unlock()
}
