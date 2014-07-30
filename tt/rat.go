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

// An assignment of a room and time.
type Rat struct {
	Room int // The assigned room.
	Time int // The assigned time.
}

// Determine if the given room and time is an assigned room and time.
func (r Rat) Assigned() bool {
	return r.Room != badRat.Room && r.Time != badRat.Time
}

// Determine the index of the room and time in the array of all rooms and
// times.
func (r Rat) index() int {
	return r.Room*NTimes + r.Time
}

// Build a Rat from an index.
func ratFromIndex(index int) Rat {
	return Rat{index / NTimes, index % NTimes}
}

func (r Rat) After(q Rat) bool {
	return r.Time > q.Time
}

func (r Rat) Before(q Rat) bool {
	return r.Time < q.Time
}
