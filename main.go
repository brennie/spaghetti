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

	"github.com/brennie/spaghetti/checker"
	"github.com/brennie/spaghetti/fetcher"
	"github.com/brennie/spaghetti/options"
	"github.com/brennie/spaghetti/solver"
)

func main() {
	log.SetFlags(log.Ltime)

	opts := options.Parse()

	switch opts.Mode() {
	case options.CheckMode:
		checker.Check(opts.(options.CheckOptions))

	case options.FetchMode:
		fetcher.Fetch(opts.(options.FetchOptions))

	case options.SolveMode:
		solver.Solve(opts.(options.SolveOptions))
	}
}
