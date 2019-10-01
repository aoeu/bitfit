package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/aoeu/bitfit"
)

func main() {
	args, err := bitfit.ParseFlags(os.Args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if err := bitfit.Init(*args.ClientID, *args.Secret, *args.TokensFilepath); err != nil {
		log.Fatal(err)
	}
	b, err := bitfit.FetchSleepLog(time.Now().AddDate(0, 0, -1))
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile("sleep_log_payload.json", b, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
