package server

import (
	"context"
	"fmt"
	"github.com/restic/restic/internal/options"
	"github.com/restic/restic/internal/web/render"
	"log"
	"net/http"
)

type Config struct {
	Password        string
	PasswordFile    string
	Repo            string
	PasswordCommand string
	KeyHint         string
	Quiet           bool
	Verbose         int
	NoLock          bool
	CacheDir        string
	NoCache         bool
	CACerts         []string
	TLSClientCert   string
	CleanupCache    bool

	LimitUploadKb   int
	LimitDownloadKb int

	Ctx context.Context

	// verbosity is set as follows:
	//  0 means: don't print any messages except errors, this is used when --quiet is specified
	//  1 is the default: print essential messages
	//  2 means: print more messages, report minor things, this is used when --verbose is specified
	//  3 means: print very detailed debug messages, this is used when --verbose=2 is specified
	Verbosity uint

	Options []string

	Extended options.Options

	Args []string
}

type HttpCmdFunc func(r *http.Request, serverConfig Config) (renderContext interface{}, err error)

type HttpCmdEndpoint struct {
	Path     string
	Config   Config
	Template string
	Cmd      HttpCmdFunc
}

func (e *HttpCmdEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	renderContext, err := e.Cmd(r, e.Config)
	if err != nil {
		log.Print(err)
		return
	}

	err = render.Render(e.Template, w, renderContext)
	if err != nil {
		log.Print(err)
		return
	}
}

type WebServer struct {
	Endpoints []*HttpCmdEndpoint
}

func (ws *WebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := render.Render("index", w, "")
	if err != nil {
		log.Print(err)
	}
}

func (ws *WebServer) Run(addr string, serverConfig Config) error {
	fmt.Printf("will start webserver at %s\n", addr)

	http.Handle("/", ws)

	for _, endpoint := range ws.Endpoints {
		endpoint.Config = serverConfig
		fmt.Printf("registered cmd handler: %s\n", endpoint.Path)
		http.Handle(endpoint.Path, endpoint)
	}

	return http.ListenAndServe(addr, nil)
}
