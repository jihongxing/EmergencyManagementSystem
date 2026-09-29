package platform

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLiveContract(t *testing.T) {
	data, err := os.ReadFile("../../../constras/platform/live.openapi.json")
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
	expected := contract.Paths["/health/live"].Get.Responses["200"].Content["application/json"].Schema.Properties["status"].Const
	if expected == "" {
		t.Fatal("missing status in contract")
	}
	resp := httptest.NewRecorder()
	NewHandler().ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	var actual map[string]string
	if err := json.Unmarshal(resp.Body.Bytes(), &actual); err != nil {
		t.Fatal(err)
	}
	if actual["status"] != expected || len(actual) != 1 || resp.Code != http.StatusOK ||
		resp.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("contract mismatch: %#v", actual)
	}
}
func TestLiveProbe(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	resp := httptest.NewRecorder()
	NewHandler().ServeHTTP(resp, req)
	if resp.Code != http.StatusOK || resp.Body.String() != "{\"status\":\"alive\"}\n" {
		t.Fatalf("unexpected probe response: %d %q", resp.Code, resp.Body.String())
	}
}

func TestNoBusinessEndpointsYet(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/inspections", nil)
	resp := httptest.NewRecorder()
	NewHandler().ServeHTTP(resp, req)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("unexpected endpoint: %d", resp.Code)
	}
}
