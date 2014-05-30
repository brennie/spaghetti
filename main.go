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

package main

import (
	"log"

	"github.com/docopt/docopt-go"

	"github.com/brennie/spaghetti/checker"
	"github.com/brennie/spaghetti/solver"
)

func main() {
	log.SetFlags(0)

	usage := `spaghetti: Applying Hierarchical Parallel Genetic Algorithms to solve the
University Timetabling Problem.

Usage:
  spaghetti solve [--seed=<seed>] <instance>
  spaghetti check <instance> <solution>
  spaghetti -h | --help
  spaghetti --version

Options:
  -h --help       Show this screen.
  --version       Show version information.
  --seed=<seed>   Specify the seed for the random number generator.`

	arguments, err := docopt.Parse(usage, nil, true, "spaghetti v0.2", false)

	if err != nil {
		log.Fatalf("Could not parse arguments: %s\n", err.Error())
	}

	if arguments["solve"].(bool) {
		filename := arguments["<instance>"].(string)
		seed := arguments["<seed>"]

		if seed == nil {
			solver.Solve(filename)
		} else {
			solver.Solve(filename, seed.(int64))
		}
	} else {
		instance := arguments["<instance>"].(string)
		solution := arguments["<solution>"].(string)

		checker.Check(instance, solution)
	}
}
