// hpgatt: Hierarchical Parallel Genetic Algorithm for Timetabling
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
)

const (
	boolError   = "expected either 0 or 1; got %d instead"
	formatError = "invalid format at line %d: %s"
)

// Read an integer as a bool, only accepting 1 (as true) or 0 (as false).
func readBool(r io.Reader) (b bool, err error) {
	var n int

	if _, err = fmt.Fscanf(r, "%d\n", &n); err != nil {
		return
	}

	switch n {
	case 0:
		b = false

	case 1:
		b = true

	default:
		err = fmt.Errorf(boolError, n)
	}

	return
}

// Read a timetabling instance from the reader.
func Parse(r io.Reader) (newInst *Instance, err error) {
	line := 1 // Line number for error reporting.
	inst := new(Instance)

	newInst = nil
	err = nil

	if _, err = fmt.Fscanf(r, "%d %d %d %d\n", &inst.nEvents, &inst.nRooms,
		&inst.nFeatures, &inst.nStudents); err != nil {

		err = fmt.Errorf(formatError, line, err.Error())
		return
	}

	line++

	inst.rooms = make([]room, inst.nRooms)
	inst.events = make([]event, inst.nEvents)
	inst.students = make([]map[int]bool, inst.nStudents)

	for event := range inst.events {
		inst.events[event].id = event
		inst.events[event].rooms = make(map[int]bool)
		inst.events[event].features = make(map[int]bool)
		inst.events[event].before = make(map[int]bool)
		inst.events[event].after = make(map[int]bool)
		inst.events[event].students = make(map[int]bool)
		inst.events[event].exclude = make(map[int]bool)
	}

	for student := range inst.students {
		inst.students[student] = make(map[int]bool)
	}

	for room := range inst.rooms {
		inst.rooms[room].features = make(map[int]bool)
	}

	// There is one line for the capacity of each room.
	for room := range inst.rooms {
		if _, err = fmt.Fscanf(r, "%d\n", &inst.rooms[room].capacity); err != nil {
			err = fmt.Errorf(formatError, line, err.Error())
			return
		}

		line++
	}

	// There is one line for each student and each event to determine if that
	// student attends the event.
	for student := range inst.students {
		for event := range inst.events {
			var attends bool

			if attends, err = readBool(r); err != nil {
				err = fmt.Errorf(formatError, line, err.Error())
				return
			}

			if attends {
				inst.students[student][event] = true
				inst.events[event].students[student] = true
			}

			line++
		}
	}

	// There is one line for each room and each feature to determine if the
	// room has the feature.
	for room := range inst.rooms {
		for feature := 0; feature < inst.nFeatures; feature++ {
			var val bool

			if val, err = readBool(r); err != nil {
				err = fmt.Errorf(formatError, line, err.Error())
				return
			}

			if val {
				inst.rooms[room].features[feature] = true
			}

			line++
		}
	}

	// There is one line for each event and each feature to determine if the
	// event requires the feature.
	for event := range inst.events {
		for feature := 0; feature < inst.nFeatures; feature++ {
			var val bool

			if val, err = readBool(r); err != nil {
				err = fmt.Errorf(formatError, line, err.Error())
				return
			}

			if val {
				inst.events[event].features[feature] = true
			}

			line++
		}
	}

	// There is one line for each event and time to determine if the event
	// can be scheduled at that time.
	for event := range inst.events {
		for time := 0; time < NTimes; time++ {
			if inst.events[event].times[time], err = readBool(r); err != nil {
				err = fmt.Errorf(formatError, line, err.Error())
				return
			}

			line++
		}
	}

	// There is one line for each event and event to determine if the first
	// event occurs before (1) or after (-1) the second event.
	for first := range inst.events {
		for second := range inst.events {
			var val int

			if _, err = fmt.Fscanf(r, "%d\n", &val); err != nil {
				err = fmt.Errorf(formatError, line, err.Error())
				return
			}

			switch val {
			case 1:
				inst.events[second].before[first] = true

			case 0:
				break

			case -1:
				inst.events[second].after[first] = true

			default:
				err = fmt.Errorf(formatError, line, "expected 1, 0, or -1")
			}

			line++
		}
	}

	// Process the room-event pairs to determine which inst.rooms can hold which
	// events.
	for event := range inst.events {
		for room := range inst.rooms {
			if inst.rooms[room].canHost(&inst.events[event]) {
				inst.events[event].rooms[room] = true
			}
		}
	}

	// Process the attends matrix to build exclusion lists (as two events that
	// share a student cannot occur at the same time).
	for event := range inst.events {
		for student := range inst.events[event].students {
			for other := range inst.students[student] {
				if event == other {
					continue
				}

				inst.events[event].exclude[other] = true
			}
		}
	}

	newInst = inst
	return
}

// Parse a solution from the given reader.
func (inst *Instance) ParseSolution(r io.Reader) (s *Solution, err error) {
	soln := inst.NewSolution()
	s = nil

	for event := range soln.rats {
		var rat Rat

		if _, err = fmt.Fscanf(r, "%d %d\n", &rat.Time, &rat.Room); err != nil {
			err = fmt.Errorf(formatError, event+1, err.Error())
			return
		}

		if rat.Assigned() {
			soln.Assign(event, rat)
		}
	}

	s = soln
	return
}
