package evm

import (
	"strconv"
	"strings"
)

func DecimalToHex(number int64) string {
	return "0x" + strconv.FormatInt(number, 16)
}

func hexNumberToString(hexString string) string {
	// replace 0x or 0X with empty String
	numberStr := strings.ReplaceAll(hexString, "0x", "")
	numberStr = strings.ReplaceAll(numberStr, "0X", "")
	return numberStr
}

func HexToInt(hex string) (int64, error) {
	return strconv.ParseInt(hexNumberToString(hex), 16, 64)
}
