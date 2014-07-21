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

	"github.com/brennie/spaghetti/tt"
)

// The message type discriminator. We don't do a switch on the actual type
// of the interface ebcause the baseMessage can carry different types of
// messages.
type messageType int

const (
	crossoverRequestMessageType messageType = iota // A message containing a crossover request from a slave.
	finMessageType                                 // The message saying the child has finished.
	gmRequestMessageType                           // A request for a solution to perform the GM operator.
	gmReplyMessageType                             // A reply to a gmRequestMessageType message.
	solutionMessageType                            // A message containing a solution.
	solutionRequestMessageType                     // An island requesting a solution from a slave.
	solutionReplyMessageType                       // A slave replying to an island for a crossover.
	stopMessageType                                // The message telling the children to stop.
	valueMessageType                               // A message containing a valuation.
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
	c <- message{s, m}
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

// A crossover request from a slave to an island. This is used for foreign crossovers.
type crossoverRequestMessage struct {
	soln []tt.Rat // The solution to use in the crossover.
}

// Get the messageType of a crossoverRequestMessage.
func (_ crossoverRequestMessage) messageType() messageType { return crossoverRequestMessageType }

// A message indicating that a child has finished.
type finMessage struct{}

// Get the messageType of a finMessage.
func (_ finMessage) messageType() messageType { return finMessageType }

// A request for a solution for genetic modification.
type gmRequestMessage struct{}

// Get the messageType of a gmRequestMessage.
func (_ gmRequestMessage) messageType() messageType { return gmRequestMessageType }

// A response to a gmRequestMessage with a solution.
type gmReplyMessage struct {
	soln []tt.Rat // The solution to modify
}

// Get the messageType of a gmReplyMessage.
func (_ gmReplyMessage) messageType() messageType { return gmReplyMessageType }

// A message containing a solution and its value.
type solutionMessage struct {
	soln  []tt.Rat // The solution as a list of assignments.
	value tt.Value // The value of the solution.
}

// Get the messageType of a solutionMessage.
func (_ solutionMessage) messageType() messageType { return solutionMessageType }

// A solution request from an island to a slave with a given identifier.
type solutionRequestMessage struct {
	id int // The request identifier.
}

// Get the messageType of a solutionRequestMessage.
func (_ solutionRequestMessage) messageType() messageType { return solutionRequestMessageType }

// A reply to a solutionRequestMessage.
type solutionReplyMessage struct {
	id   int      // The id from the solutionRequestMessage.
	soln []tt.Rat // The solution.
}

// Get the messageType of a solutionReplyMessage.
func (_ solutionReplyMessage) messageType() messageType { return solutionReplyMessageType }

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
