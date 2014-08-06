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
	"os"
	"os/signal"
	"time"

	"github.com/brennie/spaghetti/options"
	"github.com/brennie/spaghetti/tt"
)

// A controller is just a parent; its children are the islands.
type controller struct {
	parent
	inst     *tt.Instance // The timetabling instance
	topValue tt.Value     // The value of the top-valued solution.
	top      *tt.Solution // The top-valued solution
	ideal    bool         // Are we looking for an ideal solution?
}

// Create a new controller. There will be nIslands islands, each with nSlaves
// slaves.
func newController(inst *tt.Instance, opts options.SolveOptions) *controller {
	fromChildren := make(chan message, 5)

	c := &controller{
		parent{
			fromChildren,
			make([]chan<- message, opts.NIslands),
		},
		inst,
		tt.WorstValue(),
		inst.NewSolution(),
		opts.Ideal,
	}

	for i := 0; i < opts.NIslands; i++ {
		c.parent.toChildren[i] = newIsland(i, inst, fromChildren, opts)
	}

	return c
}

func (c *controller) handleSolutionMessage(msg message) (shouldExit bool) {
	soln := msg.content.(solutionMessage).soln
	value := msg.content.(solutionMessage).value

	if value.Less(c.topValue) {
		c.topValue = value
		c.top.Free()
		c.top = c.inst.SolutionFromRats(soln)

		log.Printf("Found new best solution: %s\n", c.topValue)

		if c.ideal && c.topValue.IsIdeal() {
			log.Println("Found ideal solution. Stopping...")
			c.stopChildren()

			return true
		} else if !c.ideal && c.topValue.IsValid() {
			log.Println("Found valid solution. Stopping...")
			c.stopChildren()

			return true
		}

		if msg.source != hcID {
			for child := range c.toChildren {
				if child != msg.source {
					c.sendToChild(child, valueMessage{c.topValue})
				}
			}
		}
	}

	return false
}

// Run the controller.
func (c *controller) run(timeoutInterval int) (*tt.Solution, tt.Value) {
	// Wait for islands to signal that their children have finished generating populations
	c.wait()

	log.Println("Population generation finished")

	var timeout <-chan time.Time
	if timeoutInterval == 0 {
		// We make a channel so that we never have to worry about receiving on it.
		timeout = make(chan time.Time)
	} else {
		timeout = time.After(time.Duration(timeoutInterval) * time.Minute)
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)

	hc := make(chan message)

	go runHillClimbing(c.inst, hc)

msgLoop:
	for {
		select {
		case msg := <-c.fromChildren:
			switch msg.messageType() {
			case solutionMessageType:
				if shouldExit := c.handleSolutionMessage(msg); shouldExit {
					break msgLoop
				}
			}

		case msg := <-hc:
			switch msg.messageType() {
			case solutionMessageType:
				if shouldExit := c.handleSolutionMessage(msg); shouldExit {
					break msgLoop
				}

			case orderingMessageType:
				msg.source = parentID
				for child := range c.toChildren {
					c.toChildren[child] <- msg
				}
			}

		case <-timeout:
			log.Println("Timeout: stopping...")
			c.stopChildren()
			break msgLoop

		case <-signals:
			log.Println("Caught interrupt")
			break msgLoop
		}
	}

	return c.top, c.topValue
}

// Send a stopMessageType message to all islands under the controller and wait for
// a finMessageType message from each of them. If a solutionMessageType message arrives,
// it will be processed as normal (i.e., updating c.top and c.topValue if
// better than the current solution).
func (c *controller) stopChildren() {
	for child := range c.toChildren {
		c.sendToChild(child, stopMessage{})
	}

	finished := make(map[int]bool)

	for len(finished) != len(c.toChildren) {
		msg := <-c.fromChildren

		switch msg.messageType() {
		case finMessageType:
			finished[msg.source] = true

		case solutionMessageType:
			value := msg.content.(solutionMessage).value

			if value.Less(c.topValue) {
				soln := msg.content.(solutionMessage).soln
				c.topValue = value
				c.top.Free()
				c.top = c.inst.SolutionFromRats(soln)
				log.Printf("Found new best solution: %s\n", c.topValue)
			}
		}
	}
}
