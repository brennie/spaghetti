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

package solver

import (
	"container/heap"

	"github.com/brennie/spaghetti/pqueue"
	"github.com/brennie/spaghetti/tt"
)

// Do most constrained variable first search and assign what we can.
func mcvfs(inst *tt.Instance) (soln *tt.Solution) {
	soln = inst.NewSolution()

	pq := pqueue.New(soln.Domains)

	for pq.Len() > 0 {
		mc := heap.Pop(pq).(int)

		if soln.Domains[mc].Entries.Size() == 0 {
			continue
		}

		// Our iterator through the domain entries.
		el := soln.Domains[mc].Entries.First()

		// Set the base line to be the assignment from the first entry.
		minRat := el.Value().(tt.Rat)
		soln.Assign(mc, minRat)
		minFit := soln.Fitness()

		// Now we find the actual minimum.
		for el = el.Next(); el != nil; el = el.Next() {
			rat := el.Value().(tt.Rat)
			soln.Assign(mc, rat)
			fit := soln.Fitness()

			if fit < minFit {
				minFit = fit
				minRat = rat
			}
		}

		soln.Assign(mc, minRat)
		pq.Update()
	}

	return
}
