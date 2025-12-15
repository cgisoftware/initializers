package validator

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestValidateCNPJ(t *testing.T) {
	tests := []struct {
		name     string
		cnpj     string
		expected bool
	}{
		{"Valid CNPJ", "11.444.777/0001-61", true},
		{"Valid CNPJ numbers only", "11444777000161", true},
		{"Invalid CNPJ length", "1234567890123", false},
		{"Invalid CNPJ length", "123456789012345", false},
		{"Invalid CNPJ check digits", "11.444.777/0001-62", false}, // Adjusted check digit
		{"Invalid CNPJ zero", "00.000.000/0000-00", false},
		{"Invalid CNPJ repeated", "11.111.111/1111-11", false},
		{"Invalid characters", "11.444.777/0001-XX", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateCNPJ(tt.cnpj); got != tt.expected {
				t.Errorf("validateCNPJ(%v) = %v, want %v", tt.cnpj, got, tt.expected)
			}
		})
	}
}

func TestValidateCPF(t *testing.T) {
	tests := []struct {
		name     string
		cpf      string
		expected bool
	}{
		{"Valid CPF", "123.456.789-09", true},
		{"Valid CPF numbers only", "12345678909", true},
		{"Invalid CPF length", "1234567890", false},
		{"Invalid CPF length", "123456789012", false},
		{"Invalid CPF checksum", "123.456.789-10", false},
		{"Invalid CPF Repeated", "111.111.111-11", false},
		{"Invalid CPF Zero", "000.000.000-00", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateCPF(tt.cpf); got != tt.expected {
				t.Errorf("validateCPF(%v) = %v, want %v", tt.cpf, got, tt.expected)
			}
		})
	}
}

func TestPhoneFormat(t *testing.T) {
	// Need to mock field level
	validate := validator.New()
	validate.RegisterValidation("PhoneFormat", phoneFormat)

	tests := []struct {
		name     string
		phone    string
		expected bool
	}{
		{"Valid Mobile", "(11) 91234-5678", true},
		{"Valid Landline", "(11) 1234-5678", true},
		{"Invalid Format", "11912345678", false},
		{"Invalid Format", "(11)91234-5678", false}, // Missing space
		{"Invalid Length", "(11) 123-5678", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Var(tt.phone, "PhoneFormat")
			if tt.expected && err != nil {
				t.Errorf("phoneFormat(%v) expected valid, got error: %v", tt.phone, err)
			}
			if !tt.expected && err == nil {
				t.Errorf("phoneFormat(%v) expected invalid, got nil", tt.phone)
			}
		})
	}
}

func TestIsISO8601Date(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("ISO8601date", isISO8601Date)

	tests := []struct {
		name     string
		date     string
		expected bool
	}{
		{"Valid Date Z", "2023-10-01T12:00:00Z", true},
		{"Valid Date Offset", "2023-10-01T12:00:00+03:00", true},
		{"Valid Date Millis", "2023-10-01T12:00:00.123Z", true},
		{"Invalid Date", "2023/10/01", false},
		{"Invalid Time", "2023-10-01T25:00:00Z", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Var(tt.date, "ISO8601date")
			if tt.expected && err != nil {
				t.Errorf("isISO8601Date(%v) expected valid, got error: %v", tt.date, err)
			}
			if !tt.expected && err == nil {
				t.Errorf("isISO8601Date(%v) expected invalid, got nil", tt.date)
			}
		})
	}
}

func TestValidateCPForCNPJ(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("CPForCNPJ", validateCPForCNPJ)

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"Valid CNPJ", "11.444.777/0001-61", true},
		{"Valid CPF", "123.456.789-09", true},
		{"Invalid Checksum", "11.444.777/0001-62", false},
		{"Invalid Checksum CPF", "123.456.789-10", false},
		{"Ignore other lengths", "12345", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Var(tt.value, "CPForCNPJ")
			if tt.expected && err != nil {
				t.Errorf("validateCPForCNPJ(%v) expected valid, got error: %v", tt.value, err)
			}
			if !tt.expected && err == nil {
				t.Errorf("validateCPForCNPJ(%v) expected invalid, got nil", tt.value)
			}
		})
	}

}
