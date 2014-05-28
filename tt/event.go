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

// A timetabling event, i.e. a class or an exam.
type event struct {
	id       int          // The event's identifier.
	times    [45]bool     // The times in which the event can be scheduled.
	rooms    map[int]bool // The rooms in which the event can be scheduled.
	before   map[int]bool // The events which happen before this event.
	after    map[int]bool // The events which happen after this event.
	students map[int]bool // The students which attend this event.
	exclude  map[int]bool // The events that cannot occur at the same time as this event.
}
