package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aoeu/bitfit"
	"github.com/aoeu/bitfit/proxy"
)

func main() {
	args, err := proxy.ParseFlags(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}
	switch "" {
	case *args.BaseURL, *args.Username, *args.Password:
		flag.Usage()
		os.Exit(1)
	}
	if err := proxy.Init(*args.BaseURL, *args.Username, *args.Password); err != nil {
		log.Fatal(err)
	}
	b, err := bitfit.FetchProfile()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
