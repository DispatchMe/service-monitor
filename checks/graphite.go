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
)

type GraphiteCheck struct {
	Name       string
	Metric     string
	Comparator string
	Operator   string
	From       string
	Value      float64
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

func (g *GraphiteCheck) Run(serviceName string) error {

	// Make the query string
	query := url.Values{}

	query.Add("target", g.Metric)
	query.Add("from", g.From)
	query.Add("format", "json")

	response := make([]*GraphiteResponse, 0)
	resp, err := http.Get(shared.GraphiteConf.Url + "render?" + query.Encode())

	if err != nil {
		return err
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&response)
	if err != nil {
		return err
	}

	flist := floatlist.Floatlist{}

	if len(response) == 0 {
		return errors.New(fmt.Sprintf("[%s - %s] No metric", serviceName, g.Name))
	}

	for _, p := range response[0].DataPoints {
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
		return errors.New(fmt.Sprintf("[%s - %s] failed: %s(%s) %s %.2f (got %.2f)", serviceName, g.Name, g.Operator, g.Metric, g.Comparator, g.Value, aggregateVal))
	}
	return nil
}

func (g *GraphiteCheck) GetName() string {
	return g.Name
}
