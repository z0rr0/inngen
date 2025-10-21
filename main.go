package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Define command-line flags
	checkINN := flag.String("c", "", "Check if INN is valid")
	genPhysical := flag.Int("f", -1, "Generate INNs for physical persons (default 5 items)")
	genJuridical := flag.Int("j", -1, "Generate INNs for juridical persons (default 5 items)")
	runWeb := flag.Bool("w", false, "Run as web application (default host:port is 127.0.0.1:2288)")

	flag.Parse()

	// Check which flag was provided
	flagsProvided := 0
	if *checkINN != "" {
		flagsProvided++
	}
	if *genPhysical >= 0 {
		flagsProvided++
	}
	if *genJuridical >= 0 {
		flagsProvided++
	}
	if *runWeb {
		flagsProvided++
	}

	// If no flags provided, show usage
	if flagsProvided == 0 {
		fmt.Println("INN (Taxpayer Identification Number) Generator and Validator")
		fmt.Println("\nUsage:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Handle check INN
	if *checkINN != "" {
		valid := ValidateINN(*checkINN)
		fmt.Println(FormatValidationResult(*checkINN, valid))
		return
	}

	// Handle generate physical INNs
	if *genPhysical >= 0 {
		count := *genPhysical
		if count == 0 {
			count = 5
		}
		fmt.Printf("Generated %d INN(s) for physical persons:\n", count)
		for i := 0; i < count; i++ {
			inn := GeneratePhysicalINN()
			fmt.Printf("%d. %s\n", i+1, inn)
		}
		return
	}

	// Handle generate juridical INNs
	if *genJuridical >= 0 {
		count := *genJuridical
		if count == 0 {
			count = 5
		}
		fmt.Printf("Generated %d INN(s) for juridical persons:\n", count)
		for i := 0; i < count; i++ {
			inn := GenerateJuridicalINN()
			fmt.Printf("%d. %s\n", i+1, inn)
		}
		return
	}

	// Handle run web application
	if *runWeb {
		addr := "127.0.0.1:2288"
		fmt.Printf("Starting web server on %s...\n", addr)
		if err := startWebServer(addr); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting web server: %v\n", err)
			os.Exit(1)
		}
	}
}
