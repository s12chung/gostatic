/*
Package cli has a basic CLI interface for gostatic apps. Parses the flags and calls the appropriate functions of the given app.
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

// DefaultName returns the name of the executable from the Args
func DefaultName() string {
	return os.Args[0]
}

// DefaultArgs returns the default Args for flag.FlagSet
func DefaultArgs() []string {
	return os.Args[1:]
}

// RunDefault runs the App with the default settings
func RunDefault(application App) error {
	return Run(DefaultName(), application, DefaultArgs())
}

// Run takes the args and parses the flag to run the correct App function
func Run(name string, application App, args []string) error {
	f := flag.NewFlagSet(name, flag.ContinueOnError)

	fileServerPtr := f.Bool("file-server", false, fmt.Sprintf("Serves, but not generates, files in %v on localhost:%v", application.GeneratedPath(), application.FileServerPort()))
	serverPtr := f.Bool("server", false, fmt.Sprintf("Hosts server on localhost:%v", application.ServerPort()))
	err := f.Parse(args)
	if err != nil {
		return nil
	}

	if *fileServerPtr {
		return application.RunFileServer()
	}
	if *serverPtr {
		return application.Host()
	}
	return application.Generate()
}
