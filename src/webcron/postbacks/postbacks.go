package postbacks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"webcron/dto"
	"webcron/self"

	"github.com/pkg/errors"
)

func augmentParams(params map[string]interface{}) map[string]interface{} {
	newParams := make(map[string]interface{})
	for k, v := range params {
		upperKey := strings.ToUpper(k[:1]) + k[1:]
		// Add capitalized version of lower-case params
		//  (this is because template ignores the lower-case params)
		if k != upperKey {
			newParams[upperKey] = v
		}

		// Add the original list
		newParams[k] = v
	}

	// Add the date
	newParams["Date"] = time.Now().Format(time.RFC3339)

	// Add caller info
	newParams["Caller"] = fmt.Sprintf("%s %s (%s)", self.Name, self.Version, self.GoVersion)

	return newParams
}

type PostbackTrigger interface {
	FireWebhook(call dto.Webhook) error
}

type LogPostbackTrigger struct{}

func NewLogPostbackTrigger() *LogPostbackTrigger {
	return &LogPostbackTrigger{}
}

func (t *LogPostbackTrigger) FireWebhook(call dto.Webhook) error {
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

func (t *HTTPPostbackTrigger) FireWebhook(call dto.Webhook) error {
	var templateStr string
	var err error

	jsonTemplate, ok := call.Template.(map[string]interface{})
	if ok {
		// TODO: fixme - this code path doesn't run because our type test above doesn't do what we want :( )

		// template was passed in as json?
		var bytes []byte
		bytes, err = json.Marshal(jsonTemplate)
		if err == nil {
			templateStr = string(bytes[:])
		}
	}
	if !ok || err != nil {
		// template was passed in as a string?
		templateStr, ok = call.Template.(string)
		if !ok {
			return errors.New("template must be a JSON object or string")
		}
	}

	buf := new(bytes.Buffer)
	tmpl, err := template.New("t").Parse(templateStr)
	if err != nil {
		return err
	}

	err = tmpl.Execute(buf, augmentParams(call.Params))
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
