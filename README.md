cointip
-------

Dumbest possible coinbase API client for funzies.

You probably shouldn't use this with money you aren't willing to lose.

1. Make a coinbase account with an API key with permissions to do wallet stuff.
2. Do stuff.

```
$ cointip
NAME:
   cointip - Create accounts and move currency around via the Coinbase API.

USAGE:
   cointip [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
     list-accounts    List accounts
     get-account      Get account
     create-account   Create account
     delete-account   Delete account
     create-address   Create an address for receiving funds
     transfer         Transfer funds between accounts
     withdraw         Withdraw funds to a BTC address
     get-transaction  Show a transaction
     help, h          Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --api-key value     Coinbase API key. [$COINBASE_KEY]
   --api-secret value  Coinbase API secret. [$COINBASE_SECRET]
   --help, -h          show help
   --version, -v       print the version

$ export COINBASE_KEY=lol
$ export COINBASE_SECRET=lololol

$ cointip list-accounts
6a06ef0e-306a-5809-a9ae-b61ebf21b4cd Test Wallet 2 BTC:0.00000000 USD:0.00
f670f7b0-9ebd-5b09-92ea-56758fbb1fd3 Test Wallet BTC:0.00014401 USD:0.26
7fe6e042-02ff-5f39-99b6-3768f00b8a9e USD Wallet USD:0.00000000 USD:0.00
6a447903-eec4-5499-8137-38a7854ed1ff LTC Wallet LTC:0.00000000 USD:0.00
0f9f0dd6-8532-5f81-8ea0-e44b858feba7 ETH Wallet ETH:0.00000000 USD:0.00
104df44b-6238-5c84-9303-b6d394969afd BTC Wallet BTC:0.01520705 USD:27.71

$ cointip transfer --from 104df44b-6238-5c84-9303-b6d394969afd --to 6a06ef0e-306a-5809-a9ae-b61ebf21b4cd --currency USD --amount 1.00
909091aa-7932-51eb-888c-7be1d844efaf completed BTC:-0.00054900 USD:-1.00

$ cointip get-account 6a06ef0e-306a-5809-a9ae-b61ebf21b4cd
6a06ef0e-306a-5809-a9ae-b61ebf21b4cd Test Wallet 2 BTC:0.00054900 USD:1.00
```
