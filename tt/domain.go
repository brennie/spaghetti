// hpgatt: Hierarchical Parallel Genetic Algorithm for Timetabling
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

package tt

// A domain is a set of rooms and times.
type Domain map[Rat]bool

// Create a copy of a domain.
func (domain Domain) Clone() Domain {
	clone := make(Domain)

	for key := range domain {
		clone[key] = domain[key]
	}

	return clone
}
