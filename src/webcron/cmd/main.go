package main

import "fmt"
import "github.com/robfig/cron"

func main() {
	c := cron.New()
	c.AddFunc("0 30 * * * *", func() { fmt.Println("Every hour on the half hour") })
	c.AddFunc("@hourly", func() { fmt.Println("Every hour") })
	c.AddFunc("@every 1h30m", func() { fmt.Println("Every hour thirty") })
	c.Start()

	// Funcs are invoked in their own goroutine, asynchronously.

	// Funcs may also be added to a running Cron
	c.AddFunc("@daily", func() { fmt.Println("Every day") })

	// Inspect the cron job entries' next and previous run times.
	for _, entry := range c.Entries() {
		fmt.Println(entry.Schedule)
	}

	// Stop the scheduler (does not stop any jobs already running).
	c.Stop()

	fmt.Println("Done!")
}
