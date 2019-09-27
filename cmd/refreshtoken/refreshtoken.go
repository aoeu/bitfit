package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aoeu/bitfit"
)

func main() {
	args, err := bitfit.ParseFlags(os.Args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	t, err := bitfit.FetchTokens(*args.ClientID, *args.Secret, *args.RefreshToken)
	if err != nil {
		log.Fatal(err)
	}
	b, err := json.Marshal(t)
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile("tokens.json", b, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", t)
}
