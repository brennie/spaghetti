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
	"io"

	"github.com/brennie/spaghetti/set"
)

// A solution to an instance.
type Solution struct {
	inst       *Instance          // The problem instance.
	attendance [][45]map[int]bool // Student attendance matrix.
	events     []map[int]bool     // Map each room and time to an event.
	rats       []Rat              // Map each event to a room and time.
	Domains    [][]Rat            // The domains.
}

func (s *Solution) attends(student, time int) bool {
	return len(s.attendance[student][time]) > 0
}

// Retrieve the assignments of a solution as a copy. This is a lighter-weight
// alternative to cloning the whole solution (which requires cloning the
// domains and often isn't necesssary).
func (s *Solution) Assignments() (assignments []Rat) {
	assignments = make([]Rat, s.inst.nEvents)
	for event := range s.rats {
		assignments[event] = s.rats[event]
	}
	return
}

// Determine if the event has been assigned to.
func (s *Solution) Assigned(eventIndex int) bool {
	if eventIndex > s.inst.nEvents {
		return true
	} else {
		return s.rats[eventIndex].Assigned()
	}
}

// Assign an event to a room and time.
func (s *Solution) Assign(event int, rat Rat) {
	if event > s.inst.nEvents {
		panic("Solution.Assign: event > nEvents")
	} else if rat.index() > s.inst.nRooms*NTimes {
		panic("Solution.Assign: invalid Rat")
	}

	ratIndex := rat.index()

	// If the event is already assigned to a room and time (oldRat), we
	// unassign it and replace the entries in the attendance matrix.
	//
	// Otherwise, we just add the new entries to the attendance matrix.
	if oldRat := s.rats[event]; oldRat.Assigned() {
		s.events[oldRat.index()][event] = false

		for student := range s.inst.events[event].students {
			delete(s.attendance[student][oldRat.Time], event)
			s.attendance[student][rat.Time][event] = true
		}
	} else {
		for student := range s.inst.events[event].students {
			s.attendance[student][rat.Time][event] = true
		}
	}

	s.rats[event] = rat
	s.events[ratIndex][event] = true
}

func (s *Solution) AssignAndShrink(event int, rat Rat, domains []set.Set) {
	s.Assign(event, rat)

	// Remove the assignment from the domains of all events.
	for other := range domains {
		if event != other {
			domains[other].Remove(rat)
		}
	}

	// Remove the time slot from all events that share a student.
	for exclude := range s.inst.events[event].exclude {
		for room := 0; room < s.inst.nRooms; room++ {
			domains[exclude].Remove(Rat{room, rat.Time})
		}
	}

	// Remove the domain entries from all events that must occur before it.
	for before := range s.inst.events[event].before {
		for room := 0; room < s.inst.nRooms; room++ {
			for time := rat.Time; time < NTimes; time++ {
				domains[before].Remove(Rat{room, time})
			}
		}
	}

	// Remove the domain entries from all events that must occur after it.
	for after := range s.inst.events[event].after {
		for room := 0; room < s.inst.nRooms; room++ {
			for time := 0; time <= rat.Time; time++ {
				domains[after].Remove(Rat{room, time})
			}
		}
	}

}

// Determine the quality of each assignment, as determined by the number of
// soft constraints each event breaks.
//
// It is not the case that the solution value is the sum of the assignment
// quality, as penalties will be counted for each event involved that violates
// the constraint.
func (s *Solution) AssignmentQualities() (quality []Value) {
	quality = make([]Value, s.inst.nEvents)

	for event := range quality {
		quality[event] = s.AssignmentQuality(event)
	}

	return
}

// Determine the quality of a specific assignment, as determined by the number
// of hard constraints and soft constraints it breaks.
func (s *Solution) AssignmentQuality(event int) (quality Value) {
	if event > s.inst.nEvents {
		panic("Solution.AssignmentQuality: event > nEvents")
	}

	nStudents := len(s.inst.events[event].students)
	time := s.rats[event].Time
	startOfDay := time - time%9
	endOfDay := startOfDay + 8

	// We find the number of consecutive events that the
	// event is a part of.
	for student := range s.inst.events[event].students {
		blockStart := time
		consecutive := 0

		for blockStart > startOfDay && len(s.attendance[student][blockStart-1]) > 0 {
			blockStart--
		}
		for t := blockStart; t <= endOfDay && len(s.attendance[student][blockStart]) > 0; t++ {
			consecutive++
		}

		if consecutive > 2 {
			quality.Fitness += consecutive - 2
		} else {
			// Find the total number of events in the day
			count := 0
			for t := startOfDay; t <= endOfDay; t++ {
				if len(s.attendance[student][t]) > 0 {
					count++
				}
			}

			if count == 1 {
				quality.Fitness++
			}
		}

		if nEvents := len(s.attendance[student][time]); nEvents >= 2 {
			quality.Violations += (nEvents * (nEvents - 1)) / 2
		}
	}

	// If the event is scheduled at the end of the day, then the
	// penalty is the number of students that would attend the event.
	if time == endOfDay {
		quality.Fitness += nStudents
	}

	// If there are multiple assignments to the event's room and time, then
	// the penalty is the number of assignments minus one.
	quality.Violations += len(s.events[s.rats[event].index()]) - 1

	// We consider ordering violations for events. Unlike in Violations(),
	// we consider both the before and after relations because we are
	// concerned with how well we have assigned each variable, not the
	// quality of the overall solution.
	for after := range s.inst.events[event].after {
		if s.rats[after].Time <= s.rats[event].Time {
			quality.Violations++
		}
	}

	for before := range s.inst.events[event].before {
		if s.rats[before].Time >= s.rats[event].Time {
			quality.Violations++
		}
	}

	return
}

// Determine if an assigned event has any hard constraint violations.
func (s *Solution) HasViolations(event int) bool {
	if event > s.inst.nEvents {
		panic("Solution.HasConflicts: event > nEvents")
	}

	if len(s.attendance[student][time]) >= 2 {
		return true
	}

	if len(s.events[s.rats[event].index()]) > 1 {
		return true
	}

	for after := range s.inst.events[event].after {
		if s.rats[after].Time <= s.rats[event].Time {
			return true
		}
	}

	for before := range s.inst.events[event].before {
		if s.rats[before].Time >= s.rats[event].Time {
			return true
		}
	}

	return false
}

// Determine the best assignment for the given event if one exists. In the
// even that one cannot be found, an unassigned Rat is returned.
func (s *Solution) Improve(event int) Rat {
	if event > s.inst.nEvents {
		panic("Solution.Improve : event > nEvents")
	}

	if s.HasViolations(event) {
		bestValue := s.Value()
		originalRat := s.rats[event]
		bestRat := originalRat

		for _, rat := range s.Domains[event] {
			s.Assign(event, rat)
			value := s.Value()

			if value.Less(bestValue) {
				bestValue = value
				bestRat = rat
			}
		}

		s.Assign(event, originalRat)

		// If we cannot find something better than the current assignment, we
		if bestRat != s.rats[event] {
			return bestRat
		}
	}

	return badRat
}

// Compute the distance to feasibility of a solution. The distance to
// to feasibility is defined as the sum of the number of students who attend
// unscheduled classes.
func (s *Solution) Distance() (dist int) {
	dist = 0

	for event, rat := range s.rats {
		if !rat.Assigned() {
			dist += len(s.inst.events[event].students)
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
				if len(s.attendance[student][day*9+hour]) > 0 {
					count++
					consecutive++

					if consecutive > 2 {
						fit++
					}
				} else {
					consecutive = 0
				}
			}

			if count == 1 {
				fit++
			}

			if len(s.attendance[student][day*9+8]) > 0 {
				fit++
			}
		}
	}

	return
}

// Free the solution back to the object pool.
func (s *Solution) Free() {
	for event := range s.rats {
		s.rats[event] = badRat
	}

	for ratIndex := range s.events {
		for event := range s.events[ratIndex] {
			delete(s.events[ratIndex], event)
		}
	}

	// Reset the attendance matrix
	for student := range s.attendance {
		for time := range s.attendance[student] {
			for event := range s.attendance[student][time] {
				delete(s.attendance[student][time], event)
			}
		}
	}

	s.inst.solnPool.Put(s)
}

// Determine the number of events involved in the solution (and problem
// instance).
func (s *Solution) NEvents() int {
	return s.inst.nEvents
}

// Get the Rat assigned to the index. If the eventIndex is invalid, badRat is
// returned.
func (s *Solution) RatAt(eventIndex int) Rat {
	if eventIndex > s.inst.nEvents {
		return badRat
	} else {
		return s.rats[eventIndex]
	}
}

// Determine the value of the solution (ie. the distance and fitness).
func (s *Solution) Value() Value {
	return Value{s.Violations(), s.Fitness()}
}

// Determine the number of hard constraint violations in the solution.
func (s *Solution) Violations() (violations int) {
	violations = 0

	// We consider the number of students that must attend a multiple events
	// at once. In this case, the penality is the number of events that each
	// student must attend more than one in each time slot.
	for student := range s.attendance {
		for time := range s.attendance[student] {
			if nEvents := len(s.attendance[student][time]); nEvents >= 2 {
				violations += nEvents - 1
			}
		}
	}

	// We consider the number of events assigned to each Rat. If there are
	// multiple events assigned to a single Rat, then the penalty is the
	// number of pairs of conflicting events.
	for ratIndex := range s.events {
		if nEvents := len(s.events[ratIndex]); nEvents >= 2 {
			// n choose 2 = 1 + 2 + ... + n-1 = n(n-1)/2, for n >=2
			violations += (nEvents * (nEvents - 1)) / 2
		}
	}

	// We consider the order of events. We only consider the `after' relation
	// as  A `after` B is equivalent to B `before` A. If there is an event
	// that is supposed to occur after another that is not scheduled as such,
	// the penality is 1 per such event.
	for eventIndex := range s.rats {
		event := &s.inst.events[eventIndex]

		for otherIndex := range event.after {
			if s.rats[otherIndex].Time <= s.rats[eventIndex].Time {
				violations++
			}
		}
	}

	// We do not have to check if events are scheduled in invalid timeslots or
	// rooms (e.g., that are too small or do not contain appropriate features)
	// as the domain generation at the beginning removes that possibility.
	return

}

// Write the solution to the given writer.
func (s *Solution) Write(w io.Writer) {
	for _, rat := range s.rats {
		fmt.Fprintf(w, "%d %d\n", rat.Time, rat.Room)
	}
}

// Unassign the given event.
func (s *Solution) Unassign(eventIndex int) {
	if eventIndex > s.inst.nEvents {
		panic("Solution.Unassign: eventIndex > nEvents")
	} else if s.rats[eventIndex].Assigned() {
		event := &s.inst.events[eventIndex]
		rat := s.rats[eventIndex]

		// Remove all entries from the attendance matrix.
		for student := range event.students {
			delete(s.attendance[student][rat.Time], eventIndex)
		}

		s.rats[eventIndex] = badRat
		delete(s.events[rat.index()], eventIndex)
	}
}
