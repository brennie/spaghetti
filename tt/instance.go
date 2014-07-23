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
	"sync"
)

// An instance of a timetabling problem.
type Instance struct {
	nEvents   int       // The number of events in the instance.
	nRooms    int       // The number of rooms in the instance.
	nFeatures int       // The number of features in the instance.
	nStudents int       // The number of students in the instance.
	rooms     []room    // The rooms in the instance.
	events    []event   // The events in the instance.
	solnPool  sync.Pool // A object pool for solutions.
	domains   [][]Rat   // The master copy of the domains..
}

// Allocate the memory for a solution.
func (inst *Instance) allocSolution() (s *Solution) {
	s = &Solution{
		inst,
		make([][45]map[int]bool, inst.nStudents),
		make([]map[int]bool, inst.nRooms*NTimes),
		make([]Rat, inst.nEvents),
		inst.domains,
	}

	for event := range s.rats {
		s.rats[event] = badRat
	}

	for index := range s.events {
		s.events[index] = make(map[int]bool)
	}

	for student := range s.attendance {
		for time := range s.attendance[student] {
			s.attendance[student][time] = make(map[int]bool)
		}
	}

	return
}

// Get the number of events in the instance.
func (inst *Instance) NEvents() int {
	return inst.nEvents
}

// Create a new empty solution to the instance.
func (inst *Instance) NewSolution() (s *Solution) {
	return inst.solnPool.Get().(*Solution)
}

// Create a solution with the assignments specified in rats.
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
