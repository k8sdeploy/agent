package agent

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/k8sdeploy/agent/internal/agent/deploy"
	"github.com/k8sdeploy/agent/internal/agent/info"
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
	messageTypeDeploy = "deploy"
	messageTypeInfo   = "info"
)

//nolint:gocyclo
func (a *Agent) listenForSelfUpdate(errChan chan error) {
	channel := fmt.Sprintf("%s/application/%s/message?limit=1", a.Config.K8sDeploy.SocketAddress, a.SelfUpdate.ID)
	fmt.Printf("self-update channel %s\n", channel)

	req, err := http.NewRequest("GET", channel, nil)
	if err != nil {
		errChan <- err
		return
	}
	req.Header.Set("X-Gotify-Key", a.SelfUpdate.Token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		errChan <- err
		return
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}()

	if res.StatusCode != http.StatusOK {
		errChan <- fmt.Errorf("failed get updates: %s", res.Status)
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

	type message struct {
		Version string `json:"version"`
	}

	if len(m.Messages) >= 1 {
		if m.Messages[0].Title == "update" {
			var msg message
			if err := json.Unmarshal([]byte(m.Messages[0].Message), &msg); err != nil {
				errChan <- err
			}

			switch a.Config.Local.BuildVersion {
			case "dev":
				return
			case "latest":
				return
			case msg.Version:
				return
			}

			errChan <- deploy.NewDeployment(a.KubernetesClient.ClientSet, a.KubernetesClient.Context).DeployImage(m.Messages[0].Message)
		}
	}
}

//nolint:gocyclo
func (a *Agent) listenForEvents(errChan chan error) {
	channel := fmt.Sprintf("%s/application/%s/message?limit=1", a.Config.K8sDeploy.SocketAddress, a.EventClient.ID)
	fmt.Printf("events channel %s\n", channel)

	req, err := http.NewRequest("GET", channel, nil)
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
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}()
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
		case messageTypeDeploy:
			messageErr = deploy.NewDeployment(a.KubernetesClient.ClientSet, a.KubernetesClient.Context).DeployImage(m.Messages[0].Message)
			messageParsed = true
		case messageTypeInfo:
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

func (a *Agent) deleteMessage(messageID int, errChan chan error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/message/%d", a.Config.K8sDeploy.SocketAddress, messageID), nil)
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
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}()
	if res.StatusCode != http.StatusOK {
		errChan <- fmt.Errorf("failed get service keys: %s", res.Status)
		return
	}
}
