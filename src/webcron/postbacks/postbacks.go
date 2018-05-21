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
	FireWebhook(call dto.JsonRpcCall) error
}

type LogPostbackTrigger struct{}

func NewLogPostbackTrigger() *LogPostbackTrigger {
	return &LogPostbackTrigger{}
}

func (t *LogPostbackTrigger) FireWebhook(call dto.JsonRpcCall) error {
	fmt.Println("Logging call:")
	fmt.Println(call)
	return nil
}

type HttpPostbackTrigger struct{}

func NewHttpPostbackTrigger() *HttpPostbackTrigger {
	return &HttpPostbackTrigger{}
}

func (t *HttpPostbackTrigger) FireWebhook(call dto.JsonRpcCall) error {
	tmpl, err := template.New("t").Parse(call.Jsonrpc)

	buf := new(bytes.Buffer)
	if err == nil {
		err = tmpl.Execute(buf, call.Params)
	}
	if err != nil {
		return err
	}

	// tmpl, err := template.New("test").Parse("{{.Count}} items are made of {{.Material}}")
	// if err != nil { panic(err) }
	// err = tmpl.Execute(os.Stdout, sweaters)
	// if err != nil { panic(err) }

	resp, err := http.Post(call.URL, "application/json", buf)

	if err != nil {
		return err
	} else if resp.StatusCode >= 200 || resp.StatusCode <= 299 {
		return nil
	} else {
		return errors.Errorf("Http Status Code %d received", resp.StatusCode)
	}
}
