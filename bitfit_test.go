package bitfit

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"
)

func TestUnmarshallingTokensPayload(t *testing.T) {
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

func TestUnmarshallingTokens(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/tokens.json")
	if err != nil {
		t.Fatal(err)
	}
	tt := new(Tokens)
	if err := json.Unmarshal(b, &tt); err != nil {
		t.Fatal(err)
	}

	f, s := "Mon Jan 2 15:04:05 -0700 MST 2006", "expected %v but received %v"
	expiration, err := time.Parse(f, "Tue Oct 1 08:28:25 -0400 EDT 2019")
	if err != nil {
		t.Fatal(err)
	}
	if e, a := expiration.Format(f), tt.Expiration.Format(f); e != a {
		t.Fatalf(s, e, a)
	}
	if e, a := "foo", tt.Access; e != a {
		t.Fatalf(s, e, a)
	}
	if e, a := "bar", tt.Refresh; e != a {
		t.Fatal(s, e, a)
	}
}

func TestUnmarshallingSleepSummary(t *testing.T) {
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

func TestUnmarshalSleepSession(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/sleep_log_payload_20190916.json")
	if err != nil {
		t.Fatal(err)
	}
	s := new(SleepLog)
	if err := json.Unmarshal(b, s); err != nil {
		t.Fatal(err)
	}
	if e, a := 1, len(s.Sessions); e != a {
		t.Fatalf("expected %v session(s) but there was %v", e, a)
	}
	var (
		newYork *time.Location
		f       = "2006-01-02T15:04:05.000"
		tt      time.Time
		d       time.Duration
		errFmt  = "expected %v but received %v"
	)
	if newYork, err = time.LoadLocation("America/New_York"); err != nil {
		t.Fatal(err)
	}
	if tt, err = time.ParseInLocation(f, "2019-09-16T09:09:30.000", newYork); err != nil {
		t.Fatal(err)
	}
	sess := s.Sessions[0]
	if e, a := tt.String(), sess.End.String(); e != a {
		t.Fatalf(errFmt, e, a)
	}
	if e, a := true, sess.IsPrimary; e != a {
		t.Fatalf(errFmt, e, a)
	}
	if d, err = time.ParseDuration("25260000s"); err != nil {
		t.Fatal(err)
	}
	if e, a := d, sess.Length; e != a {
		t.Fatalf(errFmt, e, a)
	}
	if e, a := 52, len(sess.Observations); e != a {
		t.Fatalf(errFmt, e, a)
	}
	for i := len(sess.Observations) - 1; i > 1; i-- {
		m, n := sess.Observations[i].Start, sess.Observations[i-1].Start
		if m.Before(n) {
			s := "expected '%v' (index %v) to be earlier in time than '%v' (index '%v')"
			t.Fatalf(s, n, i-1, m, i)
		}
	}
	o := sess.Observations[0]
	if tt, err = time.ParseInLocation(f, "2019-09-16T02:08:00.000", newYork); err != nil {
		t.Fatal(err)
	}
	if e, a := tt.String(), o.Start.String(); e != a {
		t.Fatalf(errFmt, e, a)
	}
	if e, a := 0.4, o.Duration.Hours(); e != a {
		t.Fatalf(errFmt, e, a)
	}
	// TODO(aoeu): Typed constants and test for Type field.
}