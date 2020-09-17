package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/restic/restic/internal/web/server"
	"github.com/spf13/cobra"
	"net/http"
)

const defaultPort = 6723

var serverCmd = &cobra.Command{
	Use:   "webserver [address]",
	Short: "Starts webserver at the specified address, if not given :6723 will be used as address",
	Long: `
The "webserver" command is used to start http server in localhost on the specified port.

EXIT STATUS
===========

Exit status is 0 if the command was successful, and non-zero if there was any error.
`,
	DisableAutoGenTag: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer(globalOptions, args)
	},
}

func init() {
	cmdRoot.AddCommand(serverCmd)
}

func runServer(gopts GlobalOptions, args []string) error {
	address := fmt.Sprintf(":%d", defaultPort)
	if len(args) > 0 {
		address = args[0]
	}

	websrv := server.WebServer{
		Endpoints: []*server.HttpCmdEndpoint{
			{
				Path:     "/snapshots",
				Cmd:      serveSnapshots,
				Template: "snapshots",
			},
		},
	}

	config := server.Config{
		Password:        gopts.password,
		Args:            args,
		PasswordFile:    gopts.PasswordFile,
		Repo:            gopts.Repo,
		PasswordCommand: gopts.PasswordCommand,
		KeyHint:         gopts.KeyHint,
		Quiet:           gopts.Quiet,
		Verbose:         gopts.Verbose,
		NoLock:          gopts.NoLock,
		CacheDir:        gopts.CacheDir,
		NoCache:         gopts.NoCache,
		CACerts:         gopts.CACerts,
		TLSClientCert:   gopts.TLSClientCert,
		CleanupCache:    gopts.CleanupCache,
		LimitUploadKb:   gopts.LimitUploadKb,
		LimitDownloadKb: gopts.LimitDownloadKb,
		Ctx:             gopts.ctx,
		Verbosity:       gopts.verbosity,
		Options:         gopts.Options,
		Extended:        gopts.extended,
	}

	return websrv.Run(address, config)
}

func serveSnapshots(r *http.Request, serverConfig server.Config) (renderContext interface{}, err error) {
	var buf bytes.Buffer

	severGlobalOptions := GlobalOptions{}
	severGlobalOptions.JSON = true
	severGlobalOptions.stdout = &buf
	severGlobalOptions.password = serverConfig.Password
	severGlobalOptions.PasswordFile = serverConfig.PasswordFile
	severGlobalOptions.Repo = serverConfig.Repo
	severGlobalOptions.PasswordCommand = serverConfig.PasswordCommand
	severGlobalOptions.KeyHint = serverConfig.KeyHint
	severGlobalOptions.Quiet = serverConfig.Quiet
	severGlobalOptions.Verbose = serverConfig.Verbose
	severGlobalOptions.NoLock = serverConfig.NoLock
	severGlobalOptions.CacheDir = serverConfig.CacheDir
	severGlobalOptions.NoCache = serverConfig.NoCache
	severGlobalOptions.CACerts = serverConfig.CACerts
	severGlobalOptions.TLSClientCert = serverConfig.TLSClientCert
	severGlobalOptions.CleanupCache = serverConfig.CleanupCache
	severGlobalOptions.LimitUploadKb = serverConfig.LimitUploadKb
	severGlobalOptions.LimitDownloadKb = serverConfig.LimitDownloadKb
	severGlobalOptions.ctx = serverConfig.Ctx
	severGlobalOptions.verbosity = serverConfig.Verbosity
	severGlobalOptions.Options = serverConfig.Options
	severGlobalOptions.extended = serverConfig.Extended

	err = runSnapshots(SnapshotOptions{}, severGlobalOptions, serverConfig.Args)
	if err != nil {
		return
	}

	snapshots := make([]Snapshot, 0)
	decoder := json.NewDecoder(&buf)
	err = decoder.Decode(&snapshots)
	return struct {
		Snapshots []Snapshot
		Repo      string
	}{
		Snapshots: snapshots,
		Repo:      severGlobalOptions.Repo,
	}, err
}
