package validator

import (
	"regexp"
	"strconv"

	"github.com/go-playground/validator/v10"
)

var (
	regexIsDigit     = regexp.MustCompile(`\D`)
	iso8601DateRegex = regexp.MustCompile(`^(?:[1-9]\d{3}-(?:(?:0[1-9]|1[0-2])-(?:0[1-9]|1\d|2[0-8])|(?:0[13-9]|1[0-2])-(?:29|30)|(?:0[13578]|1[02])-31)|(?:[1-9]\d(?:0[48]|[2468][048]|[13579][26])|(?:[2468][048]|[13579][26])00)-02-29)T(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d(?:\.\d{1,9})?(?:Z|[+-][01]\d:[0-5]\d)$`)
	phoneFormatRegex = regexp.MustCompile(`^\(\d{2}\) \d{4,5}-\d{4}$`)
)

func validateCPForCNPJ(fl validator.FieldLevel) bool {
	number := regexIsDigit.ReplaceAllString(fl.Field().String(), "")

	if len(number) == 11 {
		return validateCPF(fl.Field().String())
	} else if len(number) == 14 {
		return validateCNPJ(number)
	}

	return false
}

func validateCPF(cpf string) bool {
	cpf = regexIsDigit.ReplaceAllString(cpf, "")

	if len(cpf) != 11 {
		return false
	}

	if isAllSameDigits(cpf) {
		return false
	}

	digit1, _ := strconv.Atoi(string(cpf[9]))
	digit2, _ := strconv.Atoi(string(cpf[10]))

	sum := 0
	for i := range 9 {
		digit, _ := strconv.Atoi(string(cpf[i]))
		sum += digit * (10 - i)
	}

	remainder := sum % 11
	if remainder < 2 && digit1 != 0 || remainder >= 2 && digit1 != 11-remainder {
		return false
	}

	sum = 0
	for i := range 10 {
		digit, _ := strconv.Atoi(string(cpf[i]))
		sum += digit * (11 - i)
	}

	remainder = sum % 11
	if remainder < 2 && digit2 != 0 || remainder >= 2 && digit2 != 11-remainder {
		return false
	}

	return true
}

func validateCNPJ(cnpj string) bool {
	cnpj = regexIsDigit.ReplaceAllString(cnpj, "")

	if len(cnpj) != 14 {
		return false
	}

	if isAllSameDigits(cnpj) {
		return false
	}

	var weights1 = [12]int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := range 12 {
		digit, _ := strconv.Atoi(string(cnpj[i]))
		sum += digit * weights1[i]
	}

	remainder := sum % 11
	digit1, _ := strconv.Atoi(string(cnpj[12]))
	if remainder < 2 && digit1 != 0 || remainder >= 2 && digit1 != 11-remainder {
		return false
	}

	var weights2 = [13]int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum = 0
	for i := range 13 {
		digit, _ := strconv.Atoi(string(cnpj[i]))
		sum += digit * weights2[i]
	}

	remainder = sum % 11
	digit2, _ := strconv.Atoi(string(cnpj[13]))
	if remainder < 2 && digit2 != 0 || remainder >= 2 && digit2 != 11-remainder {
		return false
	}

	return true
}

func isAllSameDigits(s string) bool {
	for i := 1; i < len(s); i++ {
		if s[i] != s[0] {
			return false
		}
	}
	return true
}

func isISO8601Date(fl validator.FieldLevel) bool {
	return iso8601DateRegex.MatchString(fl.Field().String())
}

func phoneFormat(fl validator.FieldLevel) bool {
	return phoneFormatRegex.MatchString(fl.Field().String())
}
