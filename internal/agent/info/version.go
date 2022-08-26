package info

import "fmt"

const VERSION_REQUEST_TYPE = "version"

type VersionRequest struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type VersionSendResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Version   string `json:"version"`
}

func NewVersion() *VersionRequest {
	return &VersionRequest{}
}

func (n *VersionRequest) ParseRequest(msgMap map[string]interface{}) error {
	namespace := msgMap["namespace"]
	if namespace == nil {
		return fmt.Errorf("namespace is required")
	}

	name := msgMap["name"]
	if name == nil {
		return fmt.Errorf("name is required")
	}
	n.Namespace = namespace.(string)
	n.Name = name.(string)

	return nil
}

func (n *VersionRequest) SendResponse() error {
	return nil
}
