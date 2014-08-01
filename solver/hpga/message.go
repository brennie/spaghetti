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

import (
	"log"
	"sync"
	"time"

	"github.com/brennie/spaghetti/tt"
)

// The message type discriminator. We don't do a switch on the actual type
// of the interface ebcause the baseMessage can carry different types of
// messages.
type messageType int

const (
	parentID   = -1 // A parent's ID.
	hcID       = -2 // The HC operator's ID.
	newRequest = -1 // ID for a new crossover request.
)

const (
	continueMessageType  messageType = iota // A message telling a child to continue.
	crossoverMessageType                    // A message containing a crossover request from a slave.
	finMessageType                          // The message saying the child has finished.
	fullMessageType                         // The message saying the slave's population is full.
	orderingMessageType                     // A message containing a variable ordering.
	solutionMessageType                     // A message containing a solution.
	stopMessageType                         // The message telling the children to stop.
	valueMessageType                        // A message containing a valuation.
	waitMessageType                         // A message containing a sync.WaitGroup
)

// A message
type message struct {
	source  int            // The source of the message
	content messageContent // The content of the message
}

// Determine the message type via the content's type.
func (m *message) messageType() messageType {
	return m.content.messageType()
}

// Send a message on the given channel.
func send(c chan<- message, s int, m messageContent) {
	select {
	case c <- message{s, m}:
		break

	case <-time.After(10 * time.Second):
		panic("Could not send on channel after 10s -- deadlock?")

	}

}

// Send a message to the given child.
func (p *parent) sendToChild(child int, m messageContent) {
	if child >= len(p.toChildren) {
		log.Fatalf("invalid child: %d", child)
	}

	send(p.toChildren[child], parentID, m)
}

// Send a message to the child's parent.
func (c *child) sendToParent(m messageContent) {
	send(c.toParent, c.id, m)
}

// The content of a message is an arbitrary data structure that implements the
// messageType() method.
type messageContent interface {
	messageType() messageType
}

// A message telling a child to continue.
type continueMessage struct{}

// Get the messageType of a continueMessage.
func (_ continueMessage) messageType() messageType { return continueMessageType }

// A crossover request from a slave to an island. This is used for foreign crossovers.
type crossoverMessage struct {
	id int // The crossover's id (local to island). An id of -1 indicates a new request.
}

// Get the messageType of a crossoverRequestMessage.
func (_ crossoverMessage) messageType() messageType { return crossoverMessageType }

// A message indicating that a child has finished.
type finMessage struct{}

// Get the messageType of a finMessage.
func (_ finMessage) messageType() messageType { return finMessageType }

// A message indicating that a slave's population is full.
type fullMessage struct{}

// Get the messageType of a fullMessage.
func (_ fullMessage) messageType() messageType { return fullMessageType }

// A message containing a solution and its value.
type solutionMessage struct {
	soln  []tt.Rat // The solution as a list of assignments.
	value tt.Value // The value of the solution.
}

// Get the messageType of a solutionMessage.
func (_ solutionMessage) messageType() messageType { return solutionMessageType }

// A message containing a variable ordering.
type orderingMessage struct {
	varOrder []int               // The variable ordering
	valOrder []tt.WeightedValues // The value ordering.
}

// Get the messageType of an orderingMessage.
func (_ orderingMessage) messageType() messageType { return orderingMessageType }

// A message indicating that a child should stop.
type stopMessage struct{}

// Get the messageType of a stopMessage.
func (_ stopMessage) messageType() messageType { return stopMessageType }

// A message containing a solution's value. This is sent to inform of the
// best-known solution's value.
type valueMessage struct {
	value tt.Value // The best-known solution's value
}

// Get the messageType of a valueMessage.
func (_ valueMessage) messageType() messageType { return valueMessageType }

// A message containing a sync.WaitGroup
type waitMessage struct {
	wg *sync.WaitGroup // the Wait Group
}

func (_ waitMessage) messageType() messageType { return waitMessageType }
