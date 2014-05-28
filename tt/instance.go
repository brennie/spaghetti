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

// The timetabling package.
package tt

// An instance of a timetabling problem.
type Instance struct {
	nEvents   int            // The number of events in the instance.
	nRooms    int            // The number of rooms in the instance.
	nFeatures int            // The number of features in the instance.
	nStudents int            // The number of students in the instance.
	events    []event        // The events in the instance.
	students  []map[int]bool // The attendance of the students in the instance.
}
