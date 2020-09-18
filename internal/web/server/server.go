package server

import (
	"context"
	"fmt"
	"github.com/restic/restic/internal/options"
	"github.com/restic/restic/internal/web/render"
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
		handleError(err, w, http.StatusInternalServerError)
		return
	}

	err = render.Render(e.Template, w, renderContext)
	if err != nil {
		handleError(err, w, http.StatusInternalServerError)
		return
	}
}

type WebServer struct {
	Endpoints []*HttpCmdEndpoint
}

func (ws *WebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := render.Render("index", w, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

func handleError(err error, w http.ResponseWriter, status int) {
	title := ""
	if status == 404 {
		title = "Page not found"
	} else if status == 400 {
		title = "Bad request"
	} else {
		title = "Internal server error"
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)

	fmt.Printf("failed to execute: %v\n", err)

	err = render.Render("error", w, struct {
		Err    error
		Status int
		Title  string
	}{Err: err, Status: status, Title: title})
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
}
