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

package hpga

import (
	"sort"

	"github.com/brennie/spaghetti/solver/heuristics"
	"github.com/brennie/spaghetti/tt"
)

const (
	maxTries = 1000 // The maximum number of iterations
	cutOff   = 50   // The cut off for a specific solution
)

// An event-weight pair.
type pair struct {
	event  int // The event
	weight int // It's weight
}

// A slice of event-weight pairs.
type pairs []pair

// Determine the length of a pairs.
func (p *pairs) Len() int {
	return len(*p)
}

// Determine if the weight at p[i] is less than the weight at p[j].
func (p pairs) Less(i, j int) bool {
	return p[i].weight < p[j].weight
}

// Swap p[i] and p[j]
func (p *pairs) Swap(i, j int) {
	(*p)[i], (*p)[j] = (*p)[j], (*p)[i]
}

// Run hill-climbing optimzation to build a static variable ordering heuristic for the HPGA.
func runHillClimbing(inst *tt.Instance, report chan<- message) {
	weights := make(pairs, inst.NEvents())
	order := make([]int, inst.NEvents())

	for event := range weights {
		weights[event].event = event
	}

	for globalCounter := 0; globalCounter < maxTries; {
		s := heuristics.RandomAssignment(inst.NewSolution())
		found := false

		for localCounter := 0; localCounter < cutOff && globalCounter < maxTries; {
			localCounter++
			globalCounter++

			event, rat := s.FindImprovement()

			if event != -1 {
				s.Assign(event, rat)
			} else {
				break // We are at a local minimum.
			}

			if violations := s.Violations(); violations == 0 {
				send(report, hcID, solutionMessage{s.Assignments(), tt.Value{violations, s.Fitness()}})
				found = true
				break
			}
		}

		if !found {
			for event := range weights {
				weights[event].weight += s.AssignmentViolations(event)
			}
		}
	}

	sort.Sort(&weights)

	for i := range weights {
		order[i] = weights[i].event
	}

	send(report, hcID, orderingMessage{order})
}

// Run the genetic modification operator for the island. The GM operator will
// wait for requests to generate count individuals to be sent on the report
// channel. The same slice will be used to send every report so it should not
// be modified by the island.
func (i *island) runGM(count int, requests <-chan bool, report chan<- bool) {
	for {
		if <-requests == false {
			break
		}
		for individual := range i.generated {
			i.generated[individual].Soln = heuristics.RandomAssignmentWithOrdering(i.inst.NewSolution(), i.ordering)
			i.generated[individual].Value = i.generated[individual].Soln.Value()
		}
		report <- true
	}
}
