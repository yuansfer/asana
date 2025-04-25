package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/yuansfer/log"

	"git.drinkme.beer/yinghe/asana/module"
	"git.drinkme.beer/yinghe/asana/util"
)

type Data struct {
	Request Request `json:"data"`
}

type Request struct {
	ResourceSubtype string `json:"resource_subtype,omitempty"`
	Assignee        string `json:"assignee,omitempty"`
	Name            string `json:"name,omitempty"`
	Completed       bool   `json:"completed"`
	DueOn           string `json:"due_on,omitempty"`
	Liked           bool   `json:"linked,omitempty"`
	// Notes refer to the content of each Ticket
	Notes        string            `json:"notes,omitempty"`
	HtmlNotes    string            `json:"html_notes,omitempty"`
	StartOn      string            `json:"start_on,omitempty"`
	CustomFields map[string]string `json:"custom_fields,omitempty"`
	Projects     []string          `json:"projects,omitempty"`
	Workspace    string            `json:"workspace,omitempty"`

	TicketID   string `json:"-"`
	TicketType string `json:"-"`
	paToken    string
}

func (r *Request) GetPAToken() string {
	return r.paToken
}

func (r *Request) SetPAToken(token string) {
	r.paToken = token
}

type Response struct {
	ID             string `json:"gid"`
	Name           string `json:"name"`
	ResourceType   string `json:"resource_type"`
	AssigneeStatus string `json:"assignee_status"`
}

const (
	URICreateTask   = "/api/1.0/tasks"
	URIUpdateTask   = "/api/1.0/tasks/%s"
	ResourceSubtype = "default_task"
)

func (r Request) Update(taskID string) error {
	var (
		uri = fmt.Sprintf(URIUpdateTask, taskID)
		err error
	)
	c, err := r.call(uri, util.HttpPutMethod)
	if nil != err || nil == c {
		return err
	}
	if http.StatusOK != c.HTTPStatus {
		c.Print()
		log.Errorf("unexpected response")
		err = errors.New("unexpected response")
		return err
	}
	return nil
}

func (r Request) Create() (resp *Response, err error) {
	var (
		respData struct {
			Response Response      `json:"data"`
			Errors   module.Errors `json:"errors"`
		}
	)
	client, err := r.call(URICreateTask, util.HttpPostMethod)

	log.Infof("response status: %d", client.HTTPStatus)
	if http.StatusCreated != client.HTTPStatus {
		client.Print()
		log.Errorf("unexpected response")
		err = errors.New("unexpected response")
		return
	}

	err = json.Unmarshal(client.Body, &respData)
	if nil != err {
		client.Print()
		log.Errorf("illegal task result: %s", err.Error())
		return
	}

	resp = &respData.Response
	log.Infof("[%s]new task ID:%s", r.TicketID, resp.ID)

	return
}

func (r Request) call(uri, httpMethod string) (*util.Client, error) {
	var (
		err     error
		reqData Data
		client  = new(util.Client)
	)

	headers := make(map[string]string)
	headers["Authorization"] = r.paToken
	headers["Content-Type"] = util.ContentType

	if "" == r.ResourceSubtype {
		r.ResourceSubtype = ResourceSubtype
	}

	reqData.Request = r

	buf, err := json.Marshal(&reqData)
	if nil != err {
		log.Errorf("failed to generate task request: %s", err.Error())
		return client, err
	}

	log.Infof("request: %s", string(buf))

	client = util.NewHttpClient(util.AsanaHost, uri, httpMethod, buf)
	client.Headers = headers

	err = client.Request()
	if nil != err {
		log.Errorf("failed to create new Asana task: %s", err.Error())
		return client, err
	}

	return client, nil
}
