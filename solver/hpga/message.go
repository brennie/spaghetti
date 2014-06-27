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
	stopMsgType      msgType = iota // The message telling the children to stop.
	valueMsgType                    // A message containing a valuation.
	solnMsgType                     // A solution message.
	finMsgType                      // The message saying the child has finished.
	xoverReqMsgType                 // A slave requesting a crossover from an island.
	solnReqMsgType                  // An island requesting a solution from a slave.
	solnReplyMsgType                // A slave replying to an island for a crossover.
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

// Get the source of the message.
func (msg baseMessage) Source() int {
	return msg.source
}

// Get the type of the message.
func (msg baseMessage) MsgType() msgType {
	return msg.msgType
}

// A message containing a solution valuation.
type valueMessage struct {
	baseMessage
	value tt.Value // The value of the solution.
}

// A message from a slave requesting a crossover with another slave.
type xoverReqMessage struct {
	baseMessage
	soln []tt.Rat // The solution to crossover with.
}

// A message from a slave replying to an island for a crossover with another
// slave.
type solnReplyMessage struct {
	baseMessage
	id   int      // The crossover id.
	soln []tt.Rat // The solution to crossover with.
}

// A message carrying an actual solution. When sent to a child, this message
// carries the blank solution template. When sent to a parent, this contains
// an actual solution that is better than the current global one.
type solnMessage struct {
	baseMessage
	value tt.Value
	soln  []tt.Rat
}

// A message requesting a solution from a slave.
type solnReqMessage struct {
	baseMessage
	id int // The crossover id.
}

// Send a generic message along the channel.
//
//     Message Type   |      Arguments
// -------------------+-----------------------
//     valueMsgType   | tt.Value
//      solnMsgType   | tt.Value, []tt.Rat
//  solnReplyMsgType  | int, []tt.Rat
//   solnReqMsgType   | int
//  xoverReqMsgType   | []tt.Rat
func chanSend(c chan<- message, source int, msgType msgType, args ...interface{}) {
	base := baseMessage{source, msgType}

	switch msgType {
	case valueMsgType:
		c <- valueMessage{base, args[0].(tt.Value)}

	case solnMsgType:
		c <- solnMessage{base, args[0].(tt.Value), args[1].([]tt.Rat)}

	case solnReplyMsgType:
		c <- solnReplyMessage{base, args[0].(int), args[1].([]tt.Rat)}

	case solnReqMsgType:
		c <- solnReqMessage{base, args[0].(int)}

	case xoverReqMsgType:
		c <- xoverReqMessage{base, args[0].([]tt.Rat)}

	default:
		c <- base
	}
}
