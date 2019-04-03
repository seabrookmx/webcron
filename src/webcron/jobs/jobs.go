package jobs

import (
	"fmt"

	"webcron/dto"
	"webcron/postbacks"

	"github.com/globalsign/mgo/bson"
	"github.com/pkg/errors"
	"gopkg.in/robfig/cron.v2"
)

type JobManager struct {
	entryMap        map[bson.ObjectId]cron.EntryID
	cron            *cron.Cron
	session         *mongoSession
	postbackTrigger postbacks.PostbackTrigger
}

func NewJobManager(callbackTrigger postbacks.PostbackTrigger) (*JobManager, error) {
	session, err := newSession()
	if err != nil {
		return nil, errors.Wrap(err, "Error initializing JobManager")
	}

	m := &JobManager{
		make(map[bson.ObjectId]cron.EntryID),
		cron.New(),
		session,
		callbackTrigger,
	}
	/*
	 * Example schedule formats:
	 * c.AddFunc("0 30 * * * *", func() { fmt.Println("Every hour on the half hour") })
	 * c.AddFunc("@hourly", func() { fmt.Println("Every hour") })
	 * c.AddFunc("@every 1h30m", func() { fmt.Println("Every hour thirty") })
	 */

	// add a heartbeat
	_, err = m.cron.AddFunc("@every 10s", func() { fmt.Println("-tick- (10s)") })
	if err != nil {
		return nil, errors.Wrap(err, "Error adding heartbeat job")
	}

	// schedule the jobs stored in mongo
	results := make([]dto.Job, 0)
	err = m.session.get().Find(nil).Iter().All(&results)
	for _, job := range results {
		// TODO: BEDUG
		_, ok := job.Webhook.Template.(bson.M)
		if ok {
			fmt.Println("do stuff!")
		}

		cid, err := m.cron.AddFunc(job.Schedule, func() { m.RunJob(job) })
		if err != nil {
			break
		}
		m.entryMap[job.ID] = cid
	}

	if err != nil {
		return nil, errors.Wrap(err, "Error initializing jobs")
	}

	m.cron.Start()

	return m, nil
}

func (m *JobManager) RunJob(job dto.Job) {
	/*
	 * We implement the Run function on the manager
	 * instead of the Job so we can get access to some
	 * global state, such as configuration params.
	 */
	err := m.postbackTrigger.FireWebhook(job.Webhook)

	if err != nil {
		fmt.Printf("Postback error (%s): %v\n", job.ID.String(), err)
	}
}

func (m *JobManager) GetJob(jobId string) dto.Job {
	id := bson.ObjectIdHex(jobId)

	result := make([]dto.Job, 0)
	err := m.session.get().FindId(id).Iter().All(&result)
	if err != nil || len(result) == 0 {
		return dto.Job{}
	}

	return result[0]
}

func (m *JobManager) GetJobs(limit int, skip int) []dto.Job {
	result := make([]dto.Job, 0)
	err := m.session.get().Find(nil).Skip(skip).Limit(limit).Iter().All(&result)
	if err != nil {
		return []dto.Job{}
	}

	return result
}

func (m *JobManager) RemoveJob(jobId string) error {
	job := m.GetJob(jobId)

	if job.ID == bson.NewObjectId() {
		return errors.New("Job not found")
	}

	err := m.session.get().RemoveId(job.ID)
	if err != nil {
		return err
	}

	m.cron.Remove(m.entryMap[job.ID])
	return nil
}

func (m *JobManager) CreateJob(job dto.Job) (dto.Job, error) {
	if job.Name == "" {
		return dto.Job{}, errors.New("Job must have a name")
	}
	if job.ID != "" {
		return dto.Job{}, errors.New("Job ID may not be defined prior to creation")
	}
	job.ID = bson.NewObjectId()

	var cid cron.EntryID
	err := m.session.get().Insert(job)
	if err == nil {
		cid, err = m.cron.AddFunc(job.Schedule, func() { m.RunJob(job) })
	}
	if err != nil {
		return dto.Job{}, errors.Wrap(err, "Failure adding job")
	}

	m.entryMap[job.ID] = cid

	return job, nil
}
