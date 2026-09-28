package identity

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestAuthorizationContract(t *testing.T) {
	data, err := os.ReadFile("../../../constras/identity/authorization.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract struct {
		Version int
		Clauses []string
		Cases   []struct {
			Name     string
			Member   Member
			Target   string
			Expected string
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&contract); err != nil {
		t.Fatal(err)
	}
	if contract.Version != 1 || len(contract.Cases) == 0 {
		t.Fatal("invalid identity contract")
	}
	for _, scenario := range contract.Cases {
		t.Run(scenario.Name, func(t *testing.T) {
			if actual := CanManageMembers(scenario.Member, scenario.Target); actual != scenario.Expected {
				t.Fatalf("got %q, want %q", actual, scenario.Expected)
			}
		})
	}
}

func TestRevocationUsesCurrentState(t *testing.T) {
	member := Member{"u1", "o1", "enterprise", true, []string{"enterprise_admin"}}
	if CanManageMembers(member, "o1") != "allowed" {
		t.Fatal("active administrator denied")
	}
	member.Active = false
	if CanManageMembers(member, "o1") != "inactive" {
		t.Fatal("revoked member allowed")
	}
}
