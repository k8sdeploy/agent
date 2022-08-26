package agent

import (
	"encoding/json"
	"fmt"
	"github.com/k8sdeploy/agent/internal/agent/deploy"
	"github.com/k8sdeploy/agent/internal/agent/info"
	"net/http"
)

type Message struct {
	AppID     int                    `json:"appid"`
	Date      string                 `json:"date"`
	Extras    map[string]interface{} `json:"extras"`
	MessageID int                    `json:"id"`
	Message   string                 `json:"message"`
	Title     string                 `json:"title"`
	Priority  int                    `json:"priority"`
}

type Paging struct {
	Limit int    `json:"limit"`
	Next  string `json:"next"`
	Since int    `json:"since"`
	Size  int    `json:"size"`
}

const (
	MESSAGE_TYPE_DEPLOY = "deploy"
	MESSAGE_TYPE_INFO   = "info"
)

func (a *Agent) listenForEvents(errChan chan error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s?limit=1", a.EventClient.EventChannel), nil)
	if err != nil {
		errChan <- err
		return
	}
	req.Header.Set("X-Gotify-Key", a.EventClient.Token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		errChan <- err
		return
	}
	if res.StatusCode != http.StatusOK {
		errChan <- fmt.Errorf("failed get service keys: %s", res.Status)
		return
	}

	type messages struct {
		Messages []Message `json:"messages"`
		Paging   Paging    `json:"paging"`
	}
	var m messages
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		errChan <- err
		return
	}

	var messageErr error

	if len(m.Messages) >= 1 {
		messageParsed := false

		switch m.Messages[0].Title {
		case MESSAGE_TYPE_DEPLOY:
			messageErr = deploy.NewDeployment(a.KubernetesClient.ClientSet, a.KubernetesClient.Context).DeployImage(m.Messages[0].Message)
			messageParsed = true
		case MESSAGE_TYPE_INFO:
			messageErr = info.NewInfo(a.KubernetesClient.ClientSet, a.KubernetesClient.Context).ParseInfoRequest(m.Messages[0].Message)
			messageParsed = true
		default:
			fmt.Printf("unknown message type: %s\n", m.Messages[0].Title)
		}

		if messageErr != nil {
			fmt.Printf("Error: %s\n", messageErr)
		}

		if messageParsed {
			a.deleteMessage(m.Messages[0].MessageID, errChan)
		}
	}

	errChan <- messageErr
}

func (a *Agent) deleteMessage(messageId int, errChan chan error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/message/%d", a.Config.K8sDeploy.SocketAddress, messageId), nil)
	if err != nil {
		errChan <- err
		return
	}
	req.Header.Set("X-Gotify-Key", a.EventClient.Token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		errChan <- err
		return
	}
	if res.StatusCode != http.StatusOK {
		errChan <- fmt.Errorf("failed get service keys: %s", res.Status)
		return
	}
}
