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

// Run hill-climbing optimzation to build a static variable ordering and
// generate weights for each variable's values. The higher the weight of a
// value, the better that value has been determined to be.
func runHillClimbing(inst *tt.Instance, report chan<- message) {
	varOrder := make([]int, inst.NEvents())
	valOrder := make([]tt.WeightedValues, inst.NEvents())
	valWeights := make([]map[tt.Rat]int, inst.NEvents())
	varWeights := make(tt.WeightedValues, inst.NEvents())
	varViolations := make([]int, inst.NEvents())
	nSolutions := 0

	for event := 0; event < inst.NEvents(); event++ {
		valOrder[event] = make(tt.WeightedValues, len(inst.Domains[event]))
		valWeights[event] = make(map[tt.Rat]int)

		for i, value := range inst.Domains[event] {
			valOrder[event][i].Value = value
			valOrder[event][i].Weight = 1
			valWeights[event][value] = 1
		}

		varWeights[event].Value = event
	}

	for global := 0; global < maxTries; {
		soln := heuristics.RandomAssignment(inst.NewSolution())
		found := false
		nSolutions++

		for local := 0; local < cutOff && global < maxTries; local++ {
			global++

			if event, rat := soln.FindImprovement(); event != -1 {
				soln.Assign(event, rat)
			} else {
				break // We've reached a local minimum
			}

			if violations := soln.Violations(); violations == 0 {
				found = true
				fitness := soln.Fitness()
				send(report, hcID, solutionMessage{soln.Assignments(), tt.Value{violations, fitness}})
				break
			}
		}

		if !found {
			pairs := soln.ConstraintPairs()

			for constraint, violations := range pairs {
				if violations > 0 {
					varViolations[constraint.EventA] += violations
					varViolations[constraint.EventB] += violations
				}
			}

			for event, count := range varViolations {
				if count == 0 {
					valWeights[event][soln.RatAt(event)]++
				} else {
					valWeights[event][soln.RatAt(event)]--
				}
				varWeights[event].Weight += count
				varViolations[event] = 0
			}
		}
		soln.Free()
	}

	sort.Sort(varWeights)
	for event := range valOrder {
		// We end up with a totally positive weight. The minimum weight is 1.
		for i := range valOrder[event] {
			rat := valOrder[event][i].Value.(tt.Rat)
			weight := valWeights[event][rat]
			valOrder[event][i].Weight += nSolutions + weight
		}
	}

	for i := range varWeights {
		varOrder[i] = varWeights[i].Value.(int)
	}

	send(report, hcID, orderingMessage{varOrder, valOrder})
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
			i.generated[individual].Soln = heuristics.OrderedWeightedAssignment(i.inst.NewSolution(), i.varOrdering, i.valOrdering)
			i.generated[individual].Value = i.generated[individual].Soln.Value()
		}
		report <- true
	}
}
