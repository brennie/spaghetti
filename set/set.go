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

// A generic set backed by a red-black tree.
package set

// An ordering in a comparison
type Order int

const (
	Lt Order = -1 // The first value is less than the second.
	Eq Order = 0  // The two values are equal.
	Gt Order = 1  // The first value is greater than the second.
)

// A function which compares elements and determines their order.
type Compare func(a, b interface{}) Order

// A set of values.
type Set struct {
	root    *Element // The root of the RB tree
	compare Compare  // The compare function
	size    int      // The number of elements in the RB tree
}

// Create a new set.
func New(compare Compare) Set {
	return Set{
		nil,
		compare,
		0,
	}
}

// Determine if the set contains the given value.
func (s *Set) Contains(value interface{}) bool {
	return s.Find(value) != nil
}

// Find the element corresponding to the given value in the set. If the element
// is not in the set, then nil is returned instead.
func (s *Set) Find(value interface{}) *Element {
	for element := s.root; element != nil; {
		switch s.compare(value, element.value) {
		case Eq:
			return element

		case Lt:
			element = element.left

		case Gt:
			element = element.right
		}
	}

	return nil
}

// Get the first element of the set (or nil if it is empty).
func (s *Set) First() (first *Element) {
	if s.root == nil {
		return nil
	}

	return s.root.min()
}

// Get the last element of the set (or nil if it is empty).
func (s *Set) Last() (last *Element) {
	if s.root == nil {
		return nil
	}

	return s.root.max()
}

// Get the number of elements in the set.
func (s *Set) Size() int {
	return s.size
}

// Replace oldEl with newEl in s.
func (s *Set) replace(oldEl, newEl *Element) {
	if oldEl.parent == nil {
		s.root = newEl
	} else {
		if oldEl.isLeftChild() {
			oldEl.parent.left = newEl
		} else {
			oldEl.parent.right = newEl
		}
	}

	if newEl != nil {
		newEl.parent = oldEl.parent
	}
}

// Perform a left rotation on the given element
func (s *Set) rotateLeft(element *Element) {
	oldRight := element.right

	s.replace(element, oldRight)

	element.right = oldRight.left
	if element.right != nil {
		element.right.parent = element
	}

	oldRight.left = element
	element.parent = oldRight
}

// Perform a right rotation on the given element
func (s *Set) rotateRight(element *Element) {
	oldLeft := element.left

	s.replace(element, oldLeft)

	element.left = oldLeft.right
	if element.left != nil {
		element.left.parent = element
	}

	oldLeft.right = element
	element.parent = oldLeft
}
