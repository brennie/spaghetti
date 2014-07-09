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

import "fmt"

// A solution valuation.
type Value struct {
	Distance int // The distance to feasibility.
	Fitness  int // The solution fitness.
}

// Compare two values using a lexicographical compare.
func (v Value) Less(u Value) bool {
	switch {
	case v.Distance < u.Distance:
		return true

	case v.Distance == u.Distance && v.Fitness < u.Fitness:
		return true

	default:
		return false
	}
}

// Format a value as a 2-tuple of the distance at fitness
func (v Value) String() string {
	return fmt.Sprintf("(%d, %d)", v.Distance, v.Fitness)
}
