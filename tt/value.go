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

import (
	"fmt"
	"math"
)

// A solution valuation.
type Value struct {
	Violations int // The number of hard constraint violations.
	Fitness    int // The solution fitness.
}

// Determine if we have found an ideal value.
func (v Value) IsIdeal() bool {
	return v.Violations == 0 && v.Fitness == 0
}

// Compare two values using a lexicographical compare.
func (v Value) Less(u Value) bool {
	switch {
	case v.Violations < u.Violations:
		return true

	case v.Violations == u.Violations && v.Fitness < u.Fitness:
		return true

	default:
		return false
	}
}

// Format a value as a 2-tuple of the distance at fitness
func (v Value) String() string {
	return fmt.Sprintf("(%d, %d)", v.Violations, v.Fitness)
}

// The worst possible value.
func WorstValue() Value {
	return Value{math.MaxInt32, math.MaxInt32}
}
