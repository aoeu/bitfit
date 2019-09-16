package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func main() {
	args := struct {
		clientID      string
		secret        string
		refreshToken  string
		printFullResp bool
	}{}
	flag.BoolVar(&args.printFullResp, "fullresp", false, "pretty-print all respone data")
	flag.StringVar(&args.clientID, "id", "", "the OAuth2 API client ID")
	flag.StringVar(&args.secret, "secret", "", "the OAuth2 API client secret")
	flag.StringVar(&args.refreshToken, "refreshtoken", "", "a refresh token previously obtained via the fitbit API (or web dashboard)")
	flag.Parse()
	if args.clientID == "" || args.secret == "" || args.refreshToken == "" {
		flag.Usage()
	}

	s := fmt.Sprintf("%v:%v", args.clientID, args.secret)
	s = fmt.Sprintf("Basic %v", base64.StdEncoding.EncodeToString([]byte(s)))
	payload := fmt.Sprintf("grant_type=refresh_token&refresh_token=%v", args.refreshToken)
	url := "https://api.fitbit.com/oauth2/token"

	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", s)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if args.printFullResp {
		prettyPrint(b)
		return
	}
	p := struct {
		Access_token string
	}{}
	if err := json.Unmarshal(b, &p); err != nil {
		log.Fatal(err)
	}
	fmt.Println(p.Access_token)
}

func prettyPrint(payload []byte) {
	b := new(bytes.Buffer)
	if err := json.Indent(b, payload, "", "    "); err != nil {
		fmt.Printf("could not debug print '%v' due to error: %v", string(payload), err)
	} else {
		fmt.Printf("%s\n", b)
	}
}
