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

package options

import (
	"crypto/rand"
	"log"
	"math"
	"math/big"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docopt/docopt-go"
)

type Mode int

const (
	CheckMode Mode = iota
	FetchMode
	SolveMode
)

const (
	usage = `spaghetti: Applying Hierarchical Parallel Genetic Algorithms to solve the
University Timetabling Problem.

Usage:
  spaghetti solve [options] <instance>
  spaghetti check <instance> <solution>
  spaghetti fetch [<directory>]
  spaghetti -h | --help
  spaghetti --version

Options:  
  -h --help         Show this information.
  --ideal           Spaghetti will stop when it detects an ideal solution --
                    not a valid one. Specifying --ideal with --timeout 0 may
                    cause the program to never terminate.
  --islands <n>     Set the number of islands [default: 2].
  --minpop <n>      Set the minimum population size [default: 50].
  --maxpop <n>      Set the maximum population size [default: 75].
  --maxprocs <n>    Set GOMAXPROCS to the given value instead of the number of
                    CPUs.
  --profile <file>  Collect profiling information in the given file. 
  --seed <seed>     Specify the seed for the random number generator.
  --slaves <n>      Set the number of slaves per island [default: 2].
  --timeout <n>     Set the timeout time in minutes [default: 30]. A timeout of
                    0 means that spaghetti won't stop until it finds a valid
                    solution.
  --version         Show version information.
  --output <file>   Write the solution to the given file instead of stdout.`

	version = "spaghetti v0.12"
)

type Options interface {
	Mode() Mode
}

// Commandline options for the check Mode
type CheckOptions struct {
	Instance string // The instance to check against.
	Solution string // The solution to check.
}

func (o CheckOptions) Mode() Mode {
	return CheckMode
}

// Commandline options for the fetch Mode
type FetchOptions struct {
	Directory string // The directory to store the instances in.
}

func (o FetchOptions) Mode() Mode {
	return FetchMode
}

// Commandline options for the solve Mode
type SolveOptions struct {
	Instance string      // The instance we are solving or checking.
	Solution string      // The filename for the resultant solution.
	NIslands int         // The number of islands.
	NSlaves  int         // The number of slaves per island.
	MinPop   int         // The minimum population of each island.
	MaxPop   int         // The maximum population of each island.
	Profile  interface{} // Either a string or nil. Determines if profiling should be enabled.
	Seed     int64       // The seed for the random number generator.
	Timeout  int         // The timeout in minutes.
	Ideal    bool        // Should we stop when we find an ideal solution (true) or merely a valid one (false).
}

func (o SolveOptions) Mode() Mode {
	return SolveMode
}

func Parse() Options {
	args, err := docopt.Parse(usage, nil, true, version, false)

	if err != nil {
		log.Fatalf("Could not parse arguments: %s\n", err)
	}

	switch {
	case args["check"].(bool):
		return parseCheckOptions(args)

	case args["fetch"].(bool):
		return parseFetchOptions(args)

	default:
		return parseSolveOptions(args)
	}
}

func parseCheckOptions(args map[string]interface{}) (opts CheckOptions) {
	opts.Instance = args["<instance>"].(string)
	opts.Solution = args["<solution>"].(string)

	return
}

func parseFetchOptions(args map[string]interface{}) (opts FetchOptions) {
	if directory := args["<directory>"]; directory != nil {
		opts.Directory = directory.(string)
	} else {
		opts.Directory = "instances"
	}

	return
}

func parseSolveOptions(args map[string]interface{}) (opts SolveOptions) {
	var err error

	opts.Instance = args["<instance>"].(string)

	if solution := args["--output"]; solution != nil {
		opts.Solution = solution.(string)
	} else {
		if ext := filepath.Ext(opts.Instance); ext == ".tim" {
			opts.Solution = strings.TrimSuffix(opts.Instance, ".tim")
		}

		opts.Solution += ".sln"
	}

	opts.NIslands, err = strconv.Atoi(args["--islands"].(string))
	if err != nil {
		log.Fatalf("Invalid value for --islands: %s\n", args["--islands"].(string))
	} else if opts.NIslands < 2 {
		log.Fatalf("Invalid value for --islands (%d): value must be at least 2", opts.NIslands)
	}

	opts.NSlaves, err = strconv.Atoi(args["--slaves"].(string))
	if err != nil {
		log.Fatalf("Invalid value for --slaves: %s\n", args["--slaves"].(string))
	} else if opts.NSlaves < 2 {
		log.Fatalf("Invalid value for --slaves (%d): value must be at least 2", opts.NSlaves)
	}

	opts.MinPop, err = strconv.Atoi(args["--minpop"].(string))
	if err != nil {
		log.Fatalf("Invalid value for --minpop: %s\n", args["--minpop"].(string))
	}

	opts.MaxPop, err = strconv.Atoi(args["--maxpop"].(string))
	if err != nil {
		log.Fatalf("Invalid value for --maxpop: %s\n", args["--maxpop"].(string))
	} else if opts.MaxPop <= opts.MinPop {
		log.Fatalf("Value for --maxpop (%d) must exceed value for --minpop (%d)\n", opts.MaxPop, opts.MinPop)
	}

	opts.Timeout, err = strconv.Atoi(args["--timeout"].(string))
	if err != nil {
		log.Fatalf("Invalid value for --timeout: %s\n", args["--timeout"].(string))
	}

	opts.Ideal = args["--ideal"].(bool)

	if profileName := args["--profile"]; profileName != nil {
		opts.Profile = profileName
	} else {
		opts.Profile = nil
	}

	if seed := args["--seed"]; seed != nil {
		opts.Seed, err = strconv.ParseInt(args["--seed"].(string), 10, 64)

		if err != nil {
			log.Fatalf("invalid value for --seed: %s\n", args["--seed"].(string))
		}

	} else {
		seed, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			log.Fatalf("Could not read from system random number generator: %s\n", err.Error())
		}

		opts.Seed = seed.Int64()
	}

	return
}
