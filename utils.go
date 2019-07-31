package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
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

// StrToAmount convert amount in string to float
func StrToAmount(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// AmountToStr convert amount in float to string
func AmountToStr(a float64, precision ...int) string {
	if len(precision) > 0 {
		prc := precision[0]
		a = math.Trunc(a*math.Pow10(prc)) / math.Pow10(prc)
	}
	return fmt.Sprintf("%v", a)
}

// AmountToLotSize convert amoutn to lot size
func AmountToLotSize(amount, minQty, stepSize float64) float64 {
	baseAmount := amount - minQty
	if baseAmount < 0 {
		return 0
	}
	baseAmount = float64(int(baseAmount/stepSize)) * stepSize
	return baseAmount + minQty
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
