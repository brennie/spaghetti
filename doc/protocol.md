#Communication Protocol

The communication protocol consists of three phases:
 
 1. Setup
 2. Main
 3. Shutdown

## 1. Setup Phase

In the set up phase, an upper bound is established for the solution by the controller and sent to all children.

### 1.1 Controller
The controller first sends a `waitMessage` to each island, each with the same `sync.WaitGroup` to wait for population generation. It also launches the hill climbing function to generate 

### 1.2 Islands
The islands send a `waitMessage` to each slave, each with the same `sync.WaitGroup`. Then they wait for the slaves to generate their populations. The islands wait for a `waitMessage` from the controller and calls `wg.Done()` on the given `sync.WaitGroup`.

### 1.3 Slaves
The slaves generate their populations. Then they wait for a `waitMessage` from their parent island and call `wg.Done()` on the given `sync.WaitGroup`.

## 2 Main Phase
After the setup, the controller, islands, and slaves transition into the main phase. In this phase, the island and controller's main purposes are message forwarding -- all work is except for crossovers and migrations are done by the slaves.

### 2.1 Controller
The controller is only a message relayer, but it chooses when the HPGA should enter the shutdown phase. (TODO: determine when that should be). The controller only does the following in a loop:

 1. Check for a message from the islands
  1. If the message is a `solutionMessage`, check if the value contained in the `solutionMessage` is better than the currently known one, update the value and solution, and send a `valueMessage` to all children. Otherwise disregard the message.
 2. Else check for a message from the hill climbing operator.
  1. If the message is a `solutionMessage`, check if the value contained in the `solutionmessage` is better than the currently known one, update the value and solution, and send a `valueMessage` to all children.
  2. Else if the message is an `orderingMessage`, forward the `orderingMessage` to all children.

### 2.2 Islands
Islands are mostly message relayers

 1. Check for a message from the parent.
  1. If there is a `stopMessage`, enter the shutdown phase.
  2. Else If there is a `valueMessage`, forward it to the slaves if the value is better than the currently known best.
  3. Else if there is an `orderingMessage`, set the ordering field and signal the GM to start producing individuals.
 2. If there is no message from the parent, check for a message from the children.
  1. If there is a `crossoverMessage`, select a child at random to send an empty `crossoverMessage`. Add the request to the queue of outstanding crossover requests.
  2. Else if there is an `crossoverMessage`, do the crossover with the first outstanding request and send an `solutionMessage` to the origin of the first `crossoverMessage`.
  3. Else if there is a `solutionMessage` from a child, determine if the value contained is better than then currently known value and forward it to the controller if so. Likewise, a `valueMessage` is sent to all children. Otherwise, ignore it.
  4. Else if there is a `fullMesage` from a child, check if all children's populations are full. If so, do a selection and notify the children they can continue via a `continueMessage`.

### 2.3 Slaves
First the slaves each generate a number of individuals (using the `RandomVariableOrdering()` method in the `solver/heuristics` package). Then it loops forever doing the following:

 1. Check for a message from the parent
  1. If there is a `stopMessage`, enter the shutdown phase.
  2. Else if there is a `valueMessage`, then update the current global best value.
  3. Else if there is a `crossoverMessage`, then respond with a `crossoverMessage` containing a population member.
  4. Else if there is a `solutionMessage`, add the individual to the population.
 2. If there is no message, generate a value $p$ in the interval $[0, 1]$.
  1. If $p < P_\mathrm{mutate}$, mutate a population member at random
  2. Else if $p < P_\mathrm{xover}$, do a local crossover between two population members at random.
  3. Else do a foreign crossover by sending a `crossoverMessage` to the island with a population member chosen at random.
  4. If a newly generated member has a better (distance, fitness) tuple than is currently known, update it and send a `solutionMessage` with a copy of the solution to the controlling island.
 3. If the population has reached its maximum size, send a `fullMessage` to its parent and wait for a `continueMessage`. Continue processing messages until it arrives.


## 3. Shutdown Phase
When it is time to shut down the system, the controller process will send out a `stopMessage` to all islands. The islands in turn send out a `stopMessage` to all of their slaves, which each reply with a `finMessage` and return. Once an island receives a `finMessage` from each of its children, it replies to the controller with a `finMessage` and returns. Finally, when the controller has received a `finMessage` from all of its islands, it returns the best solution. In this phase, `solutionMessage` is also handled appropriately; the islands will pass them along to the controller

