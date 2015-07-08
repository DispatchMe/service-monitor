package shared

import (
	"strings"
)

type GraphiteConfig struct {
	Url      string
	Username string
	Password string
}

type PagerDutyConfig struct {
	ServiceKey string
}

type SlackConfig struct {
	WebhookUrl string
}

var GraphiteConf *GraphiteConfig
var PagerDutyConf *PagerDutyConfig
var SlackConf *SlackConfig

type MultiError struct {
	Errors []error
}

func (m *MultiError) Error() string {
	errs := make([]string, len(m.Errors))

	for i, err := range m.Errors {
		errs[i] = err.Error()
	}

	return strings.Join(errs, "\n")
}
