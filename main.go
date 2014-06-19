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
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"

	"github.com/docopt/docopt-go"

	"github.com/brennie/spaghetti/checker"
	"github.com/brennie/spaghetti/solver"
)

func main() {
	log.SetFlags(0)

	usage := `spaghetti: Applying Hierarchical Parallel Genetic Algorithms to solve the
University Timetabling Problem.

Usage:
  spaghetti solve [options] <instance>
  spaghetti check <instance> <solution>
  spaghetti -h | --help
  spaghetti --version

Options:
  -h --help         Show this information.
  --islands <n>     Set the number of islands [default: 3].
  --maxprocs <n>    Set GOMAXPROCS to the given value instead of the number of CPUs.
  --profile <file>  Collect profiling information in the given file. 
  --seed <seed>     Specify the seed for the random number generator.
  --slaves <n>      Set the number of slaves per island [default: 3].
  --version         Show version information.
  --output <file>   Write the solution to the given file instead of stdout.`

	arguments, err := docopt.Parse(usage, nil, true, "spaghetti v0.3", false)

	if err != nil {
		log.Fatalf("Could not parse arguments: %s\n", err.Error())
	}

	if arguments["solve"].(bool) {
		filename := arguments["<instance>"].(string)

		islands, err := strconv.Atoi(arguments["--islands"].(string))

		if err != nil {
			log.Fatalf("Invalid value for --islands: %s\n", err)
		}

		slaves, err := strconv.Atoi(arguments["--slaves"].(string))

		if err != nil {
			log.Fatalf("Invalid value for --slaves: %s\n", err)
		}

		if arguments["--maxprocs"] != nil {
			maxprocs, err := strconv.Atoi(arguments["--maxprocs"].(string))

			if err != nil {
				log.Fatalf("Invalid value for --maxprocs: %s\n", err)
			} else if maxprocs > runtime.NumCPU() {
				runtime.GOMAXPROCS(runtime.NumCPU())
			} else if maxprocs > 0 {
				runtime.GOMAXPROCS(maxprocs)
			}
		} else {
			runtime.GOMAXPROCS(runtime.NumCPU())
		}

		output := os.Stdout
		if arguments["--output"] != nil {
			output, err := os.Create(arguments["--output"].(string))

			if err != nil {
				log.Fatal(err)
			}

			defer output.Close()
		}

		if arguments["--profile"] != nil {
			profile, err := os.Create(arguments["--profile"].(string))

			if err != nil {
				log.Fatal(err)
			}

			pprof.StartCPUProfile(profile)
			defer profile.Close()
			defer pprof.StopCPUProfile()
		}

		if arguments["--seed"] != nil {
			seed, err := strconv.ParseInt(arguments["--seed"].(string), 10, 64)

			if err != nil {
				log.Fatal(err)
			}

			rand.Seed(seed)
		}

		solver.Solve(filename, output, islands, slaves)
	} else {
		instance := arguments["<instance>"].(string)
		solution := arguments["<solution>"].(string)

		checker.Check(instance, solution)
	}
}
