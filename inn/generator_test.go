package inn

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"unicode"
)

func TestGeneratePhysicalINN(t *testing.T) {
	t.Parallel()

	t.Run("generates valid 12-digit INN", func(t *testing.T) {
		t.Parallel()

		inn, err := GeneratePhysicalINN()
		if err != nil {
			t.Fatalf("GeneratePhysicalINN() error = %v, want nil", err)
		}

		if len(inn) != PhysicalLength {
			t.Errorf("GeneratePhysicalINN() length = %d, want %d", len(inn), PhysicalLength)
		}

		for i, r := range inn {
			if !unicode.IsDigit(r) {
				t.Errorf("GeneratePhysicalINN() char at position %d = %c, want digit", i, r)
			}
		}
	})

	t.Run("generated INN passes validation", func(t *testing.T) {
		t.Parallel()

		inn, err := GeneratePhysicalINN()
		if err != nil {
			t.Fatalf("GeneratePhysicalINN() error = %v, want nil", err)
		}

		validator := NewValidator(inn, PhysicalLength)
		if err = validator.Validate(); err != nil {
			t.Errorf("Generated INN %s failed validation: %v", inn, err)
		}
	})

	t.Run("first digit is never zero", func(t *testing.T) {
		t.Parallel()
		const iterations = 100

		for i := 0; i < iterations; i++ {
			inn, err := GeneratePhysicalINN()
			if err != nil {
				t.Fatalf("GeneratePhysicalINN() error = %v, want nil", err)
			}

			if inn[0] == '0' {
				t.Errorf("GeneratePhysicalINN() first digit = 0, want 1-9")
			}
		}
	})

	t.Run("generates diverse INNs", func(t *testing.T) {
		t.Parallel()
		const (
			iterations = 100
			minUnique  = iterations * 95 / 100 // at least 95% should be unique (allow for rare collisions)
		)
		inns := make(map[string]struct{}, iterations)

		for i := 0; i < iterations; i++ {
			inn, err := GeneratePhysicalINN()
			if err != nil {
				t.Fatalf("GeneratePhysicalINN() error = %v, want nil", err)
			}
			inns[inn] = struct{}{}
		}

		if uniqueCount := len(inns); uniqueCount < minUnique {
			t.Errorf("GeneratePhysicalINN() generated %d unique INNs out of %d, want at least %d", uniqueCount, iterations, minUnique)
		}
	})
}

func TestGenerateJuridicalINN(t *testing.T) {
	t.Parallel()

	t.Run("generates valid 10-digit INN", func(t *testing.T) {
		t.Parallel()

		inn, err := GenerateJuridicalINN()
		if err != nil {
			t.Fatalf("GenerateJuridicalINN() error = %v, want nil", err)
		}

		if len(inn) != JuridicalLength {
			t.Errorf("GenerateJuridicalINN() length = %d, want %d", len(inn), JuridicalLength)
		}

		for i, r := range inn {
			if !unicode.IsDigit(r) {
				t.Errorf("GenerateJuridicalINN() char at position %d = %c, want digit", i, r)
			}
		}
	})

	t.Run("generated INN passes validation", func(t *testing.T) {
		t.Parallel()

		inn, err := GenerateJuridicalINN()
		if err != nil {
			t.Fatalf("GenerateJuridicalINN() error = %v, want nil", err)
		}

		validator := NewValidator(inn, JuridicalLength)
		if err = validator.Validate(); err != nil {
			t.Errorf("Generated INN %s failed validation: %v", inn, err)
		}
	})

	t.Run("first digit is never zero", func(t *testing.T) {
		t.Parallel()
		const iterations = 100

		for i := 0; i < iterations; i++ {
			inn, err := GenerateJuridicalINN()
			if err != nil {
				t.Fatalf("GenerateJuridicalINN() error = %v, want nil", err)
			}

			if inn[0] == '0' {
				t.Errorf("GenerateJuridicalINN() first digit = 0, want 1-9")
			}
		}
	})

	t.Run("generates diverse INNs", func(t *testing.T) {
		t.Parallel()
		const (
			iterations = 50
			minUnique  = iterations * 95 / 100 // at least 95% should be unique (allow for rare collisions)
		)
		inns := make(map[string]struct{}, iterations)

		for i := 0; i < iterations; i++ {
			inn, err := GenerateJuridicalINN()
			if err != nil {
				t.Fatalf("GenerateJuridicalINN() error = %v, want nil", err)
			}
			inns[inn] = struct{}{}
		}

		if uniqueCount := len(inns); uniqueCount < minUnique {
			t.Errorf("GenerateJuridicalINN() generated %d unique INNs out of %d, want at least %d", uniqueCount, iterations, minUnique)
		}
	})
}

func TestGenerateINN_ErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		length  int
		capLen  int
		reader  io.Reader
		wantErr bool
	}{
		{
			name:    "invalid capLen too small",
			length:  8,
			capLen:  9,
			reader:  bytes.NewReader([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9}),
			wantErr: true,
		},
		{
			name:    "invalid capLen too large",
			length:  11,
			capLen:  13,
			reader:  bytes.NewReader([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}),
			wantErr: true,
		},
		{
			name:    "reader error on first digit",
			length:  JuridicalLength - 1,
			capLen:  JuridicalLength,
			reader:  &errorReader{},
			wantErr: true,
		},
		{
			name:    "reader returns EOF",
			length:  JuridicalLength - 1,
			capLen:  JuridicalLength,
			reader:  bytes.NewReader([]byte{}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := generateINN(tt.length, tt.capLen, tt.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateINN() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDigitsToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		digits []int
		want   string
	}{
		{
			name:   "single digit",
			digits: []int{5},
			want:   "5",
		},
		{
			name:   "multiple digits",
			digits: []int{1, 2, 3, 4, 5},
			want:   "12345",
		},
		{
			name:   "with zeros",
			digits: []int{1, 0, 0, 5},
			want:   "1005",
		},
		{
			name:   "all zeros",
			digits: []int{0, 0, 0, 0},
			want:   "0000",
		},
		{
			name:   "physical INN length",
			digits: []int{5, 0, 0, 1, 0, 0, 7, 3, 2, 2, 5, 9},
			want:   "500100732259",
		},
		{
			name:   "juridical INN length",
			digits: []int{7, 7, 0, 7, 0, 8, 3, 8, 9, 3},
			want:   "7707083893",
		},
		{
			name:   "empty slice",
			digits: []int{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := digitsToString(tt.digits)
			if got != tt.want {
				t.Errorf("digitsToString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCalculateControlValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		weights    []int
		innNumbers []int
		want       int
		wantErr    bool
	}{
		{
			name:       "physical INN first checksum",
			weights:    weightsPhysical1,
			innNumbers: []int{5, 0, 0, 1, 0, 0, 7, 3, 2, 2, 5, 9},
			want:       5,
		},
		{
			name:       "physical INN second checksum",
			weights:    weightsPhysical2,
			innNumbers: []int{5, 0, 0, 1, 0, 0, 7, 3, 2, 2, 5, 9},
			want:       9,
		},
		{
			name:       "juridical INN checksum",
			weights:    weightsJuridical,
			innNumbers: []int{7, 7, 0, 7, 0, 8, 3, 8, 9, 3},
			want:       3,
		},
		{
			name:       "checksum with remainder > 9",
			weights:    []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			innNumbers: []int{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9},
		},
		{
			name:       "all zeros",
			weights:    weightsPhysical1,
			innNumbers: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:       "inn shorter than weights",
			weights:    weightsPhysical1,
			innNumbers: []int{1, 2, 3},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := calculateControlValue(tt.weights, tt.innNumbers)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateControlValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("calculateControlValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

// errorReader always returns an error on Read
type errorReader struct{}

func (e *errorReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("read error")
}

// Test that both generator functions handle edge cases in checksum calculation
func TestGenerateINN_ChecksumEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("physical INN checksum calculation", func(t *testing.T) {
		t.Parallel()

		for i := 0; i < 50; i++ {
			inn, err := GeneratePhysicalINN()
			if err != nil {
				t.Fatalf("GeneratePhysicalINN() error = %v", err)
			}

			// Manually verify checksum calculation
			digits := make([]int, PhysicalLength)
			for j, r := range inn {
				digits[j] = int(r - '0')
			}

			checksum1, err := calculateControlValue(weightsPhysical1, digits)
			if err != nil {
				t.Fatalf("calculateControlValue() error = %v", err)
			}
			if checksum1 != digits[10] {
				t.Errorf("INN %s: 11th digit = %d, calculated = %d", inn, digits[10], checksum1)
			}

			checksum2, err := calculateControlValue(weightsPhysical2, digits)
			if err != nil {
				t.Fatalf("calculateControlValue() error = %v", err)
			}
			if checksum2 != digits[11] {
				t.Errorf("INN %s: 12th digit = %d, calculated = %d", inn, digits[11], checksum2)
			}
		}
	})

	t.Run("juridical INN checksum calculation", func(t *testing.T) {
		t.Parallel()

		for i := 0; i < 50; i++ {
			inn, err := GenerateJuridicalINN()
			if err != nil {
				t.Fatalf("GenerateJuridicalINN() error = %v", err)
			}

			digits := make([]int, JuridicalLength)
			for j, r := range inn {
				digits[j] = int(r - '0')
			}

			checksum, err := calculateControlValue(weightsJuridical, digits)
			if err != nil {
				t.Fatalf("calculateControlValue() error = %v", err)
			}
			if checksum != digits[9] {
				t.Errorf("INN %s: 10th digit = %d, calculated = %d", inn, digits[9], checksum)
			}
		}
	})
}
