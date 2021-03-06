# Workout

Workout is a work processing library for [Go](http://golang.org) built on top of [beanstalkd](http://kr.github.io/beanstalkd). It provides a simple API, allowing you to get started quickly with reasonable default settings. Under the hood, Workout uses goroutines and a configurable number of workers to coordinate concurrent processing.

## Example

Here's an entire working program. It creates 3 tubes and inserts 1000 jobs into each. It then spawns 20 concurrent workers to plow through the work, reporting the results to stdout. Give it a try!

```go
package main

import (
  "fmt"
  "github.com/jcoene/workout"
  "os"
  "os/signal"
  "time"
)

func main() {
  // These are the tubes we'll be using
  tubes := []string{"default", "person", "address"}

  // Create a new workout client for job insertion
  wc, err := workout.NewClient("localhost:11300", tubes)
  if err != nil {
    fmt.Printf("unable to connect to beanstalkd: %s\n", err)
    os.Exit(1)
  }

  // Insert 1000 jobs for each tube
  for i := 0; i < 1000; i++ {
    for _, t := range tubes {
      job := &workout.Job{
        Tube:      t,
        Priority:  1,
        TimeToRun: (60 * time.Second),
        Body:      fmt.Sprintf("%d", i),
      }
      wc.Put(job)
    }
  }

  // Setup a workout master with 20 workers
  wm := workout.NewMaster("localhost:11300", 20)

  // Assign a job handler, callback handler and duration (after which the handler is abandoned and we return an error) for each job.
  for _, t := range tubes {
    wm.RegisterHandler(t, jobHandler, jobCallback, 60*time.Second)
  }

  // Start processing!
  wm.Start()

  // Tell workout to stop on CTRL+C
  go func() {
    ch := make(chan os.Signal)
    signal.Notify(ch, os.Interrupt)
    <-ch
    wm.Stop()
  }()

  // Block until we're finished
  wm.Wait()

  return
}

// A handler function takes a job pointer and returns an error (or nil on success)
func jobHandler(job *workout.Job) (err error) {
  // we don't actually have any work to do
  return nil
}

// A callback function takes a job, error and duration - useful for reporting
func jobCallback(job *workout.Job, err error, dur time.Duration) {
  if err != nil {
    fmt.Printf("job %d encountered an error: %s (took %v)\n", job.Id, err, dur)
  } else {
    fmt.Printf("job %d succeeded (took %v)\n", job.Id, dur)
  }
  return
}
```

## License

MIT license, see [LICENSE](https://github.com/jcoene/workout/blob/master/LICENSE) for details.


## Author

Jason Coene, [@jcoene](https://twitter.com/jcoene)
