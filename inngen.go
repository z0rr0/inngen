// Package main provides INNGen command line application.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/z0rr0/inngen/inn"
)

const name = "INNGen"

var (
	// Version is a git version.
	Version = "v0.0.0" //nolint:gochecknoglobals
	// Revision is a revision number.
	Revision = "git:0000000" //nolint:gochecknoglobals
	// BuildDate is a build date.
	BuildDate = "1970-01-01T00:00:00" //nolint:gochecknoglobals
	// GoVersion is a runtime Go language version, for example "go1.00.0".
	GoVersion = runtime.Version() //nolint:gochecknoglobals
)

func main() {
	var (
		checkINN     string
		genPhysical  = 5
		genJuridical = 5
		runWeb       = "127.0.0.1:2288"
	)
	defer func() {
		if r := recover(); r != nil {
			slog.Error("abnormal termination", "version", Version, "error", r)
			_, writeErr := fmt.Fprintf(os.Stderr, "abnormal termination: %v\n", string(debug.Stack()))
			if writeErr != nil {
				slog.Error("failed to write stack trace", "error", writeErr)
			}
		}
	}()
	flag.StringVar(&checkINN, "c", "", "check if INN is valid")
	flag.StringVar(&runWeb, "w", runWeb, "run as web application")
	flag.IntVar(&genPhysical, "f", genPhysical, "generate INNs for physical persons")
	flag.IntVar(&genJuridical, "j", genJuridical, "generate INNs for juridical persons")
	version := flag.Bool("v", false, "show version")

	flag.Parse()
	if *version {
		fmt.Println(name + ": INN (Taxpayer Identification Number) Generator and Validator")
		fmt.Printf(
			"Version: %v\nRevision: %v\nBuild date: %v\nGo version: %v\n",
			Version, Revision, BuildDate, GoVersion,
		)
		fmt.Println("\nUsage:")
		flag.PrintDefaults()
		return
	}

	if checkINN != "" {
		validator := inn.NewValidator(checkINN, 0)
		err := validator.Validate()
		fmt.Println(inn.FmtResult(checkINN, err))
		return
	}

	if genPhysical > 0 {
		fmt.Printf("Generated %d INN(s) for physical persons:\n", genPhysical)
		for i := range genPhysical {
			value, err := inn.GeneratePhysicalINN()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error generating INN: %v\n", err)
				os.Exit(1) //nolint:gocritic
			}
			fmt.Printf("%-3d %s\n", i+1, value)
		}
	}

	if genJuridical > 0 {
		fmt.Printf("Generated %d INN(s) for juridical persons:\n", genJuridical)
		for i := range genJuridical {
			value, err := inn.GenerateJuridicalINN()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error generating INN: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("%-3d %s\n", i+1, value)
		}
	}

	//  if *runWeb {
	//	addr := "127.0.0.1:2288"
	//	fmt.Printf("Starting web server on %s...\n", addr)
	//	if err := startWebServer(addr); err != nil {
	//		fmt.Fprintf(os.Stderr, "Error starting web server: %v\n", err)
	//		os.Exit(1)
	//	}
	//  }
}
