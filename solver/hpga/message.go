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
	"github.com/brennie/spaghetti/tt"
	"log"
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

func (m *message) messageType() messageType {
	return m.content.messageType()
}

func send(c chan<- message, s int, m messageContent) {
	c <- message{s, m}
}

func (p *parent) sendToChild(child int, m messageContent) {
	if child >= len(p.toChildren) {
		log.Fatalf("invalid child: %d", child)
	}

	send(p.toChildren[child], parentID, m)
}

func (c *child) sendToParent(m messageContent) {
	send(c.toParent, c.id, m)
}

// The content of a message is an arbitrary data structure that implements the
// messageType() method.
type messageContent interface {
	messageType() messageType
}

type crossoverRequestMessage struct{ soln []tt.Rat }

func (_ crossoverRequestMessage) messageType() messageType { return crossoverRequestMessageType }

type finMessage struct{}

func (_ finMessage) messageType() messageType { return finMessageType }

type gmRequestMessage struct{}

func (_ gmRequestMessage) messageType() messageType { return gmRequestMessageType }

type gmReplyMessage struct{ soln []tt.Rat }

func (_ gmReplyMessage) messageType() messageType { return gmReplyMessageType }

type solutionMessage struct {
	soln  []tt.Rat
	value tt.Value
}

func (_ solutionMessage) messageType() messageType { return solutionMessageType }

type solutionRequestMessage struct{ id int }

func (_ solutionRequestMessage) messageType() messageType { return solutionRequestMessageType }

type solutionReplyMessage struct {
	id   int
	soln []tt.Rat
}

func (_ solutionReplyMessage) messageType() messageType { return solutionReplyMessageType }

type stopMessage struct{}

func (_ stopMessage) messageType() messageType { return stopMessageType }

type valueMessage struct{ value tt.Value }

func (_ valueMessage) messageType() messageType { return valueMessageType }
