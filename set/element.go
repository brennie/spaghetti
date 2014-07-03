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

// An element in the set. This can be used as an iterator.
type Element struct {
	red         bool        // The colour of the node in the tree.
	value       interface{} // The value associated with the element.
	parent      *Element    // The parent element.
	left, right *Element    // The left and right children.
}

// Clean up the tree so that it may be garbage collected.
func (e *Element) free() {
	if e != nil {
		e.left.free()
		e.right.free()

		e.left = nil
		e.right = nil
		e.parent = nil
	}
}

// Get the value associated with the element.
func (e *Element) Value() interface{} {
	return e.value
}

// Find the succeeding element of `e'. If there is no succeeding element, then
// nil is returned instead. At worst, this is O(log n) as it will have to
// traverse either up or down the full height of the tree to find the successor.
func (e *Element) Next() *Element {
	if e.right != nil {
		// Find the left-most leaf in the right sub-tree
		return e.right.min()
	} else {
		// Find the ancestor that has this element in their left sub-tree
		for p := e.parent; p != nil; p, e = p.parent, p {
			if e.isLeftChild() {
				return p
			}
		}
	}

	return nil
}

// Find the preceeding element of `e'. If there is no preceeding element, then
// nil is returned instead. At worst, this is O(log n) as it will have to
// traverse either up or down the full height of the tree to find the
// predecessor.
func (e *Element) Prev() *Element {
	if e.left != nil {
		// Find the right-most leaf in the left sub-tree
		return e.left.max()
	} else {
		// Find the ancestor that has this element in their right sub-tree
		for p := e.parent; p != nil; p, e = p.parent, p {
			if e.isRightChild() {
				return p
			}
		}
	}

	return nil
}

// Create a new element to be inserted into a set.
func newElement(value interface{}, parent *Element) *Element {
	return &Element{
		true,
		value,
		parent,
		nil,
		nil,
	}
}

// Get the grandparent node or nil if there isn't one.
func (e *Element) grandparent() *Element {
	if e.parent != nil && e.parent.parent != nil {
		return e.parent.parent
	}

	return nil
}

// Get the sibling node of the parent or nil if there isn't one.
func (e *Element) uncle() *Element {
	if g := e.grandparent(); g != nil {
		if g.left == e.parent {
			return g.right
		} else {
			return g.left
		}
	}

	return nil
}

// Get the sibling node of the element or nil if there isn't one.
func (e *Element) sibling() *Element {
	if e.parent != nil {
		if e.parent.left == e {
			return e.parent.right
		}

		return e.parent.left
	}

	return nil
}

// Find the maximum element in the subtree under the given element.
func (e *Element) max() *Element {
	m := e
	for m.right != nil {
		m = m.right
	}
	return m
}

// Find the minimum element in the subtree under the given element.
func (e *Element) min() *Element {
	m := e
	for m.left != nil {
		m = m.left
	}
	return m
}

// Determine if an element is coloured red. By RB tree definitions, each nil
// child is actually a black node.
func red(e *Element) bool {
	if e != nil {
		return e.red
	}

	return false
}

// Determine if an element is the left child of its parent.
func (e *Element) isLeftChild() bool {
	if e.parent != nil {
		return e.parent.left == e
	}

	return false
}

// Determine if an element is the right child of its parent.
func (e *Element) isRightChild() bool {
	if e.parent != nil {
		return e.parent.right == e
	}

	return false
}
