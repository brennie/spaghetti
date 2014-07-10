#Communication Protocol

The communication protocol consists of three phases:
 
 1. Setup
 2. Main
 3. Shutdown

## 1. Setup Phase

In the set up phase, an upper bound is established for the solution by the controller and sent to all children.

### 1.1 Controller
The controller runs forward search without backtracking using the most-constrained variable first heuristic and sends a `valueMsgType` to the islands.
    
The island waits for each of these messages and immediately typecasts them to the correct type; it does not use a switch on the result of the `message.MsgTypeType()` method because if it is the incorrect message type, then the controller has violated the protocol and a `panic()` will happen.

### 1.2 Islands
The islands waits for a `valueMsgType` from the controller and forward it to their respective slaves.

### 1.3 Slaves
The slaves wait for a `valueMsgType` from the controller.

## 2 Main Phase
After the setup, the controller, islands, and slaves transition into the main phase. In this phase, the island and controller's main purposes are message forwarding -- all work is except for crossovers and migrations are done by the slaves.

### 2.1 Controller
The controller is only a message relayer, but it chooses when the HPGA should enter the shutdown phase. (TODO: determine when that should be). The controller only does the following in a loop:

 1. Check for a message from the islands
  1. If the message is a `solnMsgType`, check if the value contained in the `solnMsgType` is better than the currently known one, update the value and solution and send a `valueMsgType` to all children. Otherwise disregard the message.



### 2.2 Islands
Islands are mostly message relayers

 1. Check for a message from the parent.
  1. If there is a `stopMsgType`, enter the shutdown phase.
  2. Else If there is a `valueMsgType`, forward it to the slaves if the value is better than the currently known best.
 2. If there is no message from the parent, check for a message from the children.
  1. If there is a `xoverReqMsgType`, select a child at random to send an empty `solnReqMsgType`. Add the request to the queue of outstanding crossover requests.
  2. Else if there is an `solnReplyMsgType`, do the crossover with the first outstanding request and send an `solnMsgType` to the origin of the first `xoverReqMsgType`.
  3. Else if there is a `solnMsgType` from a child, determine if the value contained is better than then currently known value and forward it to the controller if so. Likewise, a `valueMsgType` is sent to all children. Otherwise, ignore it.

### 2.3 Slaves
First the slaves each generate a number of individuals (using the `RandomVariableOrdering()` method in the `solver/heuristics` package). Then it loops forever doing the following:

 1. Check for a message from the parent
  1. If there is a `stopMsgType`, enter the shutdown phase.
  2. Else if there is a `valueMsgType`, then update the current global best value.
  3. Else if there is a `solnReqMsgType`, then respond with a `solnReplyMsgType` containing a population member.
  4. Else if there is a `solnMsgType`, add the individual to the population.
 2. If there is no message, generate a value $p$ in the interval $[0, 1]$.
  1. If $p < P_\mathrm{mutate}$, mutate a population member at random
  2. Else if $p < P_\mathrm{xover}$, do a local crossover between two population members at random.
  3. Else do a foreign crossover by sending a `xoverReqMsgType` to the island with a population member chosen at random.
  4. If a newly generated member has a better (distance, fitness) tuple than is currently known, update it and send a `solnMsgType` with a copy of the solution to the controlling island.
 3. If the population has reached its maximum size, do a selection for the minimum size.


## 3. Shutdown Phase
When it is time to shut down the system, the controller process will send out a `stopMsgType` to all islands. The islands in turn send out a `stopMsgType` to all of their slaves, which each reply with a `finMsgType` and return. Once an island receives a `finMsgType` from each of its children, it replies to the controller with a `finMsgType` and returns. Finally, when the controller has received a `finMsgType` from all of its islands, it returns the best solution. In this phase, `solnMsgType` is also handled appropriately; the islands will pass them along to the controller

