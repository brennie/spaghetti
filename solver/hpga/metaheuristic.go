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
	"sort"

	"github.com/brennie/spaghetti/tt"
)

// A meta-heuristic which updates the variable ordering and value weighting.
type metaheuristic struct {
	varWeights tt.WeightedValues // The variable weights.
	valWeights []map[tt.Rat]int  // The value weights.
	vars       map[int]int       // Since varWeights is sorted, we need a map for event -> varWeights index.
	temp       []int             // Temp storage for var violations and getting variable order.
}

// Create a new meta-heuristic
func newMH(varWeights tt.WeightedValues, valWeights []map[tt.Rat]int) (mh *metaheuristic) {
	mh = &metaheuristic{
		make(tt.WeightedValues, len(varWeights)),
		make([]map[tt.Rat]int, len(valWeights)),
		make(map[int]int),
		make([]int, len(varWeights)),
	}

	for event := range varWeights {
		mh.varWeights[event].Event = varWeights[event].Event
		mh.varWeights[event].Weight = varWeights[event].Weight
		mh.vars[event] = event
	}

	for event := range valWeights {
		mh.valWeights[event] = make(map[tt.Rat]int)
		for rat := range valWeights[event] {
			mh.valWeights[event][rat] = valWeights[event][rat]
		}
	}

	return mh
}

// Update the metaheuristic with a solution.
func (mh *metaheuristic) update(soln *tt.Solution) {
	for constraint, violations := range soln.ConstraintPairs() {
		if violations > 0 {
			mh.temp[constraint.EventA] += violations
			mh.temp[constraint.EventB] += violations
		}
	}

	for event, count := range mh.temp {
		rat := soln.RatAt(event)
		if count == 0 {
			mh.valWeights[event][rat]++
		} else {
			mh.valWeights[event][rat]--
			mh.varWeights[event].Weight += count
		}
		mh.temp[event] = 0
	}

	// Translate weights so they stay totally-positive.
	for event := range mh.valWeights {
		for rat := range mh.valWeights[event] {
			mh.valWeights[event][rat]++
		}
	}
}

// Get the variable ordering.
func (mh *metaheuristic) getVarOrdering() []int {
	sort.Sort(mh.varWeights)

	for i := range mh.varWeights {
		event := mh.varWeights[i].Event
		mh.vars[event] = i
		mh.temp[i] = event
	}

	return mh.temp
}
