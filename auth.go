package bitfit

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func FetchTokens(id, secret, refreshToken string) (respBody []byte, err error) {
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
	return b, nil
}

func FetchAuthToken(id, secret, refreshToken string) (string, error) {
	b, err := FetchTokens(id, secret, refreshToken)
	if err != nil {
		return "", err
	}
	p := struct {
		Access_token string
	}{}
	if err := json.Unmarshal(b, &p); err != nil {
		return "", err
	}
	return p.Access_token, nil
}