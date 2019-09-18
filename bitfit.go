package bitfit

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

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
	return b, nil
}
