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
	c := bitfit.NewClient(*args.ClientID, *args.Secret, *args.TokensFilepath)
	if err := c.Init(); err != nil {
		log.Fatal(err)
	}
	b, err := c.GetProfile()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}