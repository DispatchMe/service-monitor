package checks

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DispatchMe/service-monitor/shared"
	"github.com/maxwellhealth/go-floatlist"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type GraphiteCheck struct {
	serviceName string
	Name        string
	Metric      string
	Comparator  string
	Operator    string
	From        string
	Value       float64
	Whitelist   []string
	Blacklist   []string
	parsedLists bool
}

func strInSlice(slc []string, val string) bool {
	for _, s := range slc {
		if s == val {
			return true
		}
	}
	return false
}

type NullOrFloat float64

func (n *NullOrFloat) UnmarshalJSON(data []byte) error {
	// Try converting it to a float, otherwise throw an error.
	str := string(data)

	if str == "null" {
		*n = NullOrFloat(0)
	} else {
		parsed, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}

		*n = NullOrFloat(parsed)
	}
	return nil
}

type GraphiteResponse struct {
	Target     string          `json:"target"`
	DataPoints [][]NullOrFloat `json:"datapoints"`
}

func (g *GraphiteCheck) parseLists() {
	defer func() {
		g.parsedLists = true
	}()

	metricName := g.Metric
	if !strings.Contains(metricName, "*") {
		return
	}

	for i, s := range g.Whitelist {
		g.Whitelist[i] = strings.Replace(metricName, "*", s, 1)
	}

	for i, s := range g.Blacklist {
		g.Blacklist[i] = strings.Replace(metricName, "*", s, 1)
	}

}
func (g *GraphiteCheck) handleMetric(response *GraphiteResponse) error {
	flist := floatlist.Floatlist{}

	for _, p := range response.DataPoints {
		flist = append(flist, float64(p[0]))
	}

	// Get the operator value:
	var aggregateVal float64
	switch g.Operator {
	case "mean":
		aggregateVal = flist.Mean()
	case "median":
		aggregateVal = flist.Median()
	case "mode":
		aggregateVal = flist.Mode()
	case "sum":
		aggregateVal = flist.Sum()

	}

	var failed bool
	switch g.Comparator {
	case "==":
		failed = (aggregateVal == g.Value)
	case ">":
		failed = (aggregateVal > g.Value)
	case ">=":
		failed = (aggregateVal >= g.Value)
	case "<":
		failed = (aggregateVal < g.Value)
	case "<=":
		failed = (aggregateVal <= g.Value)
	case "!=":
		failed = (aggregateVal != g.Value)
	}

	if failed {
		return errors.New(fmt.Sprintf("[%s - %s] failed: %s(%s) %s %.2f (got %.2f)", g.serviceName, g.Name, g.Operator, response.Target, g.Comparator, g.Value, aggregateVal))
	}
	return nil
}

func (g *GraphiteCheck) Run(serviceName string) error {
	if !g.parsedLists {
		g.parseLists()
	}

	g.serviceName = serviceName

	// Make the query string
	query := url.Values{}

	query.Add("target", g.Metric)
	query.Add("from", g.From)
	query.Add("format", "json")

	response := make([]*GraphiteResponse, 0)
	req, err := http.NewRequest("GET", shared.GraphiteConf.Url+"render?"+query.Encode(), nil)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s - %s] %s", serviceName, g.Name, err.Error()))
	}
	req.Close = true
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s - %s] %s", serviceName, g.Name, err.Error()))
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&response)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s - %s] %s", serviceName, g.Name, err.Error()))
	}

	if len(response) == 0 {
		return errors.New(fmt.Sprintf("[%s - %s] No metric", serviceName, g.Name))
	}

	errors := []error{}
	for _, metric := range response {
		// Check white/black list if there's an asterisk
		if strInSlice(g.Blacklist, metric.Target) {
			continue
		}

		if len(g.Whitelist) > 0 && !strInSlice(g.Whitelist, metric.Target) {
			continue
		}

		err := g.handleMetric(metric)
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return &shared.MultiError{errors}
	}

	return nil

}

func (g *GraphiteCheck) GetName() string {
	return g.Name
}
