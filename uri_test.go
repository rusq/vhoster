package vhoster

import (
	"encoding/json"
	"net/url"
	"testing"
)

type teststruct struct {
	URI *URI `json:"uri"`
}

const sampleJSON = `{"uri":"http://example.com"}`

func TestURI_UnmarshalJSON(t *testing.T) {
	js := sampleJSON
	var ts teststruct
	err := json.Unmarshal([]byte(js), &ts)
	if err != nil {
		t.Error(err)
	}
	if ts.URI.String() != "http://example.com" {
		t.Errorf("URI is not http://example.com")
	}
}

func TestURI_MarshalJSON(t *testing.T) {
	uri, err := url.Parse("http://example.com")
	if err != nil {
		t.Error(err)
	}
	ts := teststruct{URI: ToURI(uri)}
	b, err := json.Marshal(ts)
	if err != nil {
		t.Error(err)
	}
	if string(b) != sampleJSON {
		t.Errorf("json.Marshal failed: %s != %s", string(b), sampleJSON)
	}
}
