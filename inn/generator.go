package inn

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
)

var (
	// maxFirst and maxNext are the maximum values for the first and next digits, respectively.
	maxFirst = big.NewInt(9)  //nolint:gochecknoglobals
	maxNext  = big.NewInt(10) //nolint:gochecknoglobals

	// ErrInnGeneration is an error indicating an error during INN generation.
	ErrInnGeneration = errors.New("failed to generate INN")
)

// GeneratePhysicalINN generates a valid 12-digit INN for a physical person.
func GeneratePhysicalINN() (string, error) {
	digits, err := generateINN(PhysicalLength-2, PhysicalLength, rand.Reader)
	if err != nil {
		return "", errors.Join(ErrInnGeneration, err)
	}

	d, err := calculateControlValue(weightsPhysical1, digits)
	if err != nil {
		return "", errors.Join(ErrInnGeneration, err)
	}

	digits[10] = d

	d, err = calculateControlValue(weightsPhysical2, digits)
	if err != nil {
		return "", errors.Join(ErrInnGeneration, err)
	}

	digits[11] = d

	return digitsToString(digits), nil
}

// GenerateJuridicalINN generates a valid 10-digit INN for a juridical person.
func GenerateJuridicalINN() (string, error) {
	digits, err := generateINN(JuridicalLength-1, JuridicalLength, rand.Reader)
	if err != nil {
		return "", errors.Join(ErrInnGeneration, err)
	}

	d, err := calculateControlValue(weightsJuridical, digits)
	if err != nil {
		return "", errors.Join(ErrInnGeneration, err)
	}

	digits[9] = d

	return digitsToString(digits), nil
}

func generateINN(length, capLen int, reader io.Reader) ([]int, error) {
	if capLen != PhysicalLength && capLen != JuridicalLength {
		return nil, fmt.Errorf("invalid INN length: %d", length)
	}

	var (
		digits = make([]int, capLen)
		d      *big.Int
		err    error
	)

	d, err = rand.Int(reader, maxFirst)
	if err != nil {
		return nil, fmt.Errorf("failed to generate first digit: %w", err)
	}

	digits[0] = int(d.Int64()) + 1 // 1st digit should not be 0

	for i := 1; i < length; i++ {
		d, err = rand.Int(reader, maxNext)
		if err != nil {
			return nil, fmt.Errorf("failed to generate digit %d: %w", i, err)
		}

		digits[i] = int(d.Int64())
	}

	return digits, nil
}

func digitsToString(digits []int) string {
	var inn strings.Builder

	for _, d := range digits {
		inn.WriteString(strconv.Itoa(d))
	}

	return inn.String()
}
