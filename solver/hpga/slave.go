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
	"fmt"
	"log"
	"math/rand"

	"github.com/brennie/spaghetti/solver/hpga/population"
	"github.com/brennie/spaghetti/tt"
)

// A slave is just a child of an island.
type slave struct {
	child
	island  int          // The island the slave belongs to
	inst    *tt.Instance // The timetabling instance.
	rng     *rand.Rand   // The random number generator for the slave.
	verbose bool         // Determines if events should be logged.
}

// Optionally log a message if the verbose flag is set.
func (slave *slave) log(format string, args ...interface{}) {
	if slave.verbose {
		msg := fmt.Sprintf(format, args...)
		log.Printf("slave(%d.%d): %s\n", slave.island, slave.id, msg)
	}
}

// Create a new slave with the given id. The given channel is the channel the
// island should use to communicate with the controller. The channel returned
// is the channel the controller should use to communicate with the island.
func newSlave(island int, id int, inst *tt.Instance, seed int64, toParent chan<- message, verbose bool) chan<- message {
	fromParent := make(chan message)
	slave := &slave{
		child{
			id,
			fromParent,
			toParent,
		},
		island,
		inst,
		rand.New(rand.NewSource(seed)),
		verbose,
	}

	go slave.run()

	return fromParent
}

// Run the slave.
func (slave *slave) run() {
	// Receive the topValue-valued solution message from the island.
	topValue := (<-slave.fromParent).(valueMessage).value

	pop := population.New(slave.rng, slave.inst)
	if best, value := pop.Best(); value.Less(topValue) {
		topValue = value
		slave.sendToParent(solnMsg, *best.Clone(), value)
		slave.log("found new best-valued solution: (%d,%d)", value.Distance, value.Fitness)
	}

	for {
		select {
		case msg := <-slave.fromParent:
			switch msg.MsgType() {

			// Update the globally known topValue value.
			case valueMsg:
				if value := msg.(valueMessage).value; value.Less(topValue) {
					topValue = value
					slave.log("received valueMsg: value was better")
				} else {
					slave.log("received valueMsg: value was worse ")
				}

			// StopValue the slave.
			case stopMsg:
				slave.log("received stopMsg: exiting")
				slave.fin()
				return
			}
		}
	}
}
