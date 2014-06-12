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

// The heirarchical parallel genetic algorithm package.
package hpga

import (
	"log"

	"github.com/brennie/spaghetti/tt"
)

// The id of a parent. The source of a message isn't important when a child
// receives it from the parent because only one go-routine can have a handle
// to that channel.
const parentID = -1

// A parent process, which has child processes that it communicates to through
// channels.
type parent struct {
	fromChildren <-chan message   // Receive channel from children
	toChildren   []chan<- message // Send channels for children
}

// A child process, which has a parent that it communicates to through
// channels. Since the parent only has one receive channel, each child must
// have an id so that we can tell the parent which child is talking.
type child struct {
	id         int            // The child's identifier
	fromParent <-chan message // Receive channel from parent
	toParent   chan<- message // Send channel to parent
}

// Send the fin message to the parent.
func (c *child) fin() {
	c.toParent <- baseMessage{c.id, finMsg}
}

// Send a generic message to a child. The message type (msgType) determine what
// elements of args are used.
//
// Message Type | Arguments
// ------------------------
//   valueMsg   | int, int
//    solnMsg   | tt.Solution
//    seedMsg   | int64
func (p *parent) send(child int, msgType msgType, args ...interface{}) {
	if child >= len(p.toChildren) {
		log.Fatalf("invalid child: %d", child)
	}

	base := baseMessage{parentID, msgType}

	switch msgType {
	case valueMsg:
		p.toChildren[child] <- valueMessage{base, args[0].(int), args[1].(int)}

	case solnMsg:
		p.toChildren[child] <- solnMessage{base, args[0].(tt.Solution)}

	case seedMsg:
		p.toChildren[child] <- seedMessage{base, args[0].(int64)}

	default:
		p.toChildren[child] <- base
	}
}

// Send the stop message to all children and wait for all fo them to reply with
// a fin message.
func (p *parent) stop() {
	for child := range p.toChildren {
		p.send(child, stopMsg)
	}

	finished := make(map[int]bool)

	for len(finished) != len(p.toChildren) {
		msg := <-p.fromChildren

		if msg.MsgType() == finMsg {
			finished[msg.Source()] = true
		}
	}
}

// Run the hpga.
func Run(nIslands, nSlaves int) {
	newController(nIslands, nSlaves).run()
}
