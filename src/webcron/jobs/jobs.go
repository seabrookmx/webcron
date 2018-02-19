package jobs

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gopkg.in/robfig/cron.v2"
)

type JsonRpcTemplate struct {
	Jsonrpc string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

type JsonRpcCall struct {
	Id      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  string `json:"params"`
}

type Job struct {
	Id           uuid.UUID       `json:"id"`
	Schedule     string          `json:"schedule"`
	CallTemplate JsonRpcTemplate `json:"callTemplate"`
}

func (j Job) Run() {
	fmt.Println(j.CallTemplate)
}

type JobManager struct {
	entryMap map[uuid.UUID]cron.EntryID
	cron     *cron.Cron
}

func NewJobManager() *JobManager {
	manager := &JobManager{
		make(map[uuid.UUID]cron.EntryID),
		cron.New(),
	}

	/*
	 * Example schedule formats:
	 * c.AddFunc("0 30 * * * *", func() { fmt.Println("Every hour on the half hour") })
	 * c.AddFunc("@hourly", func() { fmt.Println("Every hour") })
	 * c.AddFunc("@every 1h30m", func() { fmt.Println("Every hour thirty") })
	 */
	// add a heartbeat
	manager.cron.AddFunc("@every 10s", func() { fmt.Println("..") })
	manager.cron.Start()

	return manager
}

func (m *JobManager) GetJob(jobId uuid.UUID) Job {
	_ = m.entryMap[jobId]
	// 	// entry := m.cron.Entry(entryId)

	return Job{
		Id:           jobId,
		Schedule:     "placeholder",
		CallTemplate: JsonRpcTemplate{},
	}
}

func (m *JobManager) CreateJob(job Job) (Job, error) {
	if job.Id != uuid.Nil {
		return Job{}, errors.New("JobId may not be defined prior to creation")
	}

	cid, err := m.cron.AddJob(job.Schedule, job)
	if err != nil {
		return Job{}, errors.Wrap(err, "Failure adding job")
	}

	id := uuid.New()
	m.entryMap[id] = cid
	job.Id = id

	return job, nil
}
