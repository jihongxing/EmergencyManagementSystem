package platform

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestContractResponses exposes real handler responses to the OpenAPI validator.
func TestContractResponses(t *testing.T) {
	for _, scenario := range []struct {
		path  string
		check func(context.Context) error
	}{
		{path: "/health/live"},
		{path: "/health/ready", check: func(context.Context) error { return nil }},
		{path: "/health/ready", check: func(context.Context) error { return errors.New("unavailable") }},
	} {
		recorder := httptest.NewRecorder()
		NewHandler(scenario.check).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, scenario.path, nil))
		result := struct {
			Path        string          `json:"path"`
			Status      int             `json:"status"`
			ContentType string          `json:"contentType"`
			Body        json.RawMessage `json:"body"`
		}{
			Path: scenario.path, Status: recorder.Code,
			ContentType: recorder.Header().Get("Content-Type"),
			Body:        json.RawMessage(recorder.Body.Bytes()),
		}
		data, err := json.Marshal(result)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("CONTRACT_RESPONSE %s", data)
	}
}
