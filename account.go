package main

import (
	"context"
	"strings"
	"time"

	"github.com/adshao/go-binance"
	"github.com/juju/errors"
)

func newContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	return context.WithTimeout(ctx, 10*time.Second)
}

// Account define binance account
type Account struct {
	*binance.Client
	Name     string            `json:"name"`
	Balances []binance.Balance `json:"balances"`
}

// UpdateBalances update account balances
func (account *Account) UpdateBalances(assets []string) error {
	ctx, cancel := newContext()
	defer cancel()
	res, err := account.NewGetAccountService().Do(ctx)
	if err != nil {
		return errors.Trace(err)
	}
	if len(assets) > 0 {
		var insertedKeys []string
		var balances []binance.Balance
		for _, asset := range assets {
			for _, balance := range res.Balances {
				if asset == balance.Asset && !StrContains(insertedKeys, asset) {
					balances = append(balances, balance)
					insertedKeys = append(insertedKeys, asset)
				}
			}
		}
		account.Balances = balances
	} else {
		account.Balances = res.Balances
	}
	return nil
}

// ListOpenOrders list open orders
func (account *Account) ListOpenOrders(symbol string) ([]*binance.Order, error) {
	ctx, cancel := newContext()
	defer cancel()
	service := account.NewListOpenOrdersService()
	if symbol != "" {
		service = service.Symbol(symbol)
	}
	orders, err := service.Do(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return orders, nil
}

// ListPrices list latest prices for a symbol or symbols
func (account *Account) ListPrices(symbol string) ([]*binance.SymbolPrice, error) {
	ctx, cancel := newContext()
	defer cancel()
	service := account.NewListPricesService()
	if symbol != "" {
		service = service.Symbol(symbol)
	}
	prices, err := service.Do(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return prices, nil
}

// CancelOrder cancel open order
func (account *Account) CancelOrder(symbol string, orderID int64) error {
	ctx, cancel := newContext()
	defer cancel()
	_, err := account.NewCancelOrderService().Symbol(symbol).OrderID(orderID).Do(ctx)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// CreateOrder create order
func (account *Account) CreateOrder(symbol, side, quantity, price string) (*binance.CreateOrderResponse, error) {
	ctx, cancel := newContext()
	defer cancel()
	side = strings.ToUpper(side)
	sideType := binance.SideType(side)
	res, err := account.NewCreateOrderService().Symbol(symbol).Side(sideType).
		Quantity(quantity).Price(price).Type(binance.OrderTypeLimit).
		TimeInForce(binance.TimeInForceTypeGTC).Do(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return res, nil
}
