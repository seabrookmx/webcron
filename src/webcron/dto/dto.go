package dto

import (
	"github.com/globalsign/mgo/bson"
)

type JsonRpcCall struct {
	Jsonrpc string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	URL     string                 `json:"url"`
}

type Job struct {
	Id bson.ObjectId `json:"id" bson:"_id"`
	// TODO: add namespace and name properties
	Schedule     string      `json:"schedule"`
	CallTemplate JsonRpcCall `json:"callTemplate"`
}
