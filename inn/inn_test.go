package inn

import (
	"errors"
	"strings"
	"testing"
)

func TestValidator_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		inn            string
		requiredLength int
		wantErr        error
	}{
		{
			name:           "valid physical INN",
			inn:            "500100732259",
			requiredLength: 0,
			wantErr:        nil,
		},
		{
			name:           "valid physical INN with explicit length",
			inn:            "500100732259",
			requiredLength: PhysicalLength,
			wantErr:        nil,
		},
		{
			name:           "valid physical INN with whitespace",
			inn:            "  500100732259  ",
			requiredLength: 0,
			wantErr:        nil,
		},
		{
			name:           "valid juridical INN",
			inn:            "7707083893",
			requiredLength: 0,
			wantErr:        nil,
		},
		{
			name:           "valid juridical INN with explicit length",
			inn:            "7707083893",
			requiredLength: JuridicalLength,
			wantErr:        nil,
		},
		{
			name:           "valid juridical INN with whitespace",
			inn:            "  7707083893  ",
			requiredLength: 0,
			wantErr:        nil,
		},

		{
			name:           "empty INN",
			inn:            "",
			requiredLength: 0,
			wantErr:        ErrInnLength,
		},
		{
			name:           "too short INN",
			inn:            "123456789",
			requiredLength: 0,
			wantErr:        ErrInnLength,
		},
		{
			name:           "11 digits (invalid length)",
			inn:            "12345678901",
			requiredLength: 0,
			wantErr:        ErrInnLength,
		},
		{
			name:           "13 digits (too long)",
			inn:            "1234567890123",
			requiredLength: 0,
			wantErr:        ErrInnLength,
		},
		{
			name:           "length mismatch - 10 digits required but 12 provided",
			inn:            "500100732259",
			requiredLength: JuridicalLength,
			wantErr:        ErrInnLength,
		},
		{
			name:           "length mismatch - 12 digits required but 10 provided",
			inn:            "7707083893",
			requiredLength: PhysicalLength,
			wantErr:        ErrInnLength,
		},
		{
			name:           "invalid required length",
			inn:            "7707083893",
			requiredLength: 5,
			wantErr:        ErrInnLength,
		},
		{
			name:           "contains letter",
			inn:            "77070838A3",
			requiredLength: 0,
			wantErr:        ErrInnLength,
		},
		{
			name:           "contains space in middle",
			inn:            "7707 083893",
			requiredLength: 0,
			wantErr:        ErrInnLength,
		},
		{
			name:           "contains dash",
			inn:            "7707-083893",
			requiredLength: 0,
			wantErr:        ErrInnLength,
		},
		{
			name:           "contains special character",
			inn:            "7707083893!",
			requiredLength: 0,
			wantErr:        ErrInnLength,
		},
		{
			name:           "contains unicode digit",
			inn:            "770708389٣", // Arabic-Indic digit 3
			requiredLength: 0,
			wantErr:        ErrInnLength,
		},
		{
			name:           "physical INN wrong 11th digit",
			inn:            "500100732258", // the latest digit should be 9
			requiredLength: PhysicalLength,
			wantErr:        ErrInnChecksum,
		},
		{
			name:           "physical INN wrong 10th digit",
			inn:            "500100732159", // 10th digit should be 5
			requiredLength: PhysicalLength,
			wantErr:        ErrInnChecksum,
		},
		{
			name:           "physical INN all ones",
			inn:            "111111111111",
			requiredLength: PhysicalLength,
			wantErr:        ErrInnChecksum,
		},
		{
			name:           "juridical INN wrong checksum",
			inn:            "7707083892", // the latest digit should be 3
			requiredLength: JuridicalLength,
			wantErr:        ErrInnChecksum,
		},
		{
			name:           "juridical INN all nines",
			inn:            "9999999999",
			requiredLength: JuridicalLength,
			wantErr:        ErrInnChecksum,
		},
		{
			name:           "juridical INN all ones",
			inn:            "1111111111",
			requiredLength: JuridicalLength,
			wantErr:        ErrInnChecksum,
		},
		{
			name:           "physical INN all zeros is valid",
			inn:            "000000000000",
			requiredLength: PhysicalLength,
			wantErr:        nil,
		},
		{
			name:           "juridical INN all zeros is valid",
			inn:            "0000000000",
			requiredLength: JuridicalLength,
			wantErr:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			validator := NewValidator(tt.inn, tt.requiredLength)
			err := validator.Validate()

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("Validate() error = %v, wantErr nil", err)
				}
				return
			}

			if err == nil {
				t.Errorf("Validate() error = nil, wantErr %v", tt.wantErr)
				return
			}

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFmtResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		inn            string
		err            error
		wantSubstrings []string
	}{
		{
			name:           "valid physical INN",
			inn:            "500100732259",
			err:            nil,
			wantSubstrings: []string{"500100732259", "valid", "physical"},
		},
		{
			name:           "valid juridical INN",
			inn:            "7707083893",
			err:            nil,
			wantSubstrings: []string{"7707083893", "valid", "juridical"},
		},
		{
			name:           "invalid INN with error",
			inn:            "123",
			err:            ErrInnLength,
			wantSubstrings: []string{"123", "invalid"},
		},
		{
			name:           "checksum error",
			inn:            "7707083892",
			err:            ErrInnChecksum,
			wantSubstrings: []string{"7707083892", "invalid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := FmtResult(tt.inn, tt.err)

			for _, substr := range tt.wantSubstrings {
				if !strings.Contains(result, substr) {
					t.Errorf("FmtResult() = %q, want to contain %q", result, substr)
				}
			}
		})
	}
}

func BenchmarkValidator_Validate_Physical(b *testing.B) {
	validator := NewValidator("500100732259", PhysicalLength)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate()
	}
}

func BenchmarkValidator_Validate_Juridical(b *testing.B) {
	validator := NewValidator("7707083893", JuridicalLength)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate()
	}
}

func BenchmarkGeneratePhysicalINN(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GeneratePhysicalINN()
	}
}

func BenchmarkGenerateJuridicalINN(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateJuridicalINN()
	}
}
