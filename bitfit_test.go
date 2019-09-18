package bitfit

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"
)

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
