package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAmountToLotSize(t *testing.T) {
	assert := assert.New(t)
	type args struct {
		minQty   float64
		stepSize float64
		amount   float64
	}
	tests := []struct {
		name   string
		args   args
		expect float64
	}{
		{
			name: "test with lot of zero and invalid amount",
			args: args{
				minQty:   0.01,
				stepSize: 0.01,
				amount:   0.001,
			},
			expect: 0,
		},
		{
			name: "test with lot",
			args: args{
				minQty:   0.01,
				stepSize: 0.01,
				amount:   1.39,
			},
			expect: 1.39,
		},
		{
			name: "test with big decimal",
			args: args{
				minQty:   0.01,
				stepSize: 0.02,
				amount:   11.31232419283240912834434,
			},
			expect: 11.31,
		},
		{
			name: "test with big number",
			args: args{
				minQty:   0.0001,
				stepSize: 0.02,
				amount:   11232821093480213.31232419283240912834434,
			},
			expect: 11232821093480213.3001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(tt.expect, AmountToLotSize(tt.args.amount, tt.args.minQty, tt.args.stepSize), 0.00000000000001)
		})
	}
}

func TestAmountToStr(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name      string
		amount    float64
		precision int
		expect    string
	}{
		{
			name:      "test without precision",
			amount:    0.005,
			precision: 0,
			expect:    "0.005",
		},
		{
			name:      "test with precision",
			amount:    10.0005,
			precision: 5,
			expect:    "10.0005",
		},
		{
			name:      "test with precision trunc",
			amount:    10.123456789,
			precision: 6,
			expect:    "10.123456",
		},
		{
			name:      "test with exact precision",
			amount:    10.123456,
			precision: 6,
			expect:    "10.123456",
		},
		// {
		// 	name:      "test with more precision",
		// 	amount:    2.32,
		// 	precision: 8,
		// 	expect:    "2.32",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.precision != 0 {
				assert.Equal(tt.expect, AmountToStr(tt.amount, tt.precision))
			} else {
				assert.Equal(tt.expect, AmountToStr(tt.amount))
			}
		})
	}
}

func TestStrToAmount(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name   string
		s      string
		expect float64
	}{
		{
			name:   "test with zero suffix",
			s:      "0.0050000",
			expect: 0.005,
		},
		{
			name:   "test with exact float",
			s:      "10.000000012",
			expect: 10.000000012,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(tt.expect, StrToAmount(tt.s), 0.00000000000001)
		})
	}
}

func TestStrToPct(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name   string
		s      string
		expect float64
	}{
		{
			name:   "test with percentage sign",
			s:      "81%",
			expect: 0.81,
		},
		{
			name:   "test with float",
			s:      "0.12",
			expect: 0.12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(tt.expect, StrToPct(tt.s), 0.00000000000001)
		})
	}
}
