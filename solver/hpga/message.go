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

package hpga

import "github.com/brennie/spaghetti/tt"

// The message type discriminator. We don't do a switch on the actual type
// of the interface ebcause the baseMessage can carry different types of
// messages.
type msgType int

const (
	stopMsg  msgType = iota // The message telling the children to stop.
	valueMsg                // A message containing a valuation.
	solnMsg                 // A solution message.
	finMsg                  // The message saying the child has finished.
)

// A message passed in the HPGA.
type message interface {
	Source() int      // Who sent the message
	MsgType() msgType // The type of the message
}

// The base message type that all message types should extend.
type baseMessage struct {
	source  int     // The source of the message.
	msgType msgType // The type of the message.
}

// A message containing a solution valuation.
type valueMessage struct {
	baseMessage
	value tt.Value // The value of the solution.
}

// A message carrying an actual solution. When sent to a child, this message
// carries the blank solution template. When sent to a parent, this contains
// an actual solution that is better than the current global one.
type solnMessage struct {
	baseMessage
	value tt.Value
	soln  tt.Solution
}

// Get the source of the message.
func (msg baseMessage) Source() int {
	return msg.source
}

// Get the type of the message.
func (msg baseMessage) MsgType() msgType {
	return msg.msgType
}
