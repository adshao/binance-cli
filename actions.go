package main

import (
	"strconv"
	"strings"

	"github.com/adshao/go-binance"
	"github.com/juju/errors"
)

var symbols map[string]binance.Symbol

func listBalances(assets []string, total bool) error {
	return accountsDo(func(account *Account) (interface{}, error) {
		balances, err := account.ListBalances()
		if err != nil {
			return nil, errors.Trace(err)
		}
		var l []binance.Balance
		if len(assets) > 0 {
			var insertedKeys []string
			for _, asset := range assets {
				if b, ok := balances[asset]; ok && !StrContains(insertedKeys, asset) {
					l = append(l, b)
					insertedKeys = append(insertedKeys, asset)
				}
			}
		} else {
			for _, b := range balances {
				l = append(l, b)
			}
		}
		return l, nil
	}, func(results map[string]interface{}) (interface{}, error) {
		if !total {
			return results, nil
		}
		totalResults := make(map[string]float64)
		for _, res := range results {
			assets, ok := res.([]binance.Balance)
			if !ok {
				continue
			}
			for _, asset := range assets {
				free, _ := strconv.ParseFloat(asset.Free, 64)
				locked, _ := strconv.ParseFloat(asset.Locked, 64)
				totalResults[asset.Asset] += free + locked
			}
		}
		return []interface{}{results, totalResults}, nil
	})
}

func listOpenOrders(symbol string) error {
	return accountsDo(func(account *Account) (interface{}, error) {
		orders, err := account.ListOpenOrders(symbol)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return orders, nil
	})
}

func listPrices(symbol string) error {
	return runOnce(func(account *Account) (interface{}, error) {
		prices, err := account.ListPrices(symbol)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return prices, nil
	})
}

func cancelOrders(symbol string, orderID int64) error {
	return accountsDo(
		func(account *Account) (interface{}, error) {
			var canceledOrders []int64
			var cancelingOrders []int64
			if orderID == 0 {
				orders, err := account.ListOpenOrders(symbol)
				if err != nil {
					return nil, errors.Trace(err)
				}
				for _, order := range orders {
					cancelingOrders = append(cancelingOrders, order.OrderID)
				}
			} else {
				cancelingOrders = []int64{orderID}
			}
			for _, orderID := range cancelingOrders {
				err := account.CancelOrder(symbol, orderID)
				if err != nil {
					return nil, errors.Trace(err)
				}
				canceledOrders = append(canceledOrders, orderID)
			}
			return canceledOrders, nil
		})
}

func (account *Account) loadSymbols() error {
	var err error
	if symbols == nil {
		symbols, err = account.ListSymbols()
		if err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func createOrder(symbol, side, quantity, price string, isTest bool) error {
	return accountsDo(
		func(account *Account) (interface{}, error) {
			if strings.HasSuffix(quantity, "%") {
				err := account.loadSymbols()
				if err != nil {
					return nil, errors.Trace(err)
				}
				info, ok := symbols[symbol]
				if !ok {
					return nil, errors.Errorf("symbol %s not found", symbol)
				}
				lotSize := info.LotSizeFilter()
				if lotSize == nil {
					return nil, errors.Trace(err)
				}
				minQty := StrToAmount(lotSize.MinQuantity)
				stepSize := StrToAmount(lotSize.StepSize)
				// precision := info.BaseAssetPrecision

				balances, err := account.ListBalances()
				if err != nil {
					return nil, errors.Trace(err)
				}
				var amount float64
				if side == "SELL" {
					balance, ok := balances[info.BaseAsset]
					if !ok {
						return nil, errors.Errorf("balance %s not found", symbol)
					}
					amount = StrToAmount(balance.Free)
					amount *= StrToPct(quantity)
				} else if side == "BUY" {
					balance, ok := balances[info.QuoteAsset]
					if !ok {
						return nil, errors.Errorf("balance %s not found", symbol)
					}
					amount = StrToAmount(balance.Free)
					amount *= StrToPct(quantity)
					amount /= StrToAmount(price)
				}
				amount = AmountToLotSize(amount, minQty, stepSize)
				quantity = AmountToStr(amount)
			}

			if isTest {
				err := account.TestCreateOrder(symbol, side, quantity, price)
				if err != nil {
					return nil, errors.Trace(err)
				}
				return "ok", nil
			}
			res, err := account.CreateOrder(symbol, side, quantity, price)
			if err != nil {
				return nil, errors.Trace(err)
			}
			return res.OrderID, nil
		})
}

func listSymbols(symbol string) error {
	return runOnce(
		func(account *Account) (interface{}, error) {
			symbols, err := account.ListSymbols()
			if err != nil {
				return nil, errors.Trace(err)
			}
			if symbol != "" {
				s, ok := symbols[symbol]
				if !ok {
					return nil, errors.Errorf("symbol %s not found", symbol)
				}
				return s, nil
			}
			return symbols, nil
		})
}
