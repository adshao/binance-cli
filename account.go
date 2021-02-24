package main

import (
	"context"
	"strings"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/juju/errors"
)

func newContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	return context.WithTimeout(ctx, 30*time.Second)
}

// Account define binance account
type Account struct {
	*binance.Client
	Name     string            `json:"name"`
	Balances []binance.Balance `json:"balances"`
}

// ListBalances update account balances
func (account *Account) ListBalances() (map[string]binance.Balance, error) {
	ctx, cancel := newContext()
	defer cancel()
	res, err := account.NewGetAccountService().Do(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}
	balances := make(map[string]binance.Balance)
	for _, balance := range res.Balances {
		balances[balance.Asset] = balance
	}
	return balances, nil
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

// ListAllOrders list all account orders
func (account *Account) ListAllOrders(symbol string, limit int) ([]*binance.Order, error) {
	ctx, cancel := newContext()
	defer cancel()
	service := account.NewListOrdersService()
	if symbol != "" {
		service = service.Symbol(symbol)
	}
	if limit != 0 {
		service = service.Limit(limit)
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

// TestCreateOrder create order for test
func (account *Account) TestCreateOrder(symbol, side, quantity, price string) error {
	ctx, cancel := newContext()
	defer cancel()
	side = strings.ToUpper(side)
	sideType := binance.SideType(side)
	err := account.NewCreateOrderService().Symbol(symbol).Side(sideType).
		Quantity(quantity).Price(price).Type(binance.OrderTypeLimit).
		TimeInForce(binance.TimeInForceTypeGTC).Test(ctx)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// ListSymbols list symbols
func (account *Account) ListSymbols() (map[string]binance.Symbol, error) {
	ctx, cancel := newContext()
	defer cancel()
	info, err := account.NewExchangeInfoService().Do(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}
	ret := make(map[string]binance.Symbol)
	symbols := info.Symbols
	for _, symbol := range symbols {
		ret[symbol.Symbol] = symbol
	}
	return ret, nil
}

// ListTrades list trades
func (account *Account) ListTrades(symbol string, limit int) ([]*binance.TradeV3, error) {
	ctx, cancel := newContext()
	defer cancel()
	service := account.NewListTradesService()
	if symbol != "" {
		service = service.Symbol(symbol)
	}
	if limit != 0 {
		service = service.Limit(limit)
	}
	trades, err := service.Do(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return trades, nil
}

// GetMarginAccount get margin account
func (account *Account) GetMarginAccount() (*binance.MarginAccount, error) {
	ctx, cancel := newContext()
	defer cancel()
	service := account.NewGetMarginAccountService()
	marginAccount, err := service.Do(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return marginAccount, nil
}
