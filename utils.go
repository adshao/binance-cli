package main

import (
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// StrContains check if string items contains s
func StrContains(items []string, s string) bool {
	for _, item := range items {
		if item == s {
			return true
		}
	}
	return false
}

// AmountToLotSize convert amoutn to lot size
func AmountToLotSize(amount, minQty, stepSize string, precision int) string {
	amountDec := decimal.RequireFromString(amount)
	minQtyDec := decimal.RequireFromString(minQty)
	baseAmountDec := amountDec.Sub(minQtyDec)
	if baseAmountDec.LessThan(decimal.RequireFromString("0")) {
		return "0"
	}
	stepSizeDec := decimal.RequireFromString(stepSize)
	baseAmountDec = baseAmountDec.Div(stepSizeDec).Truncate(0).Mul(stepSizeDec)
	return baseAmountDec.Add(minQtyDec).Truncate(int32(precision)).String()
}

// StrToPct convert string to percentage
func StrToPct(s string) float64 {
	if strings.HasSuffix(s, "%") {
		s = strings.TrimRight(s, "%")
		v, _ := strconv.ParseFloat(s, 64)
		return v / float64(100)
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
