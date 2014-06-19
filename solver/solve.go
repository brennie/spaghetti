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
	"io"
	"log"
	"os"

	"github.com/brennie/spaghetti/solver/hpga"
	"github.com/brennie/spaghetti/tt"
)

// Attempt to solve the instance located in the given filename.
func Solve(filename string, output io.Writer, islands, slaves int) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Could not %s\n", err.Error())
	}

	inst, err := tt.Parse(file)
	file.Close()

	if err != nil {
		log.Fatalf("Could not parse %s: %s\n", filename, err.Error())
	}

	soln := hpga.Run(3, 3, inst)

	soln.Write(output)
}
