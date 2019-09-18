package main

import (
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
	t, err := bitfit.FetchTokensPayload(*args.ClientID, *args.Secret, *args.RefreshToken)
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile("tokens_payload.json", t, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(t))
}
