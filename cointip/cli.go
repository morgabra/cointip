package main

import (
	logger "log"
	"os"

	"github.com/urfave/cli"

	"github.com/morgabra/cointip"
)

var log *logger.Logger = logger.New(os.Stdout, "", 0)

var apiKey, apiSecret string

func printAccount(acct *cointip.Account) {
	log.Printf(
		"%s %s %s:%.8f %s:%.2f\n",
		acct.ID, acct.Name, acct.Balance.Currency, acct.Balance.Amount,
		acct.NativeBalance.Currency, acct.NativeBalance.Amount)
}

func printTransaction(tx *cointip.Transaction) {
	log.Printf(
		"%s %s %s:%.8f %s:%.2f\n",
		tx.ID, tx.Status, tx.Amount.Currency, tx.Amount.Amount,
		tx.NativeAmount.Currency, tx.NativeAmount.Amount)
}

func makeClient(ctx *cli.Context) *cointip.ApiKeyClient {

	if apiKey == "" {
		log.Fatal("Missing required argument 'api-key'")
	}

	if apiSecret == "" {
		log.Fatal("Missing required argument 'api-secret'")
	}

	c, err := cointip.APIKeyClient(apiKey, apiSecret)
	if err != nil {
		log.Fatal(err)
	}

	return c
}

func main() {

	app := cli.NewApp()
	app.Name = "cointip"
	app.Version = "0.0.1"
	app.Usage = "Create accounts and move currency around via the Coinbase API."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "api-key",
			Usage:       "Coinbase API key.",
			EnvVar:      "COINBASE_KEY",
			Destination: &apiKey,
		},
		cli.StringFlag{
			Name:        "api-secret",
			Usage:       "Coinbase API secret.",
			EnvVar:      "COINBASE_SECRET",
			Destination: &apiSecret,
		},
	}

	app.Commands = []cli.Command{
		ListAccounts,
		GetAccount,
		CreateAccount,
		DeleteAccount,
		CreateAddress,
		Transfer,
		Withdraw,
		GetTransaction,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

var ListAccounts = cli.Command{
	Name:  "list-accounts",
	Usage: "List accounts",
	Action: func(ctx *cli.Context) error {
		c := makeClient(ctx)

		accounts, err := c.ListAccounts()
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		for _, acct := range accounts {
			printAccount(acct)
		}

		return nil
	},
}

var GetAccount = cli.Command{
	Name:  "get-account",
	Usage: "Get account",
	Action: func(ctx *cli.Context) error {
		c := makeClient(ctx)

		if len(ctx.Args()) != 1 {
			log.Fatal("Missing required argument: AccountID")
		}

		accountID := ctx.Args()[0]

		account, err := c.GetAccount(accountID)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		printAccount(account)

		return nil
	},
}

var CreateAccount = cli.Command{
	Name:  "create-account",
	Usage: "Create account",
	Action: func(ctx *cli.Context) error {
		c := makeClient(ctx)

		if len(ctx.Args()) != 1 {
			log.Fatal("Missing required argument: 'Account Name'")
		}

		accountName := ctx.Args()[0]

		account, err := c.CreateAccount(accountName)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		printAccount(account)

		return nil
	},
}

var DeleteAccount = cli.Command{
	Name:  "delete-account",
	Usage: "Delete account",
	Action: func(ctx *cli.Context) error {
		c := makeClient(ctx)

		if len(ctx.Args()) != 1 {
			log.Fatal("Missing required argument: AccountID")
		}

		accountID := ctx.Args()[0]

		err := c.DeleteAccount(accountID)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		log.Printf("deleted account %s\n", accountID)

		return nil
	},
}

var CreateAddress = cli.Command{
	Name:  "create-address",
	Usage: "Create an address for receiving funds",
	Action: func(ctx *cli.Context) error {
		c := makeClient(ctx)

		if len(ctx.Args()) != 1 {
			log.Fatal("Missing required argument: AccountID")
		}

		accountID := ctx.Args()[0]

		addr, err := c.CreateAddress(accountID)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		log.Printf("%s\n", addr.Address)

		return nil
	},
}

var Transfer = cli.Command{
	Name:  "transfer",
	Usage: "Transfer funds between accounts",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "from",
			Usage: "Account ID to transfer FROM",
		},
		cli.StringFlag{
			Name:  "to",
			Usage: "Account ID to transfer TO",
		},
		cli.StringFlag{
			Name:  "currency",
			Usage: "Currency type to transfer",
		},
		cli.Float64Flag{
			Name:  "amount",
			Usage: "Amount to transfer",
		},
	},
	Action: func(ctx *cli.Context) error {
		c := makeClient(ctx)

		if !ctx.IsSet("from") {
			log.Fatal("Missing required flag: --from")
		}
		if !ctx.IsSet("to") {
			log.Fatal("Missing required flag: --to")
		}
		if !ctx.IsSet("currency") {
			log.Fatal("Missing required flag: --currency")
		}
		if !ctx.IsSet("amount") {
			log.Fatal("Missing required flag: --amount")
		}

		from := ctx.String("from")
		to := ctx.String("to")
		currency := ctx.String("currency")
		amount := ctx.Float64("amount")

		tx, err := c.Transfer(from, to, &cointip.Balance{Amount: amount, Currency: currency})
		if err != nil {
			log.Fatalf("Error: %s", err)
		}

		printTransaction(tx)

		return nil
	},
}

var Withdraw = cli.Command{
	Name:  "withdraw",
	Usage: "Withdraw funds to a BTC address",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "from",
			Usage: "Account ID to transfer FROM",
		},
		cli.StringFlag{
			Name:  "to",
			Usage: "BTC address to transfer TO",
		},
		cli.StringFlag{
			Name:  "currency",
			Usage: "Currency type to transfer",
		},
		cli.Float64Flag{
			Name:  "amount",
			Usage: "Amount to transfer",
		},
	},
	Action: func(ctx *cli.Context) error {
		c := makeClient(ctx)

		if !ctx.IsSet("from") {
			log.Fatal("Missing required flag: --from")
		}
		if !ctx.IsSet("to") {
			log.Fatal("Missing required flag: --to")
		}
		if !ctx.IsSet("currency") {
			log.Fatal("Missing required flag: --currency")
		}
		if !ctx.IsSet("amount") {
			log.Fatal("Missing required flag: --amount")
		}

		from := ctx.String("from")
		to := ctx.String("to")
		currency := ctx.String("currency")
		amount := ctx.Float64("amount")

		tx, err := c.Withdraw(from, to, &cointip.Balance{Amount: amount, Currency: currency})
		if err != nil {
			log.Fatalf("Error: %s", err)
		}

		printTransaction(tx)

		return nil
	},
}

var GetTransaction = cli.Command{
	Name:  "get-transaction",
	Usage: "Show a transaction",
	Action: func(ctx *cli.Context) error {
		c := makeClient(ctx)

		if len(ctx.Args()) != 2 {
			log.Fatal("Missing required argument(s): AccountID TransactionID")
		}

		accountID := ctx.Args()[0]
		txID := ctx.Args()[1]

		tx, err := c.GetTransaction(accountID, txID)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}

		printTransaction(tx)

		return nil
	},
}
