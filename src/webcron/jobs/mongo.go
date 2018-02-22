package jobs

import (
	"os"

	"github.com/globalsign/mgo"
	"github.com/pkg/errors"
)

type mongoSession struct {
	session *mgo.Session
}

func newSession() (*mongoSession, error) {
	mongoUrl, urlExists := os.LookupEnv("WEBCRON_MONGO_URL")
	if urlExists == false {
		mongoUrl = "localhost"
	}

	mgoSession, err := mgo.Dial(mongoUrl)
	if err != nil {
		return nil, errors.Wrap(err, "Could not connect to database at "+mongoUrl)
	}

	return &mongoSession{
		session: mgoSession,
	}, nil
}

func (m *mongoSession) get() *mgo.Collection {
	return m.session.Copy().DB("WebCron").C("Jobs")
}
