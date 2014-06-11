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

package set

// Insert the value into the tree and return it.
func (s *Set) Insert(value interface{}) (inserted *Element, dup bool) {
	if inserted, dup = s.treeInsert(value); !dup {
		s.insert(inserted)
		s.size++
	}
	return
}

// Insert a value into the tree using the regular binary search tree method and
// return the pointer to the element and whether or not it is a duplicate (i.e.
// the value is already in the tree).
func (s *Set) treeInsert(value interface{}) (*Element, bool) {
	if s.root == nil {
		s.root = newElement(value, nil)
		return s.root, false
	} else {
		parent := s.root

		for {
			order := s.compare(value, parent.value)

			switch order {
			case Eq:
				return parent, true

			case Lt:
				if parent.left == nil {
					parent.left = newElement(value, parent)
					return parent.left, false
				}

				parent = parent.left

			case Gt:
				if parent.right == nil {
					parent.right = newElement(value, parent)
					return parent.right, false
				}

				parent = parent.right
			}
		}
	}
}

// Re-balance the red-black tree after element has been inserted.
func (s *Set) insert(element *Element) {
	u := element.uncle()
	g := element.grandparent()

	switch {
	// Case 1: the element is the root, which must be black.
	case element.parent == nil:
		element.red = false

	// Case 2: We insert a red child under a black parent, so the tree is valid.
	case !element.parent.red:
		return

	// Case 3: Both the parent and uncle are red, so we re-paint them black and
	// paint the grandparent black. We then have to call insert on g to fix the
	// invariants we have violated.
	case u != nil && u.red:
		element.parent.red = false
		u.red = false
		g.red = true

		s.insert(g)

	// Case 4: The parent is red and the uncle is black. We rotate the tree
	// appropriately and then fix the rotated sub-tree.
	default:
		if element.isRightChild() && element.parent.isLeftChild() {
			s.rotateLeft(element.parent)
			element = element.left
			g = element.grandparent()
		} else if element.isLeftChild() && element.parent.isRightChild() {
			s.rotateRight(element.parent)
			element = element.right
			g = element.grandparent()
		}

		element.parent.red = false
		g.red = true

		if element.isLeftChild() {
			s.rotateRight(g)
		} else {
			s.rotateLeft(g)
		}
	}
}
