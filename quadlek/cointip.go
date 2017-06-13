package cointip

import (
	"context"
	"sync"

	"strings"

	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/jirwin/quadlek/quadlek"
	"github.com/Bullpeen/cointip"
)

var coinbaseClient *cointip.ApiKeyClient
var bankAccount *cointip.Account
var accountsCache []*cointip.Account
var accountsCacheLock = &sync.Mutex{}

func help(cmdMsg *quadlek.CommandMsg) {
	cmdMsg.Command.Reply() <- &quadlek.CommandResp{
		Text:      "cointip: Tip your friends!\nAvailable commands: help, balance, deposit, withdraw",
		InChannel: false,
	}
}

func sayError(cmdMsg *quadlek.CommandMsg, msg string, inChannel bool) {
	cmdMsg.Command.Reply() <- &quadlek.CommandResp{
		Text:      fmt.Sprintf("Uh Oh. Something broke: %s", msg),
		InChannel: inChannel,
	}
}

func say(cmdMsg *quadlek.CommandMsg, msg string, inChannel bool) {
	cmdMsg.Command.Reply() <- &quadlek.CommandResp{
		Text:      msg,
		InChannel: inChannel,
	}
}

func accountBalanceString(account *cointip.Account) string {
	return fmt.Sprintf(
		"%s:%.2f %s:%.8f",
		account.NativeBalance.Currency, account.NativeBalance.Amount,
		account.Balance.Currency, account.Balance.Amount,
	)
}

func sayPrice(cmdMsg *quadlek.CommandMsg, from string) error {
	price, err := coinbaseClient.Price(from, cointip.CurrencyUSD)
	if err != nil {
		log.WithError(err).Error("Failed getting coinbase price")
		sayError(cmdMsg, err.Error(), false)
		return err
	}

	say(cmdMsg, fmt.Sprintf("%.2f %s", price.Amount, price.Currency), true)
	return nil
}

func getOrCreateAccount(userId string) (*cointip.Account, error) {
	log.Infof("cointip: get or create account %s", userId)
	acctName := fmt.Sprintf("cointip_%s", userId)

	accountsCacheLock.Lock()
	defer accountsCacheLock.Unlock()

	// Warm the cache
	if len(accountsCache) == 0 {
		log.Info("cointip: listing accounts")
		accts, err := coinbaseClient.ListAccounts()
		if err != nil {
			return nil, err
		}
		accountsCache = accts
		log.Infof("cointip: found %d accounts", len(accountsCache))
	}

	for i, account := range accountsCache {
		// If we find an account in the cache, we optionally refresh it and return it
		if account.Name == acctName {
			log.Infof("cointip: refreshing account %s (%s)", account.Name, account.ID)
			account, err := coinbaseClient.GetAccount(account.ID)
			if err != nil {
				return nil, err
			}
			accountsCache[i] = account
			return account, nil
		}
	}

	// Otherwise, create and cache it
	log.Infof("cointip: creating new account %s", acctName)
	account, err := coinbaseClient.CreateAccount(acctName)
	if err != nil {
		return nil, err
	}
	accountsCache = append(accountsCache, account)
	log.Infof("cointip: created new cointip account: %s (%s)", account.Name, account.ID)

	if bankAccount == nil {
		log.Infof("cointip: skipping account priming - bank account does not exist")
		return account, nil
	}
	tx, err := coinbaseClient.Transfer(bankAccount.ID, account.ID, &cointip.Balance{Currency: cointip.CurrencyUSD, Amount: 3.00})
	if err != nil {
		log.WithError(err).Errorf("cointip: failed to prime new cointip account from bank: %s (%s)", bankAccount.Name, bankAccount.ID)
		return account, nil
	}

	log.Infof("cointip: primed new cointip account - refreshing again %s (%s) txid: %s", account.Name, account.ID, tx.ID)
	refreshed, err := coinbaseClient.GetAccount(account.ID)
	if err != nil {
		log.WithError(err).Errorf("cointip: failed refreshing new account after priming, returning non-refreshed account: %s", err)
		return account, err
	}
	return refreshed, nil
}

func cointipReaction(ctx context.Context, reactionChannel <-chan *quadlek.ReactionHookMsg) {
	for {
		select {
		case rh := <-reactionChannel:

			amount := &cointip.Balance{
				Currency: cointip.CurrencyUSD,
			}
			switch rh.Reaction.Reaction {
			case "cointip_1":
				amount.Amount = .01
			case "cointip_2":
				amount.Amount = .02
			case "cointip_5":
				amount.Amount = .05
			case "cointip_10":
				amount.Amount = .10
			case "cointip_25":
				amount.Amount = .25
			default:
				continue
			}

			if bankAccount == nil {
				log.Info("cointip: ignoring tip reaction - plugin is not initialized")
				continue
			}

			log.Infof("cointip: got reaction %s from:%s to:%s", rh.Reaction.Reaction, rh.Reaction.User, rh.Reaction.ItemUser)
			if rh.Reaction.User == rh.Reaction.ItemUser {
				log.Infof("cointip: skipping tip - user is tipping themselves")
				continue
			}

			from, err := getOrCreateAccount(rh.Reaction.User)
			if err != nil {
				log.WithError(err).Error("cointip: tip failed - failed fetching coinbase account.")
				continue
			}
			to, err := getOrCreateAccount(rh.Reaction.ItemUser)
			if err != nil {
				log.WithError(err).Error("cointip: tip failed - failed fetching coinbase account.")
				continue
			}

			tx, err := coinbaseClient.Transfer(from.ID, to.ID, amount)
			if err != nil {
				log.WithError(err).Error("cointip: tip failed - ailed creating transaction.")
				continue
			}

			log.Infof("%s (%s) tipped %s (%s) %s:%.2f txid: %s", from.Name, from.ID, to.Name, to.ID, tx.NativeAmount.Currency, tx.NativeAmount.Amount, tx.ID)

		case <-ctx.Done():
			log.Info("cointip: stopping reaction hook")
			return
		}
	}
}

func cointipCommand(ctx context.Context, cmdChannel <-chan *quadlek.CommandMsg) {
	for {
		select {
		case cmdMsg := <-cmdChannel:

			// /cointip <command> <args...>
			cmd := strings.SplitN(cmdMsg.Command.Text, " ", 1)
			if len(cmd) == 0 {
				help(cmdMsg)
				return
			}
			log.Infof("cointip: got command %s", cmd[0])
			switch cmd[0] {
			case "balance":
				account, err := getOrCreateAccount(cmdMsg.Command.UserId)
				if err != nil {
					log.WithError(err).Error("Failed fetching coinbase account.")
					sayError(cmdMsg, err.Error(), false)
					continue
				}
				say(cmdMsg, fmt.Sprintf("tipjar balance: %s", accountBalanceString(account)), false)
			case "deposit":
				account, err := getOrCreateAccount(cmdMsg.Command.UserId)
				if err != nil {
					log.WithError(err).Error("Failed fetching coinbase account.")
					sayError(cmdMsg, err.Error(), false)
					continue
				}
				address, err := coinbaseClient.CreateAddress(account.ID)
				if err != nil {
					log.WithError(err).Error("Failed fetching coinbase address.")
					sayError(cmdMsg, err.Error(), false)
					continue
				}
				//TODO QR code or something for address, also allow coinbase <-> coinbase transfers to dodge fees
				say(cmdMsg, fmt.Sprintf("deposit address: %s", address.Address), false)
			case "withdraw":
				say(cmdMsg, "withdraw is not implemented yet, sorry!", false)
			default:
				help(cmdMsg)
			}

		case <-ctx.Done():
			log.Info("cointip: stopping plugin")
			return
		}
	}
}

func btcCommand(ctx context.Context, cmdChannel <-chan *quadlek.CommandMsg) {
	for {
		select {
		case cmdMsg := <-cmdChannel:
			err := sayPrice(cmdMsg, cointip.CurrencyBTC)

			if err != nil {
				continue
			}
		case <-ctx.Done():
			log.Info("cointip: stopping btc plugin")
			return
		}
	}
}

func ethCommand(ctx context.Context, cmdChannel <-chan *quadlek.CommandMsg) {
	for {
		select {
		case cmdMsg := <-cmdChannel:
			err := sayPrice(cmdMsg, cointip.CurrencyETH)

			if err != nil {
				continue
			}
		case <-ctx.Done():
			log.Info("cointip: stopping eth plugin")
			return
		}
	}
}

func Register(apiKey, apiSecret, bankAccountId string) quadlek.Plugin {
	client, err := cointip.APIKeyClient(apiKey, apiSecret)
	if err != nil {
		log.WithError(err).Errorf("cointip: failed to create coinbase client, bailing: %s", err)
		return nil
	}
	coinbaseClient = client

	// Warm the cache and fetch the bank account
	account, err := getOrCreateAccount(bankAccountId)
	if err != nil {
		log.WithError(err).Errorf("cointip: failed to set up bank account, bailing: %s", err)
		return nil
	}
	bankAccount = account

	log.Infof("cointip: starting plugin bank:%s (%s) %s total_accounts:%d", bankAccount.Name, bankAccount.ID, accountBalanceString(bankAccount), len(accountsCache))

	return quadlek.MakePlugin(
		"cointip",
		[]quadlek.Command{
			quadlek.MakeCommand("cointip", cointipCommand),
			quadlek.MakeCommand("btc", btcCommand),
			quadlek.MakeCommand("eth", ethCommand),
		},
		nil,
		[]quadlek.ReactionHook{
			quadlek.MakeReactionHook(cointipReaction),
		},
		nil,
		nil,
	)
}
