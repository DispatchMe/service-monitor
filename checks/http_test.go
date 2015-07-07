package checks

import (
	"testing"
)

func TestHttpCheck(t *testing.T) {
	check := &HttpCheck{
		Url:  "http://google.com",
		Code: 201,
	}

	err := check.Run()
	if err == nil {
		t.Error("Expected error")
	}

	check.Code = 200
	err = check.Run()
	if err != nil {
		t.Errorf("Expected no error, got %s", err.Error())
	}
}
