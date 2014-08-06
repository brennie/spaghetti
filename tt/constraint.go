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

package tt

// A constraint pair describes that there is a constraint between two events
// in the instance.
type ConstraintPair struct {
	EventA, EventB int // The two variables with constraints.
}

// Generate a constraint pair such that EventA < EventB
func pair(a, b int) ConstraintPair {
	if a < b {
		return ConstraintPair{a, b}
	} else {
		return ConstraintPair{b, a}
	}
}

// Generate all the constraint pairs and their violations.
func (s *Solution) ConstraintPairs() (pairs map[ConstraintPair]int) {
	pairs = make(map[ConstraintPair]int)

	// There exist constraints between every variable (becuase they cannot share timeslots).
	for i := 0; i < s.inst.NEvents()-1; i++ {
		for j := 1 + 1; j < s.inst.NEvents(); j++ {
			pairs[ConstraintPair{i, j}] = 0
		}
	}

	for student := range s.attendance {
		for time := range s.attendance[student] {
			if nEvents := len(s.attendance[student][time]); nEvents >= 2 {
				for eventA := range s.attendance[student][time] {
					for eventB := range s.attendance[student][time] {
						if eventA < eventB {
							pairs[pair(eventA, eventB)]++
						}
					}
				}
			}
		}
	}

	for ratIndex := range s.events {
		if nEvents := len(s.events[ratIndex]); nEvents >= 2 {
			for eventA := range s.events[ratIndex] {
				for eventB := range s.events[ratIndex] {
					if eventA < eventB {
						pairs[pair(eventA, eventB)]++
					}
				}
			}
		}
	}

	for eventIndex := range s.rats {
		event := &s.inst.events[eventIndex]
		if rat := s.rats[eventIndex]; rat.Assigned() {
			for otherIndex := range event.after {
				if other := s.rats[otherIndex]; other.Assigned() && !other.After(rat) {
					pairs[pair(eventIndex, otherIndex)]++
				}
			}
		}
	}

	return pairs
}

// An event-weight pair.
type WeightedValue struct {
	Event  int // The value
	Weight int // It's weight
}

// A slice of event-weight pairs.
type WeightedValues []WeightedValue

// Determine the length of a pairs.
func (w WeightedValues) Len() int {
	return len(w)
}

// Determine if the weight at p[i] is less than the weight at p[j].
func (w WeightedValues) Less(i, j int) bool {
	return w[i].Weight < w[j].Weight
}

// Swap p[i] and p[j]
func (w WeightedValues) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}
