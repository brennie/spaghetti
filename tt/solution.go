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
	"log"

	"github.com/brennie/spaghetti/set"
)

// A solution to an instance.
type Solution struct {
	inst       *Instance  // The problem instance.
	attendance [][45]bool // Student attendance matrix.
	events     []int      // Map each room and time to an event.
	rats       []Rat      // Map each event to a room and time.
	Domains    []Domain   // The domains.
}

// Determine if the event has been assigned to.
func (s *Solution) Assigned(eventIndex int) bool {
	if eventIndex > s.inst.nEvents {
		return true
	} else {
		return s.rats[eventIndex].assigned()
	}
}

// Assign an event to a room and time.
func (s *Solution) Assign(eventIndex int, rat Rat) {
	if eventIndex > s.inst.nEvents {
		return
	}

	event := &s.inst.events[eventIndex]
	ratIndex := rat.index()

	// If there was an event previous scheduled in the new room and time, then
	// we unschedule it. We also mark the time slot as free in the student
	// attendance matrix for each student who was attending the old event.
	if oldEvent := s.events[ratIndex]; oldEvent != -1 {
		s.Unassign(oldEvent)
	}

	// If the event was previously assigned to a room and time (oldRat), we
	// unassign the old room and time and update the student attendance matrix
	// to reflect the change. Otherwise we just add update the student
	// attendance matrix.
	if oldRat := s.rats[eventIndex]; oldRat.assigned() {
		s.events[oldRat.index()] = -1
		for student := range event.students {
			s.attendance[student][rat.Time] = true
			s.attendance[student][oldRat.Time] = false
		}
		s.unshrink(eventIndex, oldRat)
	} else {
		for student := range event.students {
			s.attendance[student][rat.Time] = true
		}
	}

	s.rats[eventIndex] = rat
	s.events[ratIndex] = eventIndex

	s.shrink(eventIndex)
}

// Create a copy of the solution, keeping the same underlying pointer to the
// instance.
func (s *Solution) Clone() (clone *Solution) {
	clone = s.inst.NewSolution()

	for event, rat := range s.rats {
		if rat.assigned() {
			clone.Assign(event, rat)
		}
	}

	return
}

// Compute the distance to feasibility of a solution. The distance to
// to feasibility is defined as the sum of the number of students who attend
// unscheduled classes.
func (s *Solution) Distance() (dist int) {
	dist = 0

	for event, rat := range s.rats {
		if !rat.assigned() {
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

			if s.attendance[student][day*9+8] {
				fit++
			}
		}
	}

	return
}

// Generate the domain of one event.
func (s *Solution) makeDomain(eventIndex int) {
	event := &s.inst.events[eventIndex]
	domain := &s.Domains[eventIndex]

	*domain = Domain{
		set.New(ratCmp),
		make(map[Rat]map[int]bool),
	}

	// First we determine all the valid times.
	var times [NTimes]bool
	for time := 0; time < NTimes; time++ {
		times[time] = event.times[time]
	}

	// We consider all the events that must occur before this one and, for each
	// one that is assigned, we remove the times that do not occur after the
	// event's start time.
	for before := range event.before {
		rat := s.rats[before]

		if rat.assigned() {
			for time := 0; time <= rat.Time; time++ {
				times[time] = false
			}
		}
	}

	// We consider all the events that must occur after this one and, for each
	// one that is assigned, we remove the times that do not occur before the
	// event's end time.
	for after := range event.before {
		rat := s.rats[after]

		if rat.assigned() {
			for time := rat.Time; time < NTimes; time++ {
				times[time] = false
			}
		}
	}

	// Add to the domain the unassigned rooms and times subset of valid times.
	for room := range event.rooms {
		for time := 0; time < NTimes; time++ {
			if times[time] {
				rat := Rat{room, time}

				if s.events[rat.index()] == -1 {
					domain.Entries.Insert(rat)
				}
			}
		}
	}

	// Remove from the domain all rats that are the result of the event's
	// exclusion set (i.e. other events that share a student) that are
	// already scheduled.
	for exclude := range event.exclude {
		rat := s.rats[exclude]
		if rat.assigned() {
			for room := 0; room < s.inst.nRooms; room++ {
				domain.Entries.Remove(Rat{room, rat.Time})
			}
		}
	}

	for el := domain.Entries.First(); el != nil; el = el.Next() {
		rat := el.Value().(Rat)
		domain.conflicts[rat] = make(map[int]bool)
	}

	return
}

// Generate the full list of domains for each event.
func (s *Solution) makeDomains() {
	for event := range s.Domains {
		s.makeDomain(event)
	}

	return
}

// Assign and unassign without shrinking the domains.
// QuickAssign is meant to go as fast as possible. If invalid parameters are
// passed, it will cause spaghetti to exit. Only use this with correct domain
// entries as it does not fail gracefully.
func (s *Solution) QuickAssign(eventIndex int, rat Rat) (fitness int) {
	if eventIndex > s.inst.nEvents {
		log.Fatalf("Solution.QuickAssign: Invalid eventIndex `%d'\n", eventIndex)
	}

	event := &s.inst.events[eventIndex]
	ratIndex := rat.index()

	if s.events[ratIndex] != -1 || s.rats[eventIndex].assigned() {
		log.Fatal("Solution.QuickAssign: Assigned eventIndex or Rat")
	}

	s.rats[eventIndex] = rat
	s.events[ratIndex] = eventIndex

	for student := range event.students {
		s.attendance[student][rat.Time] = true
	}

	fitness = s.Fitness()

	for student := range event.students {
		s.attendance[student][rat.Time] = false
	}

	s.events[ratIndex] = -1
	s.rats[eventIndex] = badRat

	return
}

// Shrink the domains after an assignment.
func (s *Solution) shrink(eventIndex int) {
	event := &s.inst.events[eventIndex]
	rat := s.rats[eventIndex]

	// Remove the assignment from the domains of all events.
	for event := range s.Domains {
		if event == eventIndex {
			continue
		}

		s.Domains[event].addConflict(rat, eventIndex)
	}

	// Remove the time slot from all events that share a student.
	for exclude := range event.exclude {
		for room := 0; room < s.inst.nRooms; room++ {
			s.Domains[exclude].addConflict(Rat{room, rat.Time}, eventIndex)
		}
	}

	// Remove the domain entries from all events that must occur before it.
	for before := range event.before {
		for room := 0; room < s.inst.nRooms; room++ {
			for time := rat.Time; time < NTimes; time++ {
				s.Domains[before].addConflict(Rat{room, time}, eventIndex)
			}
		}
	}

	// Rmove the domain entries from all events that must occur after it.
	for after := range event.after {
		for room := 0; room < s.inst.nRooms; room++ {
			for time := 0; time <= rat.Time; time++ {
				s.Domains[after].addConflict(Rat{room, time}, eventIndex)
			}
		}
	}
}

// Determine the value of the solution (ie. the distance and fitness).
func (s *Solution) Value() Value {
	return Value{s.Distance(), s.Fitness()}
}

// Write the solution to the given writer.
func (s *Solution) Write(w io.Writer) {
	for _, rat := range s.rats {
		fmt.Fprintf(w, "%d %d\n", rat.Time, rat.Room)
	}
}

// Unassign the given event.
func (s *Solution) Unassign(eventIndex int) {
	if !s.Assigned(eventIndex) {
		return
	}

	event := &s.inst.events[eventIndex]
	rat := s.rats[eventIndex]

	// Remove all entries from the attendance matrix.
	for student := range event.students {
		s.attendance[student][rat.Time] = false
	}

	s.rats[eventIndex] = badRat
	s.events[rat.index()] = -1

	s.unshrink(eventIndex, rat)
}

// Unshrink the domains as the result of an unassignment.
func (s *Solution) unshrink(eventIndex int, rat Rat) {
	event := &s.inst.events[eventIndex]

	// Remove the conflict of the specific assignment for all other events.
	for event := range s.Domains {
		s.Domains[event].removeConflict(rat, eventIndex)
	}

	// Remove the conflict from all events which share a student.
	for exclude := range event.exclude {
		for room := 0; room < s.inst.nRooms; room++ {
			s.Domains[exclude].removeConflict(Rat{room, rat.Time}, eventIndex)
		}
	}

	// Remove the conflict from all events which must occur before this event.
	for before := range event.before {
		for room := 0; room < s.inst.nRooms; room++ {
			for time := rat.Time; time < NTimes; time++ {
				s.Domains[before].removeConflict(Rat{room, time}, eventIndex)
			}
		}
	}

	// Remove the conflict from all events which must occur after this event.
	for after := range event.after {
		for room := 0; room < s.inst.nRooms; room++ {
			for time := 0; time <= rat.Time; time++ {
				s.Domains[after].removeConflict(Rat{room, time}, eventIndex)
			}
		}
	}
}
