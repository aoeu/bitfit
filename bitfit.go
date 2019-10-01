package bitfit

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
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
	ClientID       *string
	Secret         *string
	RefreshToken   *string // Deprecated
	TokensFilepath *string
}

func ParseFlags(name string) (Args, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	a := Args{
		fs.String("id", "", "the OAuth2 API client ID"),
		fs.String("secret", "", "the OAuth2 API client secret"),
		fs.String("refreshtoken", "", "a refresh token previously obtained via the fitbit API (or web dashboard)"),
		fs.String("tokensfile", "", "a JSON file of access and refresh tokens previous obtained via fitbit API and serialized via the bitfit library"),
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
	case *a.RefreshToken == "" && *a.TokensFilepath == "":
		err = fmt.Errorf("no refresh token or tokens filepath provided\n")
	}
	return a, err
}

type Tokens struct {
	Access     string
	Refresh    string
	Expiration time.Time
}

// UnmarshalJSON accepts a JSON payload from the REST API
// xor a marshalled JSON payload of this package's Tokens type and
// deserializes the fields of either into a Tokens instance.
func (t *Tokens) UnmarshalJSON(data []byte) error {
	s := struct {
		// Fitbit API response fields
		Access_token  string
		Refresh_token string
		Expires_in    uint
		Errors        []map[string]string
		// Tokens fields
		Access     string
		Refresh    string
		Expiration time.Time
	}{}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	// Token fields are absent when error fields are present, and vice-versa.
	if len(s.Errors) > 0 {
		m, ok := s.Errors[0]["message"]
		if !ok {
			return fmt.Errorf("%+v", s.Errors[0])
		}
		return errors.New(m)
	}
	switch {
	case s.Access != "":
		t.Access = s.Access
	case s.Access_token != "":
		t.Access = s.Access_token
	}
	switch {
	case s.Refresh != "":
		t.Refresh = s.Refresh
	case s.Refresh_token != "":
		t.Refresh = s.Refresh_token
	}
	switch {
	case s.Expiration != time.Time{}:
		t.Expiration = s.Expiration
	case s.Expires_in != 0:
		d, err := parseSec(s.Expires_in)
		if err != nil {
			ss := "could not parse expiration seconds %v to time: %v"
			return fmt.Errorf(ss, s.Expires_in, err)
		}
		t.Expiration = time.Now().Add(d)
	}
	return nil
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
	t := Tokens{}
	if err := json.Unmarshal(b, &t); err != nil {
		return t, err
	}
	return t, nil
}

type Client struct {
	*http.Client
	Tokens
	id             string
	secret         string
	tokensFilepath string
}

func NewClient(id, secret, tokensFilepath string) *Client {
	c := &Client{
		id:             id,
		secret:         secret,
		tokensFilepath: tokensFilepath,
	}
	// Effectively duplicate http.DefaultClient but with an overridden Transport.RoundTrip func.
	c.Client = &http.Client{
		Transport: c,
	}
	return c
}

func (c *Client) Init() error {
	if c.tokensFilepath == "" {
		s := "filepath of an existing token (serialized as JSON) must be set on Client"
		return fmt.Errorf(s)
	}
	b, err := ioutil.ReadFile(c.tokensFilepath)
	if err != nil {
		s := "filepath of tokens '%v' could not be read: %v"
		return fmt.Errorf(s, c.tokensFilepath, err)
	}
	var t Tokens
	if err := json.Unmarshal(b, &t); err != nil {
		s := "could not unmarshal tokens at filepath '%v': %v"
		return fmt.Errorf(s, c.tokensFilepath, err)
	}
	c.Tokens = t
	if c.Expiration.Before(time.Now()) {
		if err := c.refreshTokens(); err != nil {
			s := "could not refresh expired tokens loaded from '%v' during Init func: %v"
			return fmt.Errorf(s, err)
		}
	}
	return nil
}

func (c *Client) RoundTrip(req *http.Request) (*http.Response, error) {
	if c.Expiration.Before(time.Now()) {
		if err := c.refreshTokens(); err != nil {
			s := "could not refresh expired tokens before request in round trip function: %v"
			return nil, fmt.Errorf(s, err)
		}
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", c.Access))
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err == nil {
		err = c.refreshTokens()
		if err != nil {
			s := "could not refresh tokens after request in round trip function: %v"
			return nil, fmt.Errorf(s, err)
		}
	}
	return resp, err
}

func (c *Client) refreshTokens() error {
	t, err := FetchTokens(c.id, c.secret, c.Refresh)
	if err != nil {
		return err
	}
	c.Tokens = t
	return c.saveTokens()
}

func (c *Client) saveTokens() error {
	b, err := json.Marshal(c.Tokens)
	if err != nil {
		return fmt.Errorf("could not serialize tokens: %v", err)
	}
	b, err = format(b)
	if err != nil {
		return fmt.Errorf("could not format serialized tokens: %v", err)
	}
	if err := ioutil.WriteFile(c.tokensFilepath, b, 0644); err != nil {
		s := "could not save tokens to file '%v': %v"
		return fmt.Errorf(s, c.tokensFilepath, err)
	}
	return nil
}

func (c *Client) GetProfile() (respBody []byte, err error) {
	url := "https://api.fitbit.com/1/user/-/profile.json"
	resp, err := c.Get(url)
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

func parseSec(i uint) (time.Duration, error) {
	return time.ParseDuration(fmt.Sprintf("%vs", i))
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
