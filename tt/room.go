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

// A room.
type room struct {
	capacity int          // The capacity of the room.
	features map[int]bool // The features that the room has.
}

// Determine if room can host a given event, determined by the number of
// students attending it and the features it requires.
func (r *room) canHost(e *event) bool {
	if r.capacity < len(e.students) {
		return false
	}

	for feature := range e.features {
		if e.features[feature] && !r.features[feature] {
			return false
		}
	}

	return true
}
