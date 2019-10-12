package proxy

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/aoeu/bitfit"
)

type Args struct {
	BaseURL  *string
	Username *string
	Password *string
}

// TODO(aoeu): Require caller to set config,
// i.e. `_ = fs.String("config", "", "config file (optional)")`
func ArgsWithFlagSet(fs *flag.FlagSet, configDefault string) Args {
	_ = fs.String(bitfit.ConfigFlagName, configDefault, "config file (optional)")
	s := " required on client requests for HTTP basic auth, as per RFC 7617"
	return Args{
		fs.String("url", "", "the base URL of the proxy server"),
		fs.String("username", "", "A username"+s),
		fs.String("password", "", "A password"+s),
	}
}

func (a Args) Validate() error {
	switch {
	case *a.BaseURL == "":
		return fmt.Errorf("no proxy URL\n")
	case *a.Username == "":
		return fmt.Errorf("no username provided\n")
	case *a.Password == "":
		return fmt.Errorf("no password provided\n")
	}
	return nil
}

func ParseFlags(name string) (Args, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	a := ArgsWithFlagSet(fs, "")
	err := bitfit.ParseFlagSet(fs)
	if err != nil {
		return a, err
	}
	return a, a.Validate()
}

// NewClient is a constructor for a bitfit."Client" that is configured to
// route authenticated requests through a reverse-proxy server (that handles
// OAuth2 authentication with the FitBit API on behalf of the client. See
// bitfit/cmd/serveproxy.go for the reveres-proxy server implementation. The
// username and password provided to the to the NewClient constructor must
// match that which was used to configure (start) the bitfit/cmd/serveproxy.go
// instance that will be communicated with.  To specify the base URL of the
// reverse-proxy server, set the vaule of fitbit."BaseURL" directly as part
// of manual initialization.
func NewClient(username, password string) *bitfit.Client {
	c := &bitfit.Client{}
	c.Client = &http.Client{
		Transport: c,
	}
	c.Authorizer = func(r *http.Request) error {
		r.SetBasicAuth(username, password)
		return nil
	}
	c.Initializer = func() (bool, error) {
		return true, nil
	}
	return c
}

// Init initializes the bitfit package to make requests just as it would
// to the FitBit API, but with HTTP basic authentication used to send the
// request to a reverse-proxy server that handles OAuth2 authorization on
// behalf of the client before sending the request along to the FitBit API.
func Init(baseURL, username, password string) error {
	c := NewClient(username, password)
	if err := c.Init(); err != nil {
		return err
	}
	bitfit.DefaultClient = c
	bitfit.BaseURL = baseURL
	return nil
}
