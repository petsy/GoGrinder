package gogrinder

import (
	"flag"
	"fmt"
	"os"
)

// simple helper contains
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Simple command line interface for GoGrinder.
//  * (default is to start/stop test via UI and have a console report)
//  * -no-exec
//  * -no-report
//  * -no-frontend
func GetCLI() (string, bool, bool, bool, int, string, error) {
	// for now try to work with the std. Golang flag package

	// In my research I found this tutorial useful:
	// http://blog.ralch.com/tutorial/golang-subcommands/
	// probably a suitable flag alternative:
	// https://github.com/voxelbrain/goptions

	var filename string = "loadmodel.json"
	var noExec bool
	var noReport bool
	var noFrontend bool
	var port int
	var logLevel string
	var err error = nil

	// no ExitOnError - we maintain control of the program flow
	cli := flag.NewFlagSet("gogrinder", flag.ContinueOnError)

	cli.SetOutput(stdout)

	cli.BoolVar(&noExec, "no-exec", false, "supress auto execution of the test scenario.")
	cli.BoolVar(&noReport, "no-report", false, "supress the console report.")
	cli.BoolVar(&noFrontend, "no-frontend", false, "do not start the web frontend.")
	cli.StringVar(&logLevel, "log-level", "warn", "panic, fatal, error, warn, info, debug")
	cli.IntVar(&port, "port", 3000, "specify the port for the web frontend.")

	cli.Usage = func() {
		fmt.Fprintf(stdout, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(stdout, "  %s base_loadmodel.json -no-frontend\n", os.Args[0])
		fmt.Fprintf(stdout, "\n")
		fmt.Fprintf(stdout, "  arg-1  loadmodel filename.  (defaults 'loadmodel.json')\n")
		cli.PrintDefaults()
		err = fmt.Errorf("Command line usage problem.")
	}

	cli.Parse(os.Args[1:]) // exclude the first

	if cli.NArg() > 1 {
		cli.Usage()
	}

	if cli.NArg() == 1 {
		filename = cli.Arg(0)
	}

	if !contains([]string{"panic", "fatal", "error", "warn", "info", "debug"}, logLevel) {
		cli.Usage()
	}

	if err == nil {
		// check file exists
		if _, ferr := os.Stat(filename); ferr != nil {
			err = fmt.Errorf("File %s does not exist.", filename)
		}
	}

	// check for invalid combination of options
	if noExec && noFrontend {
		err = fmt.Errorf("Invalid combination of -no-exec and -no-frontend.")
	}

	return filename, noExec, noReport, noFrontend, port, logLevel, err
}
