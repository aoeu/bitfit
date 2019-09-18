package bitfit

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/peterbourgon/ff"
)

type Args struct {
	ClientID     *string
	Secret       *string
	RefreshToken *string
}

func ParseFlags(name string) (Args, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	a := Args{
		fs.String("id", "", "the OAuth2 API client ID"),
		fs.String("secret", "", "the OAuth2 API client secret"),
		fs.String("refreshtoken", "", "a refresh token previously obtained via the fitbit API (or web dashboard)"),
	}
	_ = fs.String("config", "", "config file (optional)")
	err := ff.Parse(fs, os.Args[1:],
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.JSONParser),
		ff.WithEnvVarPrefix("BIT_FIT"),
	)
	if err != nil {
		return a, err
	}
	switch {
	case *a.ClientID == "":
		err = fmt.Errorf("no client ID provided\n")
	case *a.Secret == "":
		err = fmt.Errorf("no client secret provided\n")
	case *a.RefreshToken == "":
		err = fmt.Errorf("no refresh token provided\n")
	}
	return a, err
}

type Tokens struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
}

func FetchTokensPayload(id, secret, refreshToken string) ([]byte, error) {
	s := fmt.Sprintf("%v:%v", id, secret)
	s = fmt.Sprintf("Basic %v", base64.StdEncoding.EncodeToString([]byte(s)))
	payload := fmt.Sprintf("grant_type=refresh_token&refresh_token=%v", refreshToken)
	url := "https://api.fitbit.com/oauth2/token"

	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return []byte{}, err
	}
	req.Header.Add("Authorization", s)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return format(b)
}

func format(payload []byte) ([]byte, error) {
	b := new(bytes.Buffer)
	if err := json.Indent(b, payload, "", "    "); err != nil {
		s := "could not debug print '%v' due to error: %v"
		return []byte{}, fmt.Errorf(s, string(payload), err)
	}
	return b.Bytes(), nil
}

func FetchTokens(id, secret, refreshToken string) (Tokens, error) {
	b, err := FetchTokensPayload(id, secret, refreshToken)
	if err != nil {
		return Tokens{}, err
	}
	u := struct {
		Access_token  string
		Refresh_token string
		Errors        []map[string]string
	}{}
	if err := json.Unmarshal(b, &u); err != nil {
		return Tokens{}, err
	}
	if len(u.Errors) > 0 {
		return Tokens{}, fmt.Errorf(u.Errors[0]["message"])
	}
	return Tokens{Access: u.Access_token, Refresh: u.Refresh_token}, nil
}

func FetchProfile(authToken string) (respBody []byte, err error) {
	url := "https://api.fitbit.com/1/user/-/profile.json"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", authToken))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return b, nil
}

func FetchSleepLog(authToken string, t time.Time) (respBody []byte, err error) {
	s := "https://api.fitbit.com/1.2/user/-/sleep/date/%v.json"
	url := fmt.Sprintf(s, t.Format("2006-01-02"))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", authToken))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return format(b)
}

type SleepLog struct {
	Summary *Summary `json:"summary"`
}

type Summary struct {
	DurationPerStage *DurationPerStage `json:"stages"`
	DurationAsleep   time.Duration     `json:"totalMinutesAsleep"`
	NumSleepRecords  uint              `json:"totalSleepRecords"`
	DurationInBed    time.Duration     `json:"totalTimeInBed"`
}

func (s *Summary) UnmarshalJSON(data []byte) error {
	ss := struct {
		Stages             *DurationPerStage
		TotalMinutesAsleep uint
		TotalSleepRecords  uint
		TotalTimeInBed     uint
	}{}
	if err := json.Unmarshal(data, &ss); err != nil {
		return err
	}
	s.DurationPerStage = ss.Stages
	s.NumSleepRecords = ss.TotalSleepRecords
	var err error
	if s.DurationAsleep, err = parseMin(ss.TotalMinutesAsleep); err != nil {
		return err
	}
	if s.DurationInBed, err = parseMin(ss.TotalTimeInBed); err != nil {
		return err
	}
	return nil
}

func parseMin(i uint) (time.Duration, error) {
	return time.ParseDuration(fmt.Sprintf("%vm", i))
}

type DurationPerStage struct {
	Deep  time.Duration `json:"deep"`
	Light time.Duration `json:"light"`
	REM   time.Duration `json:"rem"`
	Awake time.Duration `json:"wake"`
}

func (d *DurationPerStage) UnmarshalJSON(data []byte) error {
	s := struct {
		Deep  uint
		Light uint
		Rem   uint
		Wake  uint
	}{}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	var err error
	if d.Deep, err = parseMin(s.Deep); err != nil {
		return err
	}
	if d.Light, err = parseMin(s.Deep); err != nil {
		return err
	}
	if d.REM, err = parseMin(s.Rem); err != nil {
		return err
	}
	if d.Awake, err = parseMin(s.Wake); err != nil {
		return err
	}
	return nil
}
