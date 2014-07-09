spaghetti
=========

Applying Hierarchical Parallel Genetic Algorithms to solve the University Timetabling Problem.

Usage
=====

    Usage:
      spaghetti solve [options] <instance>
      spaghetti check <instance> <solution>
      spaghetti fetch [<directory>]
      spaghetti -h | --help
      spaghetti --version

    Options:  
      -h --help         Show this information.
      --islands <n>     Set the number of islands [default: 2].
      --minpop <n>      Set the minimum population size [default: 50].
      --maxpop <n>      Set the maximum population size [default: 75].
      --maxprocs <n>    Set GOMAXPROCS to the given value instead of the number of CPUs.
      --profile <file>  Collect profiling information in the given file. 
      --seed <seed>     Specify the seed for the random number generator.
      --slaves <n>      Set the number of slaves per island [default: 2].
      --timeout <n>     Set the timeout time in minutes [default: 30].
      --verbose         Turn on event logging.
      --version         Show version information.
      --output <file>   Write the solution to the given file instead of stdout.
