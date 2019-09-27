package bitfit

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"
)

func TestUnmarshallingTokens(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/tokens_payload.json")
	if err != nil {
		t.Fatal(err)
	}
	tt := new(Tokens)
	if err := json.Unmarshal(b, &tt); err != nil {
		t.Fatal(err)
	}
	d, err := time.ParseDuration("8h")
	if err != nil {
		t.Fatal(err)
	}
	f, s := "Mon Jan 2 15:04:05 -0700 MST 2006", "expected %v but received %v"
	if e, a := time.Now().Add(d).Format(f), tt.Expiration.Format(f); e != a {
		t.Fatalf(s, e, a)
	}
	if e, a := "foo", tt.Access; e != a {
		t.Fatalf(s, e, a)
	}
	if e, a := "bar", tt.Refresh; e != a {
		t.Fatal(s, e, a)
	}
}

func TestUnmarshallingSleepLog(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/sleep_log_payload_20190916.json")
	if err != nil {
		t.Fatal(err)
	}
	s := new(SleepLog)
	if err := json.Unmarshal(b, s); err != nil {
		t.Fatal(err)
	}
	e, err := time.ParseDuration("59m")
	if err != nil {
		t.Fatal(err)
	}
	a := s.Summary.DurationPerStage.Deep
	if e != a {
		t.Fatalf("expected %+v but received %+v", e, a)
	}
	e, err = time.ParseDuration("359m")
	if err != nil {
		t.Fatal(err)
	}
	a = s.Summary.DurationAsleep
	if e != a {
		t.Fatalf("expected %+v but received %+v", e, a)
	}
	e, err = time.ParseDuration("421m")
	if err != nil {
		t.Fatal(err)
	}
	a = s.Summary.DurationInBed
	if e != a {
		t.Fatalf("expected %+v but received %+v", e, a)
	}
	if e, a := uint(1), s.Summary.NumSleepRecords; e != a {
		t.Fatalf("expected %+v but received %+v", e, a)
	}
}
