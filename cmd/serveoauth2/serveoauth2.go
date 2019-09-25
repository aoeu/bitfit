package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/fcgi"
	"os"
	"text/template"
)

var tmplText = `
<html>
	<head>
		<title>Oauth2 code response</title>
	</head>
	<body>
		<center>
			<h1>{{.}}</h1>
		</center>
	</body>
</html>
`

func handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s := "could not parse form of request to URL '%v': %v"
		fmt.Fprintf(os.Stderr, s, r.URL, err)
		return
	}
	code := r.Form.Get("code")
	if code == "" {
		s := "A filed named 'code' was not present in request body or query parameters"
		writeResp(w, http.StatusBadRequest, s)
		return
	}
	tmpl, err := template.New("").Parse(tmplText)
	if err != nil {
		s := "could not parse template text as HTML template"
		writeResp(w, http.StatusInternalServerError, s)
		return
	}
	if err := tmpl.Execute(w, code); err != nil {
		s := fmt.Sprintf("could not execute HTML template code: %v", code)
		writeResp(w, http.StatusInternalServerError, s)
		return
	}
}

func writeResp(w http.ResponseWriter, code int, message string) {
	log.Println(message)
	w.WriteHeader(code)
	if _, err := w.Write([]byte(message)); err != nil {
		log.Printf("could not write '%v' as response with HTTP Status '%v': %v", message, code, err)
	}
}

func main() {
	args := struct {
		port string
		fcgi bool
	}{}
	flag.StringVar(&args.port, "port", ":8080", "The port to serve on.")
	flag.BoolVar(&args.fcgi, "cgi", true, "Serve HTTP via FastCGI")
	flag.Parse()
	http.HandleFunc("/", handleOAuth2Callback)
	switch {
	case args.fcgi:
		if err := fcgi.Serve(nil, nil); err != nil {
			log.Fatal(err)
		}
	default:
		if err := http.ListenAndServe(args.port, nil); err != nil {
			log.Fatal(err)
		}
	}
}
