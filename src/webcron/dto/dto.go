package dto

import (
	"github.com/globalsign/mgo/bson"
)

type JsonRPCCall struct {
	Jsonrpc string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	URL     string                 `json:"url"`
}

type Job struct {
	ID           bson.ObjectId `json:"id" bson:"_id"`
	Name         string        `json:"name"`
	Schedule     string        `json:"schedule"`
	CallTemplate JsonRPCCall   `json:"callTemplate"`
}
