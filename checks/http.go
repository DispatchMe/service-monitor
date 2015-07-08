package checks

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type HttpCheck struct {
	Name      string
	Url       string
	Code      int
	MatchBody string
	Timeout   int
}

func (h *HttpCheck) GetName() string {
	return h.Name
}

type transportDial func(string, string) (net.Conn, error)

func dialTimeout(timeout time.Duration) transportDial {
	return func(network, addr string) (net.Conn, error) {
		return net.DialTimeout(network, addr, timeout)
	}
}

func (h *HttpCheck) Run(serviceName string) error {

	if h.Timeout == 0 {
		h.Timeout = 60
	}

	transport := http.Transport{
		Dial: dialTimeout(time.Duration(h.Timeout) * time.Second),
	}

	client := http.Client{
		Transport: &transport,
	}

	req, err := http.NewRequest("GET", h.Url, nil)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s - %s] %s", serviceName, h.Name, err.Error()))
	}
	req.Close = true
	resp, err := client.Do(req)

	if err != nil {
		return errors.New(fmt.Sprintf("[%s - %s] %s", serviceName, h.Name, err.Error()))
	}

	if resp.StatusCode != h.Code {
		buf := &bytes.Buffer{}
		buf.ReadFrom(resp.Body)
		log.Println(buf.String())
		return errors.New(fmt.Sprintf("[%s - %s] Invalid status code: got %d, expecting %d", serviceName, h.Name, resp.StatusCode, h.Code))
	}

	// Do a body check if provided
	if len(h.MatchBody) > 0 {
		buf := &bytes.Buffer{}
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			return errors.New(fmt.Sprintf("[%s - %s] %s", serviceName, h.Name, err.Error()))
		}

		if !strings.Contains(buf.String(), h.MatchBody) {
			return errors.New(fmt.Sprintf("[%s - %s] Invalid body: expecting %s but it was not present in body", serviceName, h.Name, h.MatchBody))
		}
	}

	return nil

}
