package story

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"git.drinkme.beer/yinghe/log"

	"git.drinkme.beer/yinghe/asana/module/users"
	"git.drinkme.beer/yinghe/asana/util"
)

const (
	URI = "/api/1.0/tasks/%s/stories"
)

type Data struct {
	Request Request `json:"data"`
}

type Request struct {
	HtmlText    string `json:"html_text,omitempty"`
	Text        string `json:"text,omitempty"`
	StickerName string `json:"sticker_name,omitempty"`
	IsPinned    bool   `json:"is_pinned,omitempty"`

	taskID   string
	ticketID string
	paToken  string
}

func (r *Request) GetTaskID() string {
	return r.taskID
}

func (r *Request) SetTaskID(id string) {
	r.taskID = id
}

func (r *Request) GetTicketID() string {
	return r.ticketID
}

func (r *Request) SetTicketID(id string) {
	r.ticketID = id
}

func (r *Request) GetPAToken() string {
	return r.paToken
}

func (r *Request) SetPAToken(token string) {
	r.paToken = token
}

type Response struct {
	ID              string `json:"gid"`
	ResourceType    string `json:"resource_type"`
	ResourceSubtype string `json:"resource_subtype"`
	Type            string `json:"type"`
	Text            string `json:"text"`
	// IsPinned Conditional
	// Whether the story should be pinned on the resource.
	IsPinned    bool       `json:"is_pinned"`
	StickerName string     `json:"sticker_name,omitempty"`
	CreatedAt   string     `json:"created_at"`
	CreatedBy   users.User `json:"created_by"`
}

func (r *Response) IsComplete() bool {
	return "marked_complete" == r.ResourceSubtype
}

func (r Request) validate() error {
	if "" == r.taskID {
		log.Errorf("task id is empty")
		return errors.New("task id is empty")
	}
	return nil
}

// Create comments on task
func (r Request) Create() (resp *Response, err error) {
	var (
		reqData  Data
		respData struct {
			Response Response `json:"data"`
		}
	)
	if err = r.validate(); nil != err {
		return
	}

	resp = new(Response)

	headers := make(map[string]string)
	headers["Authorization"] = r.paToken
	headers["Content-Type"] = util.ContentType

	reqData.Request = r

	buf, err := json.Marshal(&reqData)
	if nil != err {
		log.Errorf("failed to generate comment request: %s", err.Error())
		return
	}

	log.Infof("request: %s", string(buf))

	client := util.NewHttpClient(util.AsanaHost, fmt.Sprintf(URI, r.taskID), util.HttpPostMethod, buf)
	client.Headers = headers

	err = client.Request()
	if nil != err {
		log.Errorf("failed to create new task comment: %s", err.Error())
		return
	}
	log.Infof("response status: %d", client.HTTPStatus)
	if http.StatusCreated != client.HTTPStatus {
		client.Print()
		err = errors.New("unexpected response")
		return
	}

	err = json.Unmarshal(client.Body, &respData)
	if nil != err {
		log.Errorf("illegal comment result: %s", err.Error())
		return
	}
	resp = &respData.Response
	log.Infof("[%s]new comment ID:%s", r.taskID, resp.ID)
	return
}

func (r Request) Get() ([]Response, error) {
	var (
		resp struct {
			Data []Response `json:"data"`
		}
	)
	headers := make(map[string]string)
	headers["Authorization"] = r.paToken
	headers["Content-Type"] = util.ContentType

	client := util.NewHttpClient(util.AsanaHost, fmt.Sprintf(URI, r.taskID), util.HttpGetMethod, nil)
	client.Headers = headers

	err := client.Request()
	if nil != err || http.StatusOK != client.HTTPStatus {
		client.Print()
		log.Errorf("failed")
	}

	err = json.Unmarshal(client.Body, &resp)
	if nil != err {
		log.Errorf("invalid response")
		return nil, fmt.Errorf("invalid response")
	}

	return resp.Data, nil
}
