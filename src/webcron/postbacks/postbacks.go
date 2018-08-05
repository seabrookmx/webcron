package postbacks

import (
	"bytes"
	"fmt"
	"net/http"
	"text/template"

	"webcron/dto"

	"github.com/pkg/errors"
)

type PostbackTrigger interface {
	FireWebhook(call dto.JsonRPCCall) error
}

type LogPostbackTrigger struct{}

func NewLogPostbackTrigger() *LogPostbackTrigger {
	return &LogPostbackTrigger{}
}

func (t *LogPostbackTrigger) FireWebhook(call dto.JsonRPCCall) error {
	fmt.Println("Logging call:")
	fmt.Println(call)
	return nil
}

type HTTPPostbackTrigger struct {
	client *http.Client
}

func NewHTTPPostbackTrigger() *HTTPPostbackTrigger {
	return &HTTPPostbackTrigger{
		&http.Client{},
	}
}

func (t *HTTPPostbackTrigger) FireWebhook(call dto.JsonRPCCall) error {
	tmpl, err := template.New("t").Parse(call.Jsonrpc)

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, call.Params)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(call.Method, call.URL, buf)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := t.client.Do(req)

	if err != nil {
		return err
	} else if resp.StatusCode >= 200 || resp.StatusCode <= 299 {
		return nil
	} else {
		return errors.Errorf("Http Status Code %d received", resp.StatusCode)
	}
}
