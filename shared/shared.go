package shared

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
