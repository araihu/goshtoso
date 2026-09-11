// Package deployments holds the simulated deployment console's state.
package deployments

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const MaxRecords = 8

type Record struct {
	ID                                    int
	Service, Version, Environment, Status string
}
type State struct{ Records []Record }

func Initial() State {
	return State{Records: []Record{
		{1, "payments-api", "v2.4.0", "Production", "Pending review"},
		{2, "identity-worker", "v1.8.2", "Staging", "Approved"},
		{3, "docs-site", "v3.1.0", "Production", "Approved"},
	}}
}

func Validate(service, version, environment string) map[string]string {
	errors := map[string]string{}
	if len(strings.TrimSpace(service)) < 2 || len(service) > 60 {
		errors["service"] = "Enter a service name between 2 and 60 characters."
	}
	if strings.TrimSpace(version) == "" || len(version) > 30 {
		errors["version"] = "Enter a version of up to 30 characters."
	}
	if environment != "Staging" && environment != "Production" {
		errors["environment"] = "Choose Staging or Production."
	}
	return errors
}

func Decode(value string) State {
	if value == "" || len(value) > 5000 {
		return Initial()
	}
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return Initial()
	}
	var s State
	if json.Unmarshal(data, &s) != nil || len(s.Records) == 0 || len(s.Records) > MaxRecords {
		return Initial()
	}
	ids := map[int]bool{}
	for _, r := range s.Records {
		if r.ID < 1 || r.ID > 100000 || ids[r.ID] || len(Validate(r.Service, r.Version, r.Environment)) > 0 || (r.Status != "Pending review" && r.Status != "Approved") {
			return Initial()
		}
		ids[r.ID] = true
	}
	return s
}

func (s State) Encode() string {
	data, _ := json.Marshal(s)
	return base64.RawURLEncoding.EncodeToString(data)
}
func (s State) URL(view string, id int) string {
	q := url.Values{"view": {view}, "state": {s.Encode()}}
	if id != 0 {
		q.Set("id", fmt.Sprint(id))
	}
	return "/examples/deployments?" + q.Encode()
}
func (s State) Find(id int) (Record, bool) {
	for _, r := range s.Records {
		if r.ID == id {
			return r, true
		}
	}
	return Record{}, false
}
func (s State) Filter(query string) []Record {
	var records []Record
	for _, r := range s.Records {
		if strings.Contains(strings.ToLower(r.Service+" "+r.Version+" "+r.Environment), strings.ToLower(strings.TrimSpace(query))) {
			records = append(records, r)
		}
	}
	return records
}
func (s *State) Create(service, version, environment string) int {
	id := 1
	for _, r := range s.Records {
		if r.ID >= id {
			id = r.ID + 1
		}
	}
	s.Records = append([]Record{{id, strings.TrimSpace(service), strings.TrimSpace(version), environment, "Pending review"}}, s.Records...)
	if len(s.Records) > MaxRecords {
		s.Records = s.Records[:MaxRecords]
	}
	return id
}
func (s *State) Approve(id int) {
	for i := range s.Records {
		if s.Records[i].ID == id {
			s.Records[i].Status = "Approved"
		}
	}
}
