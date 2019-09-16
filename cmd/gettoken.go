package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aoeu/bitfit"
	"github.com/peterbourgon/ff"
)

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	args := struct {
		clientID      *string
		secret        *string
		refreshToken  *string
		printFullResp *bool
	}{
		fs.String("id", "", "the OAuth2 API client ID"),
		fs.String("secret", "", "the OAuth2 API client secret"),
		fs.String("refreshtoken", "", "a refresh token previously obtained via the fitbit API (or web dashboard)"),
		fs.Bool("fullresp", false, "pretty-print all respone data"),
	}
	_ = fs.String("config", "", "config file (optional)")
	ff.Parse(fs, os.Args[1:],
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithEnvVarPrefix("BIT_FIT"),
	)
	switch {
	case *args.clientID == "", *args.secret == "", *args.refreshToken == "":
		flag.Usage()
	}

	if *args.printFullResp {
		b, err := bitfit.FetchTokens(*args.clientID, *args.secret, *args.refreshToken)
		if err != nil {
			log.Fatal(err)
		}
		prettyPrint(b)
		return
	}

	t, err := bitfit.FetchAuthToken(*args.clientID, *args.secret, *args.refreshToken)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(t)
}

func prettyPrint(payload []byte) {
	b := new(bytes.Buffer)
	if err := json.Indent(b, payload, "", "    "); err != nil {
		fmt.Printf("could not debug print '%v' due to error: %v", string(payload), err)
	} else {
		fmt.Printf("%s\n", b)
	}
}
