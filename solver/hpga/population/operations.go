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
	useMother parentMask = false
	useFather parentMask = true
)

func Crossover(mother, father *Individual, child *tt.Solution) (childValue tt.Value) {
	for event := range mother.Assignments {
		parent := mother
		if mask(mother, father, event) == useFather {
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

func mask(mother, father *Individual, event int) parentMask {

	pMother := float64(0.5 + (mother.Success.Ratio()-father.Success.Ratio())*0.5)

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
