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

package population

import "sync"

// A success ratio, which describes how often the associated individual in a
// population is involved in successful reproductions (i.e., ones where the
// resultant individual is of a higher quality than the parent).
type Success struct {
	crossovers int // The number of crossovers.
	successes  int // The number of successful reproductions.
	mutex      sync.RWMutex
}

// Get the success ratio.
func (s *Success) Ratio() (ratio float64) {
	s.mutex.RLock()
	if s.successes == 0 {
		ratio = 0
	} else {
		ratio = float64(s.successes) / float64(s.crossovers)
	}
	s.mutex.RUnlock()

	return
}
