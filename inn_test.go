package main

import (
	"testing"
)

// Test cases with known valid INNs
var validINNs = []struct {
	inn     string
	innType string
}{
	{"7707083893", "juridical"},  // 10-digit valid INN
	{"500100732259", "physical"}, // 12-digit valid INN
}

// Test ValidateINN function
func TestValidateINN(t *testing.T) {
	tests := []struct {
		name      string
		inn       string
		wantValid bool
	}{
		{"valid 10-digit INN", "7707083893", true},
		{"valid 12-digit INN", "500100732259", true},
		{"invalid length 11", "12345678901", false},
		{"invalid length 9", "123456789", false},
		{"invalid checksum 10-digit", "7707083892", false},
		{"invalid checksum 12-digit", "500100732258", false},
		{"contains non-digits", "123456789a", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateINN(tt.inn)
			if got != tt.wantValid {
				t.Errorf("ValidateINN(%q) = %v, want %v", tt.inn, got, tt.wantValid)
			}
		})
	}
}

// Test GeneratePhysicalINN function
func TestGeneratePhysicalINN(t *testing.T) {
	for i := 0; i < 100; i++ {
		inn := GeneratePhysicalINN()

		// Check length
		if len(inn) != 12 {
			t.Errorf("GeneratePhysicalINN() generated INN with length %d, want 12", len(inn))
		}

		// Check validity
		if !ValidateINN(inn) {
			t.Errorf("GeneratePhysicalINN() generated invalid INN: %s", inn)
		}

		// Check all characters are digits
		for _, c := range inn {
			if c < '0' || c > '9' {
				t.Errorf("GeneratePhysicalINN() generated INN with non-digit character: %s", inn)
			}
		}
	}
}

// Test GenerateJuridicalINN function
func TestGenerateJuridicalINN(t *testing.T) {
	for i := 0; i < 100; i++ {
		inn := GenerateJuridicalINN()

		// Check length
		if len(inn) != 10 {
			t.Errorf("GenerateJuridicalINN() generated INN with length %d, want 10", len(inn))
		}

		// Check validity
		if !ValidateINN(inn) {
			t.Errorf("GenerateJuridicalINN() generated invalid INN: %s", inn)
		}

		// Check all characters are digits
		for _, c := range inn {
			if c < '0' || c > '9' {
				t.Errorf("GenerateJuridicalINN() generated INN with non-digit character: %s", inn)
			}
		}
	}
}

// Test checksum calculation
func TestCalculateChecksum(t *testing.T) {
	tests := []struct {
		name   string
		digits []int
		coeffs []int
		want   int
	}{
		{
			name:   "10-digit INN checksum",
			digits: []int{7, 7, 0, 7, 0, 8, 3, 8, 9},
			coeffs: coeff10_1,
			want:   3,
		},
		{
			name:   "12-digit INN first checksum",
			digits: []int{5, 0, 0, 1, 0, 0, 7, 3, 2, 2},
			coeffs: coeff12_1,
			want:   5,
		},
		{
			name:   "12-digit INN second checksum",
			digits: []int{5, 0, 0, 1, 0, 0, 7, 3, 2, 2, 5},
			coeffs: coeff12_2,
			want:   9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateChecksum(tt.digits, tt.coeffs)
			if got != tt.want {
				t.Errorf("calculateChecksum() = %d, want %d", got, tt.want)
			}
		})
	}
}

// Test FormatValidationResult
func TestFormatValidationResult(t *testing.T) {
	tests := []struct {
		name  string
		inn   string
		valid bool
		want  string
	}{
		{
			name:  "valid juridical INN",
			inn:   "7707083893",
			valid: true,
			want:  "INN 7707083893 is valid (juridical person)",
		},
		{
			name:  "valid physical INN",
			inn:   "500100732259",
			valid: true,
			want:  "INN 500100732259 is valid (physical person)",
		},
		{
			name:  "invalid INN",
			inn:   "1234567890",
			valid: false,
			want:  "INN 1234567890 is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatValidationResult(tt.inn, tt.valid)
			if got != tt.want {
				t.Errorf("FormatValidationResult() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidateINN(b *testing.B) {
	inn := "7707083893"
	for i := 0; i < b.N; i++ {
		ValidateINN(inn)
	}
}

func BenchmarkGeneratePhysicalINN(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GeneratePhysicalINN()
	}
}

func BenchmarkGenerateJuridicalINN(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateJuridicalINN()
	}
}
