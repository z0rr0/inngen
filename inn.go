package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// INN represents a Taxpayer Identification Number
type INN struct {
	Value string
}

// coefficients for checksum calculation
var (
	coeff10_1 = []int{2, 4, 10, 3, 5, 9, 4, 6, 8}
	coeff12_1 = []int{7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
	coeff12_2 = []int{3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// calculateChecksum calculates checksum for given digits using coefficients
func calculateChecksum(digits []int, coeffs []int) int {
	sum := 0
	for i, coeff := range coeffs {
		sum += digits[i] * coeff
	}
	checksum := sum % 11
	if checksum == 10 {
		checksum = 0
	}
	return checksum
}

// ValidateINN checks if INN is valid
func ValidateINN(inn string) bool {
	// Check length
	if len(inn) != 10 && len(inn) != 12 {
		return false
	}

	// Check all characters are digits
	digits := make([]int, len(inn))
	for i, c := range inn {
		if c < '0' || c > '9' {
			return false
		}
		digits[i] = int(c - '0')
	}

	if len(inn) == 10 {
		// Juridical person (10 digits)
		checksum := calculateChecksum(digits, coeff10_1)
		return checksum == digits[9]
	} else {
		// Physical person (12 digits)
		checksum1 := calculateChecksum(digits, coeff12_1)
		if checksum1 != digits[10] {
			return false
		}
		checksum2 := calculateChecksum(digits, coeff12_2)
		return checksum2 == digits[11]
	}
}

// GeneratePhysicalINN generates a valid 12-digit INN for a physical person
func GeneratePhysicalINN() string {
	// Generate first 10 digits randomly
	digits := make([]int, 12)
	for i := 0; i < 10; i++ {
		if i == 0 {
			// First digit should not be 0
			digits[i] = rand.Intn(9) + 1
		} else {
			digits[i] = rand.Intn(10)
		}
	}

	// Calculate and set checksums
	digits[10] = calculateChecksum(digits, coeff12_1)
	digits[11] = calculateChecksum(digits, coeff12_2)

	// Convert to string
	inn := ""
	for _, d := range digits {
		inn += strconv.Itoa(d)
	}
	return inn
}

// GenerateJuridicalINN generates a valid 10-digit INN for a juridical person
func GenerateJuridicalINN() string {
	// Generate first 9 digits randomly
	digits := make([]int, 10)
	for i := 0; i < 9; i++ {
		if i == 0 {
			// First digit should not be 0
			digits[i] = rand.Intn(9) + 1
		} else {
			digits[i] = rand.Intn(10)
		}
	}

	// Calculate and set checksum
	digits[9] = calculateChecksum(digits, coeff10_1)

	// Convert to string
	inn := ""
	for _, d := range digits {
		inn += strconv.Itoa(d)
	}
	return inn
}

// FormatValidationResult formats the validation result
func FormatValidationResult(inn string, valid bool) string {
	if valid {
		innType := "juridical"
		if len(inn) == 12 {
			innType = "physical"
		}
		return fmt.Sprintf("INN %s is valid (%s person)", inn, innType)
	}
	return fmt.Sprintf("INN %s is invalid", inn)
}
