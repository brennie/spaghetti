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

// A domain is a set of rooms and times. The keys of the conflicts field are
// valid intial entries of the domains. As assigments happen with
// Solution.Assign, conflict entries will be added (the key to the second map
// is the conflicting event). When len(conflicts[K]) == 0, then K will be a key
// in Entries.
type Domain struct {
	Entries   map[Rat]bool         // The actual domain entries.
	conflicts map[Rat]map[int]bool // A set of conflicting assignments.
}

// Determine if the given rat is in the base domain.
func (d *Domain) inBaseDomain(rat Rat) bool {
	_, hasKey := d.conflicts[rat]
	return hasKey
}

// Determine if there is a conflict with the given Rat.
func (d *Domain) hasConflict(rat Rat) bool {
	if d.inBaseDomain(rat) {
		return len(d.conflicts[rat]) > 0
	} else {
		return false
	}
}
