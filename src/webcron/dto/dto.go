package dto

import (
	"github.com/globalsign/mgo/bson"
)

type Webhook struct {
	Template interface{}            `json:"template"`
	Method   string                 `json:"method"`
	Params   map[string]interface{} `json:"params"`
	URL      string                 `json:"url"`
}

type Job struct {
	ID       bson.ObjectId `json:"id" bson:"_id"`
	Name     string        `json:"name"`
	Schedule string        `json:"schedule"`
	Webhook  Webhook       `json:"webhook"`
}
