package jobs

import (
	"fmt"

	"github.com/globalsign/mgo/bson"
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
	Id bson.ObjectId `json:"id" bson:"_id"`
	// TODO: add namespace and name properties
	Schedule     string          `json:"schedule"`
	CallTemplate JsonRpcTemplate `json:"callTemplate"`
}

type JobManager struct {
	entryMap map[bson.ObjectId]cron.EntryID
	cron     *cron.Cron
	session  *mongoSession
	dryRun   bool
}

func NewJobManager(dryRun bool) (*JobManager, error) {
	session, err := newSession()
	if err != nil {
		return nil, errors.Wrap(err, "Error initializing JobManager")
	}

	m := &JobManager{
		make(map[bson.ObjectId]cron.EntryID),
		cron.New(),
		session,
		dryRun,
	}
	/*
	 * Example schedule formats:
	 * c.AddFunc("0 30 * * * *", func() { fmt.Println("Every hour on the half hour") })
	 * c.AddFunc("@hourly", func() { fmt.Println("Every hour") })
	 * c.AddFunc("@every 1h30m", func() { fmt.Println("Every hour thirty") })
	 */

	// add a heartbeat
	_, err = m.cron.AddFunc("@every 10s", func() { fmt.Println("..") })
	if err != nil {
		return nil, errors.Wrap(err, "Error adding heartbeat job")
	}

	// schedule the jobs stored in mongo
	results := make([]Job, 0)
	err = m.session.get().Find(nil).Iter().All(&results)
	for _, job := range results {
		cid, err := m.cron.AddFunc(job.Schedule, func() { m.RunJob(job) })
		if err != nil {
			break
		}
		m.entryMap[job.Id] = cid
	}

	if err != nil {
		return nil, errors.Wrap(err, "Error initializing jobs")
	}

	m.cron.Start()

	return m, nil
}

func (m *JobManager) RunJob(job Job) {
	/*
	 * We implement the Run function on the manager
	 * instead of the Job so we can get access to some
	 * global state, such as configuration params.
	 */
	if m.dryRun {
		// TODO: replace fmt.Println with calls to logrus
		fmt.Println(job.CallTemplate)
	} else {
		panic(errors.New("Only dryRun mode is currently supported - FIXME"))
	}
}

func (m *JobManager) GetJob(jobId string) Job {
	id := bson.ObjectIdHex(jobId)

	result := make([]Job, 0)
	err := m.session.get().FindId(id).Iter().All(&result)
	if err != nil || len(result) == 0 {
		return Job{}
	}

	return result[0]
}

// TODO: GetJobs (with paging)

// TODO: RemoveJob

func (m *JobManager) CreateJob(job Job) (Job, error) {
	if job.Id != "" {
		return Job{}, errors.New("JobId may not be defined prior to creation")
	}
	job.Id = bson.NewObjectId()

	cid, err := m.cron.AddFunc(job.Schedule, func() { m.RunJob(job) })
	if err == nil {
		err = m.session.get().Insert(job)
	}
	if err != nil {
		return Job{}, errors.Wrap(err, "Failure adding job")
	}

	m.entryMap[job.Id] = cid

	return job, nil
}
