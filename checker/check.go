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

// The checker package for evaluating solutions.
package checker

import (
	"fmt"
	"log"
	"os"

	"github.com/brennie/spaghetti/options"
	"github.com/brennie/spaghetti/tt"
)

func Check(opts options.CheckOptions) {
	instFile, err := os.Open(opts.Instance)
	if err != nil {
		log.Fatalf("Could not %s\n", err.Error())
	}
	defer instFile.Close()

	solnFile, err := os.Open(opts.Solution)
	if err != nil {
		log.Fatalf("Could not %s\n", err.Error())
	}
	defer solnFile.Close()

	inst, err := tt.Parse(instFile)
	if err != nil {
		log.Fatalf("Could not parse %s: %s\n", opts.Instance, err.Error())
	}

	soln, err := inst.ParseSolution(solnFile)
	if err != nil {
		log.Fatalf("Could not parse %s: %s\n", opts.Solution, err.Error())
	}

	fmt.Printf("Distance to feasibility: %d\nSoft Constraint Violations: %d\n",
		soln.Distance(), soln.Fitness())
}
