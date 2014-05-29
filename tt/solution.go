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
	inst       *Instance  //The problem instance
	attendance [][45]bool // Student attendance matrix
	events     []int      // Map each room and time to an event.
	rats       []Rat      // Map each event to a room and time.
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
func (s *Solution) domain(eventIndex int) (domain Domain) {
	event := &s.inst.events[eventIndex]

	domain = make(Domain)

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
					domain[rat] = true
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
				// NB: delete(m, k) is a no-op if k is not in m's keys.
				delete(domain, Rat{room, rat.Time})
			}
		}
	}

	return
}

// Generate the full list of domains for each event.
func (s *Solution) Domains() (domains []Domain) {
	domains = make([]Domain, s.inst.nEvents)

	for event := range domains {
		domains[event] = s.domain(event)
	}

	return
}

func (s *Solution) Shrink(eventIndex int, domains []Domain) {
	if eventIndex > s.inst.nEvents || !s.rats[eventIndex].assigned() {
		return
	}

	domains[eventIndex] = nil
	event := &s.inst.events[eventIndex]
	rat := s.rats[eventIndex]

	// Remove the assigned room and time from all events
	for _, domain := range domains {
		delete(domain, rat)
	}

	// For each event that shares a student, remove all rooms and times with
	// the same time as the recently assigned event.
	for exclude := range event.exclude {
		if domains[exclude] != nil {
			for room := 0; room < s.inst.nRooms; room++ {
				delete(domains[exclude], Rat{room, rat.Time})
			}
		}
	}

	// For each event that occurs before the recently assigned event, remove
	// rooms and times that have a time that occurs during or after the event.
	for before := range event.before {
		if domains[before] != nil {
			for room := 0; room < s.inst.nRooms; room++ {
				for time := rat.Time; time < NTimes; time++ {
					delete(domains[before], Rat{room, time})
				}
			}
		}
	}

	for after := range event.after {
		if domains[after] != nil {
			for room := 0; room < s.inst.nRooms; room++ {
				for time := 0; time <= rat.Time; time++ {
					delete(domains[after], Rat{room, time})
				}
			}
		}
	}
}
