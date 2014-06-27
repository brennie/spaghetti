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

import "github.com/brennie/spaghetti/set"

// An instance of a timetabling problem.
type Instance struct {
	nEvents   int            // The number of events in the instance.
	nRooms    int            // The number of rooms in the instance.
	nFeatures int            // The number of features in the instance.
	nStudents int            // The number of students in the instance.
	rooms     []room         // The rooms in the instance.
	events    []event        // The events in the instance.
	students  []map[int]bool // The attendance of the students in the instance.
}

// Get the number of events in the instance.
func (inst *Instance) NEvents() int {
	return inst.nEvents
}

// Create a new empty solution to the instance.
func (inst *Instance) NewSolution() (s *Solution) {
	s = &Solution{
		inst,
		make([][45]bool, inst.nStudents),
		make([]int, inst.nEvents*NTimes),
		make([]Rat, inst.nEvents),
		make([]Domain, inst.nEvents),
	}

	for event := range s.rats {
		s.rats[event] = badRat
	}

	for index := range s.events {
		s.events[index] = -1
	}

	for eventIndex := range s.Domains {
		domain := &s.Domains[eventIndex]
		event := &inst.events[eventIndex]

		*domain = Domain{
			set.New(ratCmp),
			make(map[Rat]map[int]bool),
		}

		for room := range event.rooms {
			for time, ok := range event.times {
				if ok {
					rat := Rat{room, time}
					domain.Entries.Insert(rat)
					domain.conflicts[rat] = make(map[int]bool)
				}
			}
		}
	}

	return
}

func (inst *Instance) SolutionFromRats(rats []Rat) (s *Solution) {
	if len(rats) != inst.nEvents {
		panic("len(rats) != inst.nEvents")
	}

	s = inst.NewSolution()
	for event, rat := range rats {
		if rat.Assigned() {
			s.Assign(event, rat)
		}
	}

	return s
}
