package drivers

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const EVENTS_API = "https://events.pagerduty.com/generic/2010-04-15/create_event.json"

type PagerDuty struct {
	ServiceKey string
}

func (p *PagerDuty) Trigger() *PagerDutyAlert {
	return &PagerDutyAlert{
		ServiceKey: p.ServiceKey,
		EventType:  "trigger",
	}
}

type PagerDutyAlert struct {
	ServiceKey  string                 `json:"service_key"`
	IncidentKey string                 `json:"incident_key"`
	EventType   string                 `json:"event_type"`
	Description string                 `json:"description"`
	Client      string                 `json:"client"`
	ClientUrl   string                 `json:"client_url"`
	Details     map[string]interface{} `json:"details"`
	Contexts    []PagerDutyContext     `json:"contexts"`
}

type PagerDutyContext struct {
	Type string `json:"type"`
	Href string `json:"href"`
	Text string `json:"text"`
	Src  string `json:"src"`
}

type PagerDutyResponse struct {
	Status      string   `json:"status"`
	Message     string   `json:"message"`
	IncidentKey string   `json:"incident_key"`
	Errors      []string `json:"errors"`
}

func (p *PagerDutyAlert) Send() error {
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(p)
	if err != nil {
		return err
	}
	resp, err := http.Post(EVENTS_API, "application/json", buf)
	if err != nil {
		return err
	}

	// Set the incident key based on response
	pdResponse := &PagerDutyResponse{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(pdResponse)

	if err != nil {
		return err
	}

	p.IncidentKey = pdResponse.IncidentKey
	return nil
}
