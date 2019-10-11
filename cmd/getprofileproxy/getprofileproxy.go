package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aoeu/bitfit"
)

func main() {
	args, err := bitfit.ParseProxyFlags(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}
	switch "" {
	case *args.BaseURL, *args.Username, *args.Password:
		flag.Usage()
		os.Exit(1)
	}
	if err := bitfit.InitProxy(*args.BaseURL, *args.Username, *args.Password); err != nil {
		log.Fatal(err)
	}
	b, err := bitfit.FetchProfile()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
