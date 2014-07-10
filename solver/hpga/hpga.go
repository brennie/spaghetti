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

	"github.com/brennie/spaghetti/options"
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
	c.sendToParent(finMsgType)
}

// Send a message from a parent to one of its children. See the chanSend
// for more details.
func (p *parent) sendToChild(child int, msgType msgType, args ...interface{}) {
	if child >= len(p.toChildren) {
		log.Fatalf("invalid child: %d", child)
	}

	chanSend(p.toChildren[child], parentID, msgType, args...)
}

// Send a message from the child to its parent. See the chanSend function for
// more details.
func (c *child) sendToParent(msgType msgType, args ...interface{}) {
	chanSend(c.toParent, c.id, msgType, args...)
}

// Run the HPGA.
func Run(inst *tt.Instance, opts options.SolveOptions) (*tt.Solution, tt.Value) {
	return newController(inst, opts).run(opts.Timeout)
}
