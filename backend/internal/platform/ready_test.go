package platform

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestReadyContract(t *testing.T) {
	data, err := os.ReadFile("../../../constras/platform/ready.openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract struct {
		Paths map[string]struct {
			Get struct {
				Responses map[string]struct {
					Content map[string]struct {
						Schema struct {
							Properties map[string]struct{ Const string }
						}
					}
				}
			}
		}
	}
	if err := json.Unmarshal(data, &contract); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		code    int
		key     string
		failure error
	}{
		{200, "200", nil}, {503, "503", errors.New("private database detail")},
	} {
		check := func(ctx context.Context) error {
			if _, ok := ctx.Deadline(); !ok {
				t.Fatal("missing readiness timeout")
			}
			return tc.failure
		}
		resp := httptest.NewRecorder()
		NewHandler(check).ServeHTTP(resp, httptest.NewRequest("GET", "/health/ready", nil))
		expected := contract.Paths["/health/ready"].Get.Responses[tc.key].Content["application/json"].Schema.Properties["status"].Const
		var body map[string]string
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil || expected == "" || body["status"] != expected ||
			len(body) != 1 || resp.Code != tc.code || resp.Header().Get("Content-Type") != "application/json; charset=utf-8" {
			t.Fatalf("contract mismatch: %d %s", resp.Code, resp.Body.String())
		}
		live := httptest.NewRecorder()
		NewHandler(check).ServeHTTP(live, httptest.NewRequest(http.MethodGet, "/health/live", nil))
		if live.Code != 200 {
			t.Fatal("liveness depends on database")
		}
	}
}
