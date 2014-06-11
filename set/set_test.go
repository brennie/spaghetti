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

package set

import (
	"math/rand"
	"sort"
	"testing"
)

func intCmp(a, b interface{}) Order {
	aVal := a.(int)
	bVal := b.(int)

	switch {
	case aVal > bVal:
		return Gt

	case aVal < bVal:
		return Lt

	default:
		return Eq
	}
}

const testSize = 1000

func TestSet(t *testing.T) {
	slice := make([]int, testSize)
	set := New(intCmp)
	for i := range(slice) {
		slice[i] = rand.Int()
		set.Insert(slice[i])
	}

	sort.Ints(slice)

	if el := set.Find(slice[0]); el == nil {
		t.Error("Could not find member")
	}

	el := set.First()
	for _, value := range(slice) {
		if value != el.Value().(int) {
			t.Error("Not in sorted order")
		}
		el = el.Next()
	}

	for _, value := range(slice) {
		set.Remove(value)
	}

	if set.Size() != 0 {
		t.Error("Set should be empty")
	}
}
