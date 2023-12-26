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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/queues/%s/%s/get", a.Config.K8sDeploy.RabbitHost, queue, queue), bytes.NewBuffer(payload))
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
	updateMessage, err := a.getMessage(a.Config.K8sDeploy.Queues.Master, true)
	if err != nil {
		errChan <- logs.Errorf("failed to get message: %v", err)
		return
	}

	if updateMessage != "" {
		fmt.Printf("updateMessage: %+v\n", updateMessage)
	}
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
		d := deploy.NewDeployment(a.KubernetesClient.ClientSet, a.KubernetesClient.Context)
		d.SetDeploymentType(deploy.TypeDeploy(payload.ActionDetails.Type))
		d.SetRequestID(payload.RequestID)
		errChan <- d.ParseRequest(payload.DeployDetails)
		errChan <- d.SendResponse(a.Config)
	case Information:
		i := info.NewInfo(a.KubernetesClient.ClientSet, a.KubernetesClient.Context)
		i.SetInfoType(info.TypeInfo(payload.ActionDetails.Type))
		i.SetRequestID(payload.RequestID)
		errChan <- i.ParseRequest(payload.InfoDetails)
		errChan <- i.SendResponse(a.Config)
	default:
		logs.Info("unknown, %s", queueMessage)
		errChan <- logs.Errorf("unknown action: %s", payload.Action)
	}
}
