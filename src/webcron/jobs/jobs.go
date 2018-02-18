package jobs

import (
	"fmt"

	"github.com/google/uuid"
	"gopkg.in/robfig/cron.v2"
)

type HookPayload struct {
}

type Job struct {
	JobId    uuid.UUID
	Schedule string
	Payload  HookPayload
}

type JobManager struct {
	entryMap map[string]int
}

func NewJobManager() *JobManager {
	manager := &JobManager{
		make(map[string]int),
	}

	// TODO: perform initialization here
	c := cron.New()
	c.AddFunc("@every 10s", func() { fmt.Println("..") })
	// c.AddFunc("0 30 * * * *", func() { fmt.Println("Every hour on the half hour") })
	// c.AddFunc("@hourly", func() { fmt.Println("Every hour") })
	// c.AddFunc("@every 1h30m", func() { fmt.Println("Every hour thirty") })

	c.Start()

	return manager
}

// func (m *JobManager) GetJobs() []Job {

// }
// func (m *JobManager) CreateJob(job Job) {

// }
