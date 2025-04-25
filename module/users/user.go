package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/yuansfer/log"

	"git.drinkme.beer/yinghe/asana/util"
)

const (
	URI = "/api/1.0/users/%s"
)

type User struct {
	ID           string `json:"gid,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
	Name         string `json:"name,omitempty"`
	Email        string `json:"email,omitempty"`

	Photo      map[string]string `json:"photo,omitempty"`
	Workspaces []struct {
		ID           string `json:"gid,omitempty"`
		ResourceType string `json:"resource_type,omitempty"`
		Name         string `json:"name,omitempty"`
	} `json:"workspaces,omitempty"`
}

type Response struct {
	Data User `json:"data,omitempty"`
}

func (u User) Get(paToken string) (*User, error) {
	var (
		err  error
		resp Response
	)
	headers := make(map[string]string)
	headers["Authorization"] = paToken
	headers["Content-Type"] = util.ContentType

	client := util.NewHttpClient(util.AsanaHost, fmt.Sprintf(URI, u.ID), util.HttpGetMethod, nil)
	client.Headers = headers

	err = client.Request()
	if nil != err {
		log.Errorf("failed to obtain user info: %s", err.Error())
		return nil, err
	}

	log.Infof("response status: %d", client.HTTPStatus)
	if http.StatusOK != client.HTTPStatus {
		log.Errorf("unexpected response")
		err = errors.New("unexpected response")
		return nil, err
	}

	err = json.Unmarshal(client.Body, &resp)
	if nil != err {
		log.Errorf("illegal task result: %s", err.Error())
		return nil, err
	}
	return &resp.Data, err
}
