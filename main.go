package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/DispatchMe/service-monitor/checks"
	"github.com/DispatchMe/service-monitor/drivers"
	"github.com/DispatchMe/service-monitor/shared"
	"log"
	"os"
	"time"
)

type Config struct {
	Graphite  *shared.GraphiteConfig
	PagerDuty *shared.PagerDutyConfig
	Slack     *shared.SlackConfig
	Services  []*Service
}

type Service struct {
	Name           string
	GraphiteChecks []*checks.GraphiteCheck
	HttpChecks     []*checks.HttpCheck
}

var slackDriver *drivers.Slack

func main() {
	args := os.Args

	var confFile string
	if len(args) < 2 {
		confFile = "conf.toml"
	} else {
		confFile = args[1]
	}

	conf := &Config{}
	_, err := toml.DecodeFile(confFile, conf)
	if err != nil {
		log.Fatal(err)
	}

	shared.GraphiteConf = conf.Graphite
	shared.PagerDutyConf = conf.PagerDuty
	shared.SlackConf = conf.Slack

	slackDriver = &drivers.Slack{shared.SlackConf.WebhookUrl}

	errChan := make(chan error, 50)
	okChan := make(chan string, 50)

	log.Println("Starting processes...")
	go checkForAlerts(errChan, okChan)

	for {
		for _, s := range conf.Services {
			for _, c := range s.HttpChecks {
				log.Printf("Running %s - %s\n", c.Name, s.Name)
				go runCheck(c, s, errChan, okChan)
			}
			for _, c := range s.GraphiteChecks {
				log.Printf("Running %s - %s\n", c.Name, s.Name)
				go runCheck(c, s, errChan, okChan)
			}
		}
		time.Sleep(60 * time.Second)
	}

}

func checkForAlerts(errChan chan error, okChan chan string) {
	for {
		select {
		case err := <-errChan:
			log.Println("Error:", err)
			slackDriver.SendError(err)
		case msg := <-okChan:
			log.Println(msg)
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func runCheck(check checks.Check, service *Service, errChan chan error, okChan chan string) {
	err := check.Run(service.Name)
	if err != nil {
		errChan <- err
	} else {
		okChan <- fmt.Sprintf("Success: %s - %s", service.Name, check.GetName())
	}
}
