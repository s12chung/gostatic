/*
	Basic CLI interface for gostatic apps. Parses the flags and calls the appropriate functions of the given app.
*/
package cli

import (
	"flag"
	"fmt"
	"os"
)

type App interface {
	// Runs the server to host the generated files of the static web page
	RunFileServer() error
	// Runs a web application server that computes the route responses in real time
	Host() error
	// Generates the static web pages
	Generate() error

	// The path of the generates files of the static web page
	GeneratedPath() string
	// The port of the file server
	FileServerPort() int
	// The port of the web application server
	ServerPort() int
}

// Returns the name of the executable from the Args
func DefaultName() string {
	return os.Args[0]
}

// Returns the default Args for flag.FlagSet
func DefaultArgs() []string {
	return os.Args[1:]
}

// Runs the App with the default settings
func Run(application App) error {
	return NewCli(DefaultName(), application).Run(DefaultArgs())
}

type Cli struct {
	app  App
	flag *flag.FlagSet
}

func NewCli(name string, app App) *Cli {
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	return &Cli{app, f}
}

// Takes the args and parses the flag to run the correct App function
func (cli *Cli) Run(args []string) error {
	app := cli.app

	fileServerPtr := cli.flag.Bool("file-server", false, fmt.Sprintf("Serves, but not generates, files in %v on localhost:%v", app.GeneratedPath(), app.FileServerPort()))
	serverPtr := cli.flag.Bool("server", false, fmt.Sprintf("Hosts server on localhost:%v", app.ServerPort()))
	err := cli.flag.Parse(args)
	if err != nil {
		return nil
	}

	if *fileServerPtr {
		return app.RunFileServer()
	} else {
		if *serverPtr {
			return app.Host()
		} else {
			return app.Generate()
		}
	}
}
