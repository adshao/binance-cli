package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/adshao/go-binance"
	"github.com/juju/errors"
	"gopkg.in/urfave/cli.v1"
)

var config = new(Config)

// Config define cli config
type Config struct {
	Accounts []AccountConfig `json:"accounts"`
}

// AccountBalances return account balance map
func (c *Config) AccountBalances() map[string]map[string]binance.Balance {
	accountMap := make(map[string]map[string]binance.Balance)
	for _, info := range c.Accounts {
		accountMap[info.Name] = make(map[string]binance.Balance)
		for _, balance := range info.Balances {
			accountMap[info.Name][balance.Asset] = balance
		}
	}
	return accountMap
}

// AccountConfig define account config
type AccountConfig struct {
	Name     string            `json:"name"`
	Balances []binance.Balance `json:"balances"`
}

func loadConfig(c *cli.Context) (*Config, error) {
	if !c.GlobalIsSet("configfile") {
		return config, nil
	}
	keyBytes, err := ioutil.ReadFile(c.GlobalString("configfile"))
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = json.Unmarshal(keyBytes, config)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return config, nil
}
