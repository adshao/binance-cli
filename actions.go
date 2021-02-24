package main

import (
	"log"
	"strings"

	binance "github.com/adshao/go-binance/v2"
	"github.com/juju/errors"
	"github.com/shopspring/decimal"
	"gopkg.in/urfave/cli.v1"
)

var symbols map[string]binance.Symbol

func listBalances(c *cli.Context) error {
	config, err := loadConfig(c)
	if err != nil {
		return errors.Trace(err)
	}
	accountBalances := config.AccountBalances()
	assets := c.StringSlice("assets")
	total := c.Bool("total")

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
		return &AccountConfig{account.Name, l}, nil
	}, func(results map[string]interface{}) (interface{}, error) {
		if !total {
			return results, nil
		}
		totalResults := make(map[string]decimal.Decimal)
		for _, res := range results {
			accountBalance, ok := res.(*AccountConfig)
			if !ok {
				continue
			}
			for _, balance := range accountBalance.Balances {
				free := decimal.RequireFromString(balance.Free)
				locked := decimal.RequireFromString(balance.Locked)
				total := totalResults[balance.Asset].Add(free.Add(locked))
				if balanceMap, ok := accountBalances[accountBalance.Name]; ok {
					if b, ok := balanceMap[balance.Asset]; ok {
						if b.Locked != "" {
							total = total.Sub(decimal.RequireFromString(b.Locked))
						}
					}
				}
				totalResults[balance.Asset] = total
			}
		}
		return []interface{}{results, totalResults}, nil
	})
}

func listOrders(c *cli.Context) error {
	symbol := c.String("symbol")
	all := c.Bool("all")
	limit := c.Int("limit")
	return accountsDo(func(account *Account) (interface{}, error) {
		var orders []*binance.Order
		var err error
		if all {
			if symbol == "" {
				log.Fatal("symbol is required")
			}
			orders, err = account.ListAllOrders(symbol, limit)
		} else {
			orders, err = account.ListOpenOrders(symbol)
		}
		if err != nil {
			return nil, errors.Trace(err)
		}
		return orders, nil
	})
}

func listPrices(c *cli.Context) error {
	symbol := c.String("symbol")
	return runOnce(func(account *Account) (interface{}, error) {
		prices, err := account.ListPrices(symbol)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return prices, nil
	})
}

func cancelOrders(c *cli.Context) error {
	symbol := c.String("symbol")
	orderID := c.Int64("id")
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

func createOrder(c *cli.Context) error {
	config, err := loadConfig(c)
	if err != nil {
		return errors.Trace(err)
	}
	accountBalances := config.AccountBalances()

	symbol := c.String("symbol")
	side := c.String("side")
	quantity := c.String("quantity")
	price := c.String("price")
	isTest := c.Bool("test")
	return accountsDo(
		func(account *Account) (interface{}, error) {
			newQuantity := quantity
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
				minQty := decimal.RequireFromString(lotSize.MinQuantity)
				stepSize := decimal.RequireFromString(lotSize.StepSize)
				precision := info.BaseAssetPrecision

				balances, err := account.ListBalances()
				if err != nil {
					return nil, errors.Trace(err)
				}
				var amount decimal.Decimal
				if side == "SELL" {
					balance, ok := balances[info.BaseAsset]
					if !ok {
						return nil, errors.Errorf("balance %s not found", symbol)
					}
					amount = decimal.RequireFromString(balance.Free)

					if balanceMap, ok := accountBalances[account.Name]; ok {
						if b, ok := balanceMap[info.BaseAsset]; ok {
							if b.Locked != "" {
								amount = amount.Sub(decimal.RequireFromString(b.Locked))
							}
						}
					}

					percent := decimal.NewFromFloat(StrToPct(quantity))
					amount = amount.Mul(percent)
				} else if side == "BUY" {
					balance, ok := balances[info.QuoteAsset]
					if !ok {
						return nil, errors.Errorf("balance %s not found", symbol)
					}
					amount = decimal.RequireFromString(balance.Free)

					if balanceMap, ok := accountBalances[account.Name]; ok {
						if b, ok := balanceMap[info.QuoteAsset]; ok {
							if b.Locked != "" {
								amount = amount.Sub(decimal.RequireFromString(b.Locked))
							}
						}
					}

					percent := decimal.NewFromFloat(StrToPct(quantity))
					amount = amount.Mul(percent)
					p := decimal.RequireFromString(price)
					amount = amount.DivRound(p, int32(precision))
				}
				newQuantity = AmountToLotSize(amount.String(), minQty.String(), stepSize.String(), precision)
			}

			if isTest {
				err := account.TestCreateOrder(symbol, side, newQuantity, price)
				if err != nil {
					return nil, errors.Trace(err)
				}
				return "ok", nil
			}
			res, err := account.CreateOrder(symbol, side, newQuantity, price)
			if err != nil {
				return nil, errors.Trace(err)
			}
			return res.OrderID, nil
		})
}

func listSymbols(c *cli.Context) error {
	symbol := c.String("symbol")
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

func listTrades(c *cli.Context) error {
	symbol := c.String("symbol")
	limit := c.Int("limit")
	return accountsDo(
		func(account *Account) (interface{}, error) {
			trades, err := account.ListTrades(symbol, limit)
			if err != nil {
				return nil, errors.Trace(err)
			}
			return trades, nil
		})
}
