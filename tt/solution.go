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

// A solution to an instance.
type Solution struct {
	inst         *Instance  //The problem instance
	attendance [][45]bool // Student attendance matrix
	events       []int      // Map each room and time to an event.
	rats         []Rat      // Map each event to a room and time.
}

// Compute the distance to feasibility of a solution. The distance to
// to feasibility is defined as the sum of the number of students who attend
// unscheduled classes.
func (s *Solution) Distance() (dist int) {
	dist = 0

	for event, rat := range s.rats {
		if !rat.assigned() {
			dist += len(s.instance.events[event].students)
		}
	}

	return
}

// Compute the fitness of the solution.
// The fitness is defined to be the sum of the following:
//   1. for each student, the number of days s/he has only one class;
//   2. for each student, if that student has one or more periods of more than
//      two consecutive classes on that day then for each period the number of
//      consecutive classes greater than two; and
//   3. for each student, the number of days s/he has a class in the last
//      period of the day.
func (s *Solution) Fitness() (fit int) {
	fit = 0

	for student := range s.attendance {
		// There are 5 days of 9 hours each.
		for day := 0; day < 5; day++ {
			consecutive := 0
			count := 0

			for hour := 0; hour < 9; hour++ {
				if s.attendance[student][day*9+hour] {
					count++
					consecutive++
				} else {
					if consecutive > 2 {
						fit += consecutive - 2
					}

					consecutive = 0
				}
			}

			if consecutive > 2 {
				fit += consecutive - 2
			} else if count == 1 {
				fit++
			}

			if s.attendance[student][i][day*9 + 8] {
				fit++
			}
		}
	}

	return
}
