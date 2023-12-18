package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/k8sdeploy/agent/internal/agent/deploy"
	"github.com/k8sdeploy/agent/internal/agent/info"
	"net/http"
)

type ActionType string

const (
	Deploy ActionType = "deploy"
	Delete ActionType = "delete"

	Information ActionType = "info"
)

//nolint:gocyclo
//func (a *Agent) listenForSelfUpdate(errChan chan error) {
//	channel := fmt.Sprintf("%s/application/%s/message?limit=1", a.Config.K8sDeploy.SocketAddress, a.SelfUpdate.ID)
//	// fmt.Printf("self-update channel %s\n", channel)
//
//	req, err := http.NewRequest("GET", channel, nil)
//	if err != nil {
//		errChan <- logs.Errorf("failed to create request: %v", err)
//		return
//	}
//	req.Header.Set("X-Gotify-Key", a.SelfUpdate.Token)
//	// fmt.Printf("self-update token %s\n", a.SelfUpdate.Token)
//	res, err := http.DefaultClient.Do(req)
//	if err != nil {
//		errChan <- logs.Errorf("failed to get self-update: %v", err)
//		return
//	}
//
//	defer func() {
//		if err := res.Body.Close(); err != nil {
//			_ = logs.Errorf("failed to close body: %v", err)
//		}
//	}()
//
//	if res.StatusCode != http.StatusOK {
//		errChan <- logs.Errorf("failed get self-update: %s", res.Status)
//		return
//	}
//
//	type messages struct {
//		Messages []Message `json:"messages"`
//		Paging   Paging    `json:"paging"`
//	}
//	var m messages
//	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
//		errChan <- logs.Errorf("failed to decode self-update: %v", err)
//		return
//	}
//
//	type message struct {
//		Version string `json:"version"`
//	}
//
//	if len(m.Messages) >= 1 {
//		if m.Messages[0].Title == "update" {
//			var msg message
//			if err := json.Unmarshal([]byte(m.Messages[0].Message), &msg); err != nil {
//				errChan <- logs.Errorf("failed to unmarshal self-update: %v", err)
//			}
//
//			switch a.Config.K8sDeploy.BuildVersion {
//			case "dev":
//				return
//			case "latest":
//				return
//			case msg.Version:
//				return
//			}
//
//			errChan <- deploy.NewDeployment(a.KubernetesClient.ClientSet, a.KubernetesClient.Context).DeployImage(m.Messages[0].Message)
//		}
//	}
//}

func (a *Agent) getMessage(queue string, requeue bool) (string, error) {
	ackMode := "ack_requeue_false"
	if requeue {
		ackMode = "ack_requeue_true"
	}

	type Payload struct {
		AckMode  string `json:"ackmode"`
		Count    int    `json:"count"`
		Encoding string `json:"encoding"`
		Truncate int    `json:"truncate"`
	}
	payload, err := json.Marshal(&Payload{
		AckMode:  ackMode,
		Count:    1,
		Encoding: "auto",
		Truncate: 50000,
	})
	if err != nil {
		return "", logs.Errorf("failed to marshal %s payload: %v", queue, err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/queues/%s/%s/get", a.Config.Rabbit.Host, queue, queue), bytes.NewBuffer(payload))
	if err != nil {
		return "", logs.Errorf("failed to create %s request: %v", queue, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(a.Config.K8sDeploy.Credentials.Queue.Key, a.Config.K8sDeploy.Credentials.Queue.Secret)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", logs.Errorf("failed to get %s events: %v", queue, err)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			_ = logs.Errorf("failed to close %s queue body: %v", queue, err)
		}
	}()
	if res.StatusCode != http.StatusOK {
		return "", logs.Errorf("failed get %s events: %s", queue, res.Status)
	}

	// there isn't any messages
	if res.ContentLength <= 10 {
		return "", nil
	}

	type Message struct {
		Exchange        string   `json:"exchange"`
		MessageCount    int      `json:"message_count"`
		Payload         string   `json:"payload"`
		PayloadBytes    int      `json:"payload_bytes"`
		PayloadEncoding string   `json:"payload_encoding"`
		Properties      []string `json:"properties"`
		Redelivered     bool     `json:"redelivered"`
		RoutingKey      string   `json:"routing_key"`
	}

	var m []Message
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return "", logs.Errorf("failed to decode %s events: %v", queue, err)
	}

	return m[0].Payload, nil
}

func (a *Agent) listenForSelfUpdate(errChan chan error) {
	updateMessage, err := a.getMessage(a.Config.K8sDeploy.Queues.Master, false)
	if err != nil {
		errChan <- logs.Errorf("failed to get message: %v", err)
		return
	}

	fmt.Printf("updateMessage: %+v\n", updateMessage)
}

func (a *Agent) listenForEvents(errChan chan error) {
	queueMessage, err := a.getMessage(a.Config.K8sDeploy.Queues.Agent, false)
	if err != nil {
		errChan <- logs.Errorf("failed to get message: %v", err)
		return
	}

	type AgentMessage struct {
		Action ActionType `json:"action"`
	}

	if queueMessage == "" {
		return
	}

	var am AgentMessage
	if err := json.Unmarshal([]byte(queueMessage), &am); err != nil {
		errChan <- logs.Errorf("failed to unmarshal queueMessage: %v", err)
		return
	}

	switch am.Action {
	case Deploy:
		errChan <- deploy.NewDeployment(a.KubernetesClient.ClientSet, a.KubernetesClient.Context).DeployImage(queueMessage)
	case Delete:
		errChan <- deploy.NewDeployment(a.KubernetesClient.ClientSet, a.KubernetesClient.Context).DeleteDeployment(queueMessage)
	case Information:
		errChan <- info.NewInfo(a.KubernetesClient.ClientSet, a.KubernetesClient.Context).ParseInfoRequest(queueMessage)
	default:
		errChan <- logs.Errorf("unknown action: %s", am.Action)
	}

	fmt.Printf("agentMessage: %+v\n", am)
}

//nolint:gocyclo
//func (a *Agent) listenForEventsOld(errChan chan error) {
//	channel := fmt.Sprintf("%s/application/%s/message?limit=1", a.Config.K8sDeploy.SocketAddress, a.EventClient.ID)
//	// fmt.Printf("events channel %s\n", channel)
//
//	req, err := http.NewRequest("GET", channel, nil)
//	if err != nil {
//		errChan <- logs.Errorf("failed to create request: %v", err)
//		return
//	}
//	req.Header.Set("X-Gotify-Key", a.EventClient.Token)
//	// fmt.Printf("events token %s\n", a.EventClient.Token)
//	res, err := http.DefaultClient.Do(req)
//	if err != nil {
//		errChan <- logs.Errorf("failed to get service keys events: %v", err)
//		return
//	}
//	defer func() {
//		if err := res.Body.Close(); err != nil {
//			_ = logs.Errorf("failed to close body: %v", err)
//		}
//	}()
//	if res.StatusCode != http.StatusOK {
//		errChan <- logs.Errorf("failed get service keys events: %s", res.Status)
//		return
//	}
//
//	type messages struct {
//		Messages []Message `json:"messages"`
//		Paging   Paging    `json:"paging"`
//	}
//	var m messages
//	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
//		errChan <- logs.Errorf("failed to decode service keys events: %v", err)
//		return
//	}
//
//	var messageErr error
//
//	if len(m.Messages) >= 1 {
//		messageParsed := false
//
//		switch m.Messages[0].Title {
//		case messageTypeDeploy:
//			messageErr = deploy.NewDeployment(a.KubernetesClient.ClientSet, a.KubernetesClient.Context).DeployImage(m.Messages[0].Message)
//			messageParsed = true
//		case messageTypeInfo:
//			messageErr = info.NewInfo(a.KubernetesClient.ClientSet, a.KubernetesClient.Context).ParseInfoRequest(m.Messages[0].Message)
//			messageParsed = true
//		default:
//			logs.Infof("unknown message type: %s\n", m.Messages[0].Title)
//		}
//
//		if messageErr != nil {
//			logs.Infof("Error: %s\n", messageErr)
//		}
//
//		if messageParsed {
//			a.deleteMessage(m.Messages[0].MessageID, errChan)
//		}
//	}
//
//	errChan <- messageErr
//}
