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

// Remove the specified value from the set.
func (s *Set) Remove(value interface{}) {
	if element := s.Find(value); element != nil {
		// If an element to remove has two children, we replace it with the
		// preceeding value and remove that element instead.
		if element.left != nil && element.right != nil {
			prev := element.Prev()
			element.value = prev.value
			element = prev
		}

		// An element with one non-leaf child can be replaced by its
		// non-leaf child; an element with no non-leaf children is just
		// removed.
		child := element.right
		if child == nil {
			child = element.left
		}

		// If we remove a black element, we have to fix the tree to maintain
		// the invariants.
		if !element.red {
			element.red = red(child)
			s.remove(element)
		}

		s.replace(element, child)

		// We have replaced the root so we should make sure that it is black.
		if element.parent == nil && child != nil {
			child.red = false
		}

		element.free()

		s.size--
	}
}

// Fix the invariants we have violated by removing an element.
func (s *Set) remove(element *Element) {
	// Case 1: We have removed the root node. No invariants are violated.
	if element.parent == nil {
		return
	}

	// Case 2: The sibling is red. We exchange the colours of the parent and
	// the sibling; we then rotate so that the sibling is now the parent of its
	// previous parent.
	if sibling := element.sibling(); red(sibling) {
		element.parent.red = true
		sibling.red = false

		if element.isLeftChild() {
			s.rotateLeft(element.parent)
		} else {
			s.rotateRight(element.parent)
		}
	}

	// Case 3: The element's parent, sibling, and sibling's children are all
	// black. We paint the sibling red and then fix from the parent.
	if sibling := element.sibling(); !red(element.parent) && !red(sibling) && !red(sibling.left) && !red(sibling.right) {
		sibling.red = true
		s.remove(element.parent)
	} else if sibling := element.sibling(); red(element.parent) && !red(sibling) && !red(sibling.left) && !red(sibling.right) {
		// Case 4: The element's sibling and its children are black, but it has
		// a red parent. We swap the colours of the parent and child
		sibling.red = true
		element.parent.red = false
	} else {
		// Case 5: The sibling is black, the sibling's left child is red (black),
		// the sibling's right child is black (red), and the element is the left
		// (right) child of its parent. We exchange the colours of S and its left
		// (right) child and rotate right (left) at the sibling.
		if sibling := element.sibling(); !red(sibling) {
			if element.isLeftChild() && red(sibling.left) && !red(sibling.right) {

				sibling.red = true
				sibling.left.red = false
				s.rotateRight(sibling)
			} else if element.isRightChild() && !red(sibling.left) && red(sibling.right) {

				sibling.red = true
				sibling.right.red = false
				s.rotateLeft(sibling)
			}
		}

		// Case 6: The sibling is black, the right child is red, and the
		// element is the left (right) child of its parent. We swap the colors
		// of the element's parent and its sibling, make the sibling's right (left)
		// child black, and then rotate left (right) at the parent.
		sibling := element.sibling()
		sibling.red = red(element.parent)
		element.parent.red = false

		if element.isLeftChild() {
			sibling.right.red = false
			s.rotateLeft(element.parent)
		} else {
			sibling.left.red = false
			s.rotateRight(element.parent)
		}
	}
}
