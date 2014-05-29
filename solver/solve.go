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

// The instance solver
package solver

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/brennie/spaghetti/tt"
)

// Attempt to solve the instance located in file, with the optional seed for
// random number generator.
func Solve(filename string, seed ...int64) {
	if len(seed) > 0 {
		rand.Seed(seed[0])
	}

	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not %s\n", err.Error())
		os.Exit(1)
	}

	inst, err := tt.Parse(file)
	file.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse %s: %s\n", filename,
			err.Error())
		os.Exit(1)
	}

	soln := mcvfs(inst)

	soln.Write(os.Stdout)
}
