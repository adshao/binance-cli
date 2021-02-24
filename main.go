package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/adshao/go-binance/v2"
	"github.com/juju/errors"
	"gopkg.in/urfave/cli.v1"
)

var (
	name     string
	keyfile  string
	debug    bool
	accounts map[string]*Account
	assets   []string
)

// AccountKey define key info for account
type AccountKey struct {
	Name      string `json:"name"`
	APIKey    string `json:"api_key"`
	SecretKey string `json:"secret_key"`
}

func loadKeys(filePath string) ([]AccountKey, error) {
	keyBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Trace(err)
	}
	var keys []AccountKey
	err = json.Unmarshal(keyBytes, &keys)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return keys, nil
}

func initAccounts() {
	if keyfile == "" {
		keyfile = "keys.json"
	}
	keys, err := loadKeys(keyfile)
	if err != nil {
		log.Fatal("failed to load keys: ", err)
	}
	accounts = make(map[string]*Account)
	for _, key := range keys {
		client := binance.NewClient(
			key.APIKey,
			key.SecretKey,
		)
		if debug {
			client.Debug = true
		}
		account := new(Account)
		account.Client = client
		account.Name = key.Name
		accounts[account.Name] = account
	}
}

var findAccounts func(name string) map[string]*Account

func findAccountsImpl(name string) map[string]*Account {
	initAccounts()
	if name == "" {
		return accounts
	}
	return map[string]*Account{name: accounts[name]}
}

func runOnce(action func(*Account) (interface{}, error),
	postAction ...func(map[string]interface{}) (interface{}, error)) error {

	var origFindAccounts = findAccounts
	defer func() {
		findAccounts = origFindAccounts
	}()
	findAccounts = func(name string) map[string]*Account {
		initAccounts()
		for k, v := range accounts {
			return map[string]*Account{k: v}
		}
		return nil
	}
	return accountsDo(action, postAction...)
}

func accountsDo(action func(*Account) (interface{}, error),
	postAction ...func(map[string]interface{}) (interface{}, error)) error {
	accounts := findAccounts(name)
	var ret interface{}
	var err error
	results := make(map[string]interface{})
	for _, account := range accounts {
		res, err := action(account)
		if err != nil {
			// return errors.Trace(err)
			results[account.Name] = fmt.Sprintf("error: %s", err)
		} else {
			results[account.Name] = res
		}
	}
	if len(postAction) > 0 {
		ret, err = postAction[0](results)
		if err != nil {
			return errors.Trace(err)
		}
	} else {
		ret = results
	}
	return print(ret)
}

func print(ret interface{}) error {
	out, err := json.MarshalIndent(ret, "", "    ")
	if err != nil {
		return errors.Trace(err)
	}
	fmt.Println(string(out))
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "binance-cli"
	app.Usage = "Binance CLI"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "name",
			Usage:       "account name",
			Destination: &name,
		},
		cli.StringFlag{
			Name:        "keyfile",
			Usage:       "file path of api keys",
			Destination: &keyfile,
		},
		cli.BoolFlag{
			Name:        "debug, d",
			Usage:       "show debug info",
			Destination: &debug,
		},
		cli.StringFlag{
			Name:  "configfile, f",
			Usage: "config file",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "list-balance",
			Usage: "list account balances",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:   "assets",
					EnvVar: "BINANCE_ASSETS",
					Usage:  "list balances with asset BTC, BNB ...",
					Value:  &cli.StringSlice{"BTC", "BNB", "USDT"},
				},
				cli.BoolTFlag{
					Name:  "total",
					Usage: "show total balance",
				},
			},
			Action: func(c *cli.Context) error {
				return listBalances(c)
			},
		},
		{
			Name:  "list-price",
			Usage: "list latest price for a symbol or symbols",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "symbol, s",
					Usage: "filter with symbol",
				},
			},
			Action: func(c *cli.Context) error {
				return listPrices(c)
			},
		},
		{
			Name:  "list-order",
			Usage: "list open orders",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "symbol, s",
					Usage: "list orders with symbol",
				},
				cli.BoolFlag{
					Name:  "all",
					Usage: "List all account orders",
				},
				cli.IntFlag{
					Name:  "limit, l",
					Usage: "limit num of trades",
				},
			},
			Action: func(c *cli.Context) error {
				return listOrders(c)
			},
		},
		{
			Name:  "create-order",
			Usage: "create order",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "symbol, s",
					Usage: "symbol name: BNBBTC",
				},
				cli.StringFlag{
					Name:  "side",
					Usage: "side type: SELL or BUY",
				},
				cli.StringFlag{
					Name:  "quantity",
					Usage: "quantity of symbol: 20.120 or 50%",
				},
				cli.StringFlag{
					Name:  "price",
					Usage: "price of symbol",
				},
				cli.BoolFlag{
					Name:  "test",
					Usage: "for test only, will not actually create order",
				},
			},
			Action: func(c *cli.Context) error {
				return createOrder(c)
			},
		},
		{
			Name:  "cancel-order",
			Usage: "cancel open orders",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "symbol, s",
					Usage: "cancel open orders with symbol",
				},
				cli.Int64Flag{
					Name:  "order-id, id",
					Usage: "cancel open order with order id",
				},
			},
			Action: func(c *cli.Context) error {
				return cancelOrders(c)
			},
		},
		{
			Name:  "list-symbol",
			Usage: "list symbols info",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "symbol, s",
					Usage: "symbol name",
				},
			},
			Action: func(c *cli.Context) error {
				return listSymbols(c)
			},
		},
		{
			Name:  "list-trade",
			Usage: "list trades",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "symbol, s",
					Usage: "symbol name",
				},
				cli.IntFlag{
					Name:  "limit, l",
					Usage: "limit num of trades",
				},
			},
			Action: func(c *cli.Context) error {
				return listTrades(c)
			},
		},
		{
			Name:  "list-margin-balance",
			Usage: "list margin account balances",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:   "assets",
					EnvVar: "BINANCE_ASSETS",
					Usage:  "list balances with asset BTC, BNB ...",
				},
				cli.BoolTFlag{
					Name:  "total",
					Usage: "show total balance",
				},
				cli.BoolFlag{
					Name:  "borrowed",
					Usage: "only show borrowed asset",
				},
			},
			Action: func(c *cli.Context) error {
				return listMarginBalances(c)
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
}

func init() {
	findAccounts = findAccountsImpl
}
