// Package inn provides a validator for Russian Taxpayer Identification Numbers (INN).
package inn

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	// JuridicalLength is the valid length for a juridical entity INN.
	JuridicalLength = 10
	// PhysicalLength is the valid length for a physical person INN.
	PhysicalLength = 12
)

var (
	// ErrInnLength is an error indicating an invalid INN length.
	ErrInnLength = errors.New("invalid INN length")
	// ErrInnChecksum is an error indicating an invalid INN checksum.
	ErrInnChecksum = errors.New("invalid INN checksum")

	// weights for checksum calculation (from https://www.egrul.ru/test_inn.html)
	weightsPhysical1 = []int{7, 2, 4, 10, 3, 5, 9, 4, 6, 8, 0}    //nolint:gochecknoglobals
	weightsPhysical2 = []int{3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8, 0} //nolint:gochecknoglobals
	weightsJuridical = []int{2, 4, 10, 3, 5, 9, 4, 6, 8, 0}       //nolint:gochecknoglobals
)

// Validator validates Russian Taxpayer Identification Numbers (INN).
type Validator struct {
	inn            string
	requiredLength int
}

// NewValidator creates a new validator instance.
func NewValidator(inn string, requiredLength int) *Validator {
	trimmedInn := strings.TrimSpace(inn)
	if requiredLength == 0 {
		requiredLength = len(trimmedInn)
	}
	return &Validator{
		inn:            trimmedInn,
		requiredLength: requiredLength,
	}
}

// Validate checks the correctness of the INN string.
func (v *Validator) Validate() error {
	if v.requiredLength != PhysicalLength && v.requiredLength != JuridicalLength {
		return fmt.Errorf(
			"%w: valid required lengths are %d or %d, got %d",
			ErrInnLength, PhysicalLength, JuridicalLength, v.requiredLength,
		)
	}

	innLength := len(v.inn)
	if innLength != v.requiredLength {
		return fmt.Errorf("%w: got %d, expected %d", ErrInnLength, innLength, v.requiredLength)
	}

	innNumbers := make([]int, innLength)
	for i, r := range v.inn {
		if !unicode.IsDigit(r) {
			return fmt.Errorf("%w: not a decimal number '%c'", ErrInnLength, r)
		}

		digit, err := strconv.Atoi(string(r))
		if err != nil {
			return fmt.Errorf("%w: failed to parse digit '%c': %w", ErrInnChecksum, r, err)
		}

		innNumbers[i] = digit
	}

	if v.requiredLength == PhysicalLength {
		return v.validatePhysical(innNumbers)
	}

	return v.validateJuridical(innNumbers)
}

// validatePhysical checks the validity of a physical person's INN (12 digits).
func (v *Validator) validatePhysical(innNumbers []int) error {
	if n := len(innNumbers); n != PhysicalLength {
		return fmt.Errorf("%w, expected %d digits for physical INN, got %d", ErrInnLength, PhysicalLength, n)
	}

	part1, err := calculateControlValue(weightsPhysical1, innNumbers)
	if err != nil {
		return fmt.Errorf("calculating part1 checksum: %w", err)
	}

	part2, err := calculateControlValue(weightsPhysical2, innNumbers)
	if err != nil {
		return fmt.Errorf("calculating part2 checksum: %w", err)
	}

	if part1 != innNumbers[10] {
		return fmt.Errorf(
			"%w: invalid physical inn, 10th digit is %d, expected %d",
			ErrInnChecksum, innNumbers[10], part1,
		)
	}

	if part2 != innNumbers[11] {
		return fmt.Errorf(
			"%w: invalid physical inn, 11th digit is %d, expected %d",
			ErrInnChecksum, innNumbers[11], part2,
		)
	}

	return nil
}

// validateJuridical checks the validity of a juridical entity's INN (10 digits).
func (v *Validator) validateJuridical(innNumbers []int) error {
	if n := len(innNumbers); n != JuridicalLength {
		return fmt.Errorf("expected %d digits for juridical INN, got %d", JuridicalLength, n)
	}

	controlValue, err := calculateControlValue(weightsJuridical, innNumbers)
	if err != nil {
		return fmt.Errorf("calculating juridical checksum: %w", err)
	}

	if controlValue != innNumbers[9] {
		return fmt.Errorf(
			"%w: invalid juridical inn, expected %d, got %d",
			ErrInnChecksum, controlValue, innNumbers[9],
		)
	}

	return nil
}

// calculateControlValue calculates the checksum digit based on weights and INN digits.
func calculateControlValue(weights []int, innNumbers []int) (int, error) {
	const checkpointThreshold = 9

	if n, m := len(innNumbers), len(weights); n < m {
		return 0, fmt.Errorf("inn length %d is less than weights length %d", n, m)
	}

	sum := 0
	for i, w := range weights {
		sum += innNumbers[i] * w
	}

	remainder := sum % 11
	if remainder > checkpointThreshold {
		return remainder % 10, nil
	}
	return remainder, nil
}

// FmtResult returns a result string for a given INN.
func FmtResult(inn string, err error) string {
	if err != nil {
		return fmt.Sprintf("INN '%s' invalid: %v", inn, err)
	}

	innType := "juridical"
	if len(inn) == PhysicalLength {
		innType = "physical"
	}
	return fmt.Sprintf("INN %s is valid (%s person)", inn, innType)
}
