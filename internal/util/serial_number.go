package util

import "fmt"

// GenerateSerialNumber creates a formatted serial number based on a count and a specified prefix type.
func GenerateSerialNumber(count int, paramType string) string {
	var serialNumber string
	adjustedCount := count - 1
	blockIndex := adjustedCount / 999

	letter := string('A' + rune(blockIndex))

	numericalPart := (adjustedCount % 999) + 1
	formatedNumericPart := fmt.Sprintf("%03d", numericalPart)

	switch paramType {
	case "p":
		serialNumber = fmt.Sprintf("P_%s%s", letter, formatedNumericPart)
	case "ap":
		serialNumber = fmt.Sprintf("AP_%s%s", letter, formatedNumericPart)
	case "f":
		serialNumber = fmt.Sprintf("F_%s%s", letter, formatedNumericPart)
	case "e":
		serialNumber = fmt.Sprintf("E_%s%s", letter, formatedNumericPart)
	}

	return serialNumber
}
