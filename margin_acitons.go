package main

import (
	binance "github.com/adshao/go-binance/v2"
	"github.com/juju/errors"
	"github.com/shopspring/decimal"
	"gopkg.in/urfave/cli.v1"
)

// MarginAccount define margin account
type MarginAccount struct {
	Name   string
	Margin *binance.MarginAccount `json:"margin"`
}

// MarginAssets get assets of margin account
func (account *MarginAccount) MarginAssets() map[string]binance.UserAsset {
	m := make(map[string]binance.UserAsset)
	if account.Margin == nil {
		return m
	}
	for _, userAsset := range account.Margin.UserAssets {
		m[userAsset.Asset] = userAsset
	}
	return m
}

func listMarginBalances(c *cli.Context) error {
	assets := c.StringSlice("assets")
	total := c.Bool("total")
	filterBorrowed := c.Bool("borrowed")

	return accountsDo(func(account *Account) (interface{}, error) {
		marginAccount, err := account.GetMarginAccount()
		if err != nil {
			return nil, errors.Trace(err)
		}
		a := &MarginAccount{Name: account.Name, Margin: marginAccount}
		marginAssets := a.MarginAssets()
		var l []binance.UserAsset
		if len(assets) > 0 {
			var insertedKeys []string
			for _, asset := range assets {
				if b, ok := marginAssets[asset]; ok && !StrContains(insertedKeys, asset) {
					if !c.IsSet("borrowed") ||
						(filterBorrowed && decimal.RequireFromString(b.Borrowed).GreaterThan(decimal.RequireFromString("0"))) ||
						(!filterBorrowed && decimal.RequireFromString(b.Borrowed).Equal(decimal.RequireFromString("0"))) {
						l = append(l, b)
						insertedKeys = append(insertedKeys, asset)
					}
				}
			}
		} else {
			for _, b := range marginAssets {
				if !c.IsSet("borrowed") ||
					(filterBorrowed && decimal.RequireFromString(b.Borrowed).GreaterThan(decimal.RequireFromString("0"))) ||
					(!filterBorrowed && decimal.RequireFromString(b.Borrowed).Equal(decimal.RequireFromString("0"))) {
					l = append(l, b)
				}
			}
		}
		a.Margin.UserAssets = l
		return a, nil
	}, func(results map[string]interface{}) (interface{}, error) {
		if !total {
			return results, nil
		}
		totalAssetOfBTC := decimal.Decimal{}
		totalLiabilityOfBTC := decimal.Decimal{}
		totalNetAssetOfBTC := decimal.Decimal{}
		m := make(map[string]map[string]decimal.Decimal)
		for _, res := range results {
			account, ok := res.(*MarginAccount)
			if !ok {
				continue
			}
			totalAssetOfBTC = totalAssetOfBTC.Add(decimal.RequireFromString(account.Margin.TotalAssetOfBTC))
			totalLiabilityOfBTC = totalLiabilityOfBTC.Add(decimal.RequireFromString(account.Margin.TotalLiabilityOfBTC))
			totalNetAssetOfBTC = totalNetAssetOfBTC.Add(decimal.RequireFromString(account.Margin.TotalNetAssetOfBTC))
			for _, asset := range account.Margin.UserAssets {
				borrowed := decimal.RequireFromString(asset.Borrowed)
				free := decimal.RequireFromString(asset.Free)
				interest := decimal.RequireFromString(asset.Interest)
				locked := decimal.RequireFromString(asset.Locked)
				netAsset := decimal.RequireFromString(asset.NetAsset)
				userAsset, ok := m[asset.Asset]
				if !ok {
					userAsset = map[string]decimal.Decimal{
						"borrowed": decimal.Decimal{},
						"free":     decimal.Decimal{},
						"interest": decimal.Decimal{},
						"locked":   decimal.Decimal{},
						"netAsset": decimal.Decimal{},
					}
					m[asset.Asset] = userAsset
				}
				userAsset["borrowed"] = userAsset["borrowed"].Add(borrowed)
				userAsset["free"] = userAsset["free"].Add(free)
				userAsset["interest"] = userAsset["interest"].Add(interest)
				userAsset["locked"] = userAsset["locked"].Add(locked)
				userAsset["netAsset"] = userAsset["netAsset"].Add(netAsset)
			}
		}
		res := map[string]interface{}{
			"TotalAssetOfBTC":     totalAssetOfBTC,
			"TotalLiabilityOfBTC": totalLiabilityOfBTC,
			"TotalNetAssetOfBTC":  totalNetAssetOfBTC,
			"UserAssets":          m,
		}
		return []interface{}{results, res}, nil
	})
}
