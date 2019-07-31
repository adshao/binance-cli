### binance-cli

Binance CLI
币安交易所命令行工具

Manipulate multiple accounts with one command!
支持批量操作多账号！

### Installation

```shell
go install github.com/adshao/binance-cli
```

### Prepare key file

save api/secret keys into keys.json
```json
[
    {
        "name": "demo",
        "api_key": "xxxx",
        "secret_key": "xxx"
    },
    {
    }
]
```

### Run CLI

use ```-h``` to get help.

```shell
./binance-cli -h

NAME:
   binance-cli - Binance CLI

USAGE:
   binance-cli [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     list-balance  list account balances
     list-price    list latest price for a symbol or symbols
     list-order    list open orders
     create-order  create order
     cancel-order  cancel open orders
     list-symbol   list symbols info
     help, h       Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --name value     account name
   --keyfile value  file path of api keys
   --debug, -d      show debug info
   --help, -h       show help
   --version, -v    print the version
```

#### Check Latest Price

```shell
./binance-cli list-price --symbol BNBBTC

{
    "test1": [
        {
            "symbol": "BNBBTC",
            "price": "0.00283210"
        }
    ]
}
```

#### List Balances

```shell
./binance-cli list-balance

[
    {
        "test1": [
            {
                "asset": "BNB",
                "free": "2027.68758027",
                "locked": "1000.00000000"
            },
            {
                "asset": "BTC",
                "free": "0.00001550",
                "locked": "0.00000000"
            }
        ],
        "test2": [
            {
                "asset": "BNB",
                "free": "300.00000000",
                "locked": "0.00000000"
            },
            {
                "asset": "BTC",
                "free": "0.00000000",
                "locked": "0.00000000"
            }
        ],
        "test3": [
            {
                "asset": "BNB",
                "free": "603.98788625",
                "locked": "0.00000000"
            },
            {
                "asset": "BTC",
                "free": "0.00881320",
                "locked": "0.00000000"
            }
        ]
    },
    {
        "BNB": 3931.6754665199996,
        "BTC": 0.0088287
    }
]
```
