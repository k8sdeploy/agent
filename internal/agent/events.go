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

type PayloadDetails struct {
	Action        ActionType `json:"action"`
	RequestID     string     `json:"request_id"`
	ActionDetails struct {
		Type string `json:"type"`
	} `json:"action_details"`
	DeployDetails interface{} `json:"deploy_details"`
	InfoDetails   interface{} `json:"info_details"`
}

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
		Truncate: 5000000,
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

	if len(m) == 0 {
		return "", nil
	}

	if m[0].PayloadBytes < 10 || m[0].Payload == "" {
		return "", nil
	}

	return m[0].Payload, nil
}

func (a *Agent) listenForSelfUpdate(errChan chan error) {
	//updateMessage, err := a.getMessage(a.Config.K8sDeploy.Queues.Master, false)
	//if err != nil {
	//	errChan <- logs.Errorf("failed to get message: %v", err)
	//	return
	//}

	//fmt.Printf("updateMessage: %+v\n", updateMessage)
}

func (a *Agent) listenForEvents(errChan chan error) {
	queueMessage, err := a.getMessage(a.Config.K8sDeploy.Queues.Agent, false)
	if err != nil {
		errChan <- logs.Errorf("failed to get message: %v", err)
		return
	}

	if queueMessage == "" {
		errChan <- nil
		return
	}

	var payload PayloadDetails
	if err := json.Unmarshal([]byte(queueMessage), &payload); err != nil {
		errChan <- logs.Errorf("failed to unmarshal queueMessage: %v", err)
		return
	}

	switch payload.Action {
	case Deploy:
		deployType := deploy.DeployType(payload.ActionDetails.Type)
		errChan <- deploy.NewDeployment(a.KubernetesClient.ClientSet, a.KubernetesClient.Context, deployType, payload.RequestID).ParseRequest(payload.DeployDetails)
	case Information:
		infoType := info.InfoType(payload.ActionDetails.Type)
		errChan <- info.NewInfo(a.KubernetesClient.ClientSet, a.KubernetesClient.Context, infoType, payload.RequestID).ParseRequest(payload.InfoDetails)
	default:
		logs.Info("unknown, %s", queueMessage)
		errChan <- logs.Errorf("unknown action: %s", payload.Action)
	}
}
