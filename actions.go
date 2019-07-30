package main

import (
	"strconv"

	"github.com/adshao/go-binance"
	"github.com/juju/errors"
)

func listBalances(assets []string, total bool) error {
	return accountsDo(func(account *Account) (interface{}, error) {
		err := account.UpdateBalances(assets)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return account.Balances, nil
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

func cancelOrders(symbol string) error {
	return accountsDo(
		func(account *Account) (interface{}, error) {
			var canceledOrders []int64
			orders, err := account.ListOpenOrders(symbol)
			if err != nil {
				return nil, errors.Trace(err)
			}
			for _, order := range orders {
				err = account.CancelOrder(symbol, order.OrderID)
				if err != nil {
					return nil, errors.Trace(err)
				}
				canceledOrders = append(canceledOrders, order.OrderID)
			}
			return canceledOrders, nil
		})
}

func createOrder(symbol, side, quantity, price string) error {
	return accountsDo(
		func(account *Account) (interface{}, error) {
			var orderIDs []int64
			res, err := account.CreateOrder(symbol, side, quantity, price)
			if err != nil {
				return nil, errors.Trace(err)
			}
			orderIDs = append(orderIDs, res.OrderID)
			return orderIDs, nil
		})
}
