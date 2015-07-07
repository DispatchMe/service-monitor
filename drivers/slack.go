package drivers

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Slack struct {
	WebhookUrl string
}

type SlackMessage struct {
	Text string `json:"text"`
}

func (s *Slack) SendError(errToSend error) error {
	buf := &bytes.Buffer{}

	msg := &SlackMessage{
		Text: errToSend.Error(),
	}

	encoder := json.NewEncoder(buf)
	err := encoder.Encode(msg)
	if err != nil {
		return err
	}
	_, err = http.Post(s.WebhookUrl, "application/json", buf)
	return err
}
