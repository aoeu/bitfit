package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/fcgi"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/aoeu/bitfit"
)

var (
	apiProxy *httputil.ReverseProxy
	username string
	password string
)

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	args := struct {
		bitfit.Args
		username     *string
		password     *string
		certFilepath *string
		keyFilepath  *string
		port         *string
		useFCGI      *bool
	}{
		bitfit.ArgsWithFlagSet(fs),
		fs.String("username", "", "A username required on client requests for HTTP basic auth, as per RFC 7617"),
		fs.String("password", "", "A password required on client requests for HTTP basic auth, as per RFC 7617"),
		fs.String("certfile", "cert.txt", "a cert to use for the proxy server for TLS / HTTPS"),
		fs.String("keyfile", "key.txt", "a key to use for the proxy server for TLS / HTTPS"),
		fs.String("port", ":9090", "The port to serve on."),
		fs.Bool("cgi", false, "Serve HTTP via FastCGI"),
	}
	if err := bitfit.ParseFlagSet(fs); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("'%v'\n", *args.ClientID)
	if err := args.Validate(); err != nil {
		log.Fatal(err)
	}
	if err := bitfit.Init(*args.ClientID, *args.Secret, *args.TokensFilepath); err != nil {
		log.Fatal(err)
	}
	username, password = *args.username, *args.password

	u, err := url.Parse(bitfit.BaseURL)
	if err != nil {
		log.Fatal(err)
	}

	apiProxy = httputil.NewSingleHostReverseProxy(u)
	apiProxy.Transport = bitfit.DefaultClient
	d := apiProxy.Director
	apiProxy.Director = func(r *http.Request) {
		d(r)
		r.Host = u.Host
	}

	http.HandleFunc("/", handleAPICall)
	switch {
	case *args.useFCGI:
		if err := fcgi.Serve(nil, apiProxy); err != nil {
			log.Fatal(err)
		}
	default:
		p, c, k := *args.port, *args.certFilepath, *args.keyFilepath
		if err := http.ListenAndServeTLS(p, c, k, apiProxy); err != nil {
			log.Fatal(err)
		}
	}
}

func handleAPICall(w http.ResponseWriter, r *http.Request) {
	u, p, ok := r.BasicAuth()
	authErr := ""
	switch {
	case !ok:
		authErr = "basic HTTP authentication is required (RFC 7617)"
	case u == "":
		authErr = "username in basic authentication is required (RFC 7617)"
	case p == "":
		authErr = "password in basic authentication is required (RFC 7617)"
	case u != username || p != password:
		authErr = "incorrect username or password"
	}
	if authErr != "" {
		writeResp(w, http.StatusUnauthorized, authErr)
		return
	}
	apiProxy.ServeHTTP(w, r)
}

func writeResp(w http.ResponseWriter, code int, message string) {
	log.Println(message)
	w.WriteHeader(code)
	if _, err := w.Write([]byte(message)); err != nil {
		log.Printf("could not write '%v' as response with HTTP Status '%v': %v", message, code, err)
	}
}
