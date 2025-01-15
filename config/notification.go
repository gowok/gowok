package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type Notification struct {
	AppID      string `yaml:"app_id"`
	RestApiKey string `yaml:"rest_api_key"`
}

type CreateNotification struct {
	Included_segments []string       `json:"included_segments"`
	App_id            string         `json:"app_id"`
	Data              map[string]any `json:"data"`
	Contents          struct {
		English string `json:"en"`
	} `json:"contents"`
}

type CreateNotificationResponse struct {
	ID         string `json:"id"`
	ExternalID string `json:"external_id"`
}

var (
	baseUrl string = "https://onesignal.com/api/v1"
	client         = &http.Client{}
)

func (n *Notification) options(path string, body []byte) map[string]any {
	return map[string]any{
		"url": fmt.Sprintf("%s/%s", baseUrl, path),
		"headers": map[string]string{
			"Content-Type":  "application/json",
			"Authorization": fmt.Sprintf("Basic %s", n.RestApiKey),
		},
		"body": body,
	}
}

func (n *Notification) Create(method string, data CreateNotification) error {
	var clientResponse CreateNotificationResponse
	// marshall body
	dataMarshall, err := json.Marshal(data)
	if err != nil {
		return errors.New("failed to marshall body in notification")
	}

	options := n.options("notifications", dataMarshall)

	// init request
	request, err := http.NewRequest("POST", options["url"].(string), bytes.NewBuffer(options["body"].([]byte)))
	if err != nil {
		return err
	}

	// set headers
	request.Header.Set("Content-Type", options["headers"].(map[string]string)["Content-Type"])
	request.Header.Set("Authorization", options["headers"].(map[string]string)["Authorization"])

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&clientResponse); err != nil {
		return err
	}

	slog.Info("Notification delivered : ", clientResponse.ID)

	return nil
}
