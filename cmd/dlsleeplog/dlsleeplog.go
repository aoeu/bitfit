package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aoeu/bitfit"
)

func main() {
	fs := flag.NewFlagSet("wat", flag.ContinueOnError)
	args := struct {
		bitfit.Args
		from *string
		to   *string
		as   *string
		in   *string
	}{
		bitfit.ArgsWithFlagSet(fs),
		fs.String("from", "", "the date download a sleep log from"),
		fs.String("to", "", "the date to download a range of sleep logs until (inclusive"),
		fs.String("as", "sleep_log_payload", "the filename template to use for saved payloads"),
		fs.String("in", ".", "the path in which to write files of payloads to save"),
	}

	if err := bitfit.ParseFlagSet(fs); err != nil {
		log.Fatal(err)
	}
	if err := args.Validate(); err != nil {
		log.Fatal(err)
	}

	p, err := filepath.Abs(*args.in)
	if err != nil {
		log.Fatalf("could not get absolute path of %v: %v", *args.in, err)
	}
	*args.in = p

	if *args.from == "" {
		fmt.Fprintf(os.Stderr, "must provide a date in as the '-from' argument")
	}
	if *args.to == "" {
		args.to = args.from
	}

	layout := "2006-01-02"
	from, err := time.Parse(layout, *args.from)
	if err != nil {
		log.Fatal(err)
	}
	// TODO(aoeu): Check if the date is within the past 30 days, which the API requires.
	to, err := time.Parse(layout, *args.to)
	if err != nil {
		log.Fatal(err)
	}

	if err := bitfit.Init(*args.ClientID, *args.Secret, *args.TokensFilepath); err != nil {
		log.Fatal(err)
	}

	for i := 0; i <= int(to.Sub(from).Hours()/24); i++ {
		t := from.AddDate(0, 0, i)
		b, err := bitfit.FetchSleepLog(t)
		if err != nil {
			log.Fatal(err)
		}
		s := fmt.Sprintf("%v/%v_%v.json", *args.in, *args.as, t.Format(layout))
		if err := ioutil.WriteFile(s, b, 0644); err != nil {
			err = fmt.Errorf("could not write to file '%v': %v", s, err)
			log.Fatal(err)
		}
	}
}
