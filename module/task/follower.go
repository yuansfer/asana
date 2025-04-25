package task

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yuansfer/log"

	"github.com/yuansfer/asana/util"
)

const (
	AddFollowersURI = "/api/1.0/tasks/%s/addFollowers"
)

type Follower struct {
	Followers []string `json:"followers,omitempty"`
}

type FollowersRequest struct {
	Data Follower `json:"data"`
}

func (f Follower) Add(token, taskID string) error {
	var (
		req FollowersRequest
	)
	if "" == token || "" == taskID {
		log.Errorf("illegal request")
		return fmt.Errorf("illegal request")
	}
	if nil == f.Followers {
		log.Errorf("invalid request, followers")
		return fmt.Errorf("none followers")
	}

	req.Data = f
	buf, err := json.Marshal(&req)
	if nil != err {
		log.Errorf("invalid arguments: %s", err.Error())
		return err
	}
	log.Infof("request ==> %s", string(buf))

	headers := make(map[string]string)
	headers["Authorization"] = token
	headers["Content-Type"] = util.ContentType

	client := util.NewHttpClient(util.AsanaHost,
		fmt.Sprintf(AddFollowersURI, taskID),
		util.HttpPostMethod, buf)
	client.Headers = headers

	err = client.Request()
	if nil != err {
		log.Errorf("failed to create new Asana task: %s", err.Error())
		return err
	}
	log.Infof("response status: %d", client.HTTPStatus)
	if http.StatusOK != client.HTTPStatus {
		log.Errorf("unexpected response")
		err = fmt.Errorf("unexpected response")
		return err
	} else {
		return nil
	}
}
