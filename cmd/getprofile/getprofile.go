package main

import (
	"fmt"
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
	if err := bitfit.Init(*args.ClientID, *args.Secret, *args.TokensFilepath); err != nil {
		log.Fatal(err)
	}
	b, err := bitfit.Profile()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
