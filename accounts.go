package cointip

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	CurrencyUSD = "USD"
	CurrencyBTC = "BTC"
)

type Balance struct {
	Amount   float64 `json:"amount,string"`
	Currency string  `json:"currency"`
}

type Account struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Currency      string  `json:"currency"`
	Balance       Balance `json:"balance"`        // Amount in Cryptocurrency
	NativeBalance Balance `json:"native_balance"` // Amount in USD
}

type Transaction struct {
	ID           string  `json:"id"`
	Type         string  `json:"type"`
	Status       string  `json:"status"`
	Amount       Balance `json:"amount"`        // Amount in Cryptocurrency
	NativeAmount Balance `json:"native_amount"` // Amount in USD
	Description  string  `json:"description"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type Address struct {
	ID        string `json:"id"`
	Address   string `json:"address"`
	Name      string `json:"name"`
	Network   string `json:"network"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (c *ApiKeyClient) ListAccounts() ([]*Account, error) {
	code, body, err := c.Request("GET", "accounts", nil)
	if err != nil {
		return nil, err
	}

	if code != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", code)
	}

	accounts := []*Account{}
	err = json.Unmarshal(body, &accounts)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func (c *ApiKeyClient) GetAccount(id string) (*Account, error) {

	code, body, err := c.Request("GET", fmt.Sprintf("accounts/%s", id), nil)
	if err != nil {
		return nil, err
	}

	if code != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", code)
	}

	account := &Account{}
	err = json.Unmarshal(body, account)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (c *ApiKeyClient) CreateAccount(Name string) (*Account, error) {

	code, body, err := c.Request("POST", "accounts", map[string]string{"name": Name})
	if err != nil {
		return nil, err
	}

	if code != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code %d", code)
	}

	account := &Account{}
	err = json.Unmarshal(body, account)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (c *ApiKeyClient) DeleteAccount(id string) error {

	code, _, err := c.Request("DELETE", fmt.Sprintf("accounts/%s", id), nil)
	if err != nil {
		return err
	}

	if code != http.StatusNoContent {
		return fmt.Errorf("unexpected status code %d", code)
	}

	return nil
}

// CreateAddress creates an address for the given account id, letting users deposit funds.
func (c *ApiKeyClient) CreateAddress(id string) (*Address, error) {

	code, body, err := c.Request("POST", fmt.Sprintf("accounts/%s/addresses", id), nil)
	if err != nil {
		return nil, err
	}

	if code != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code %d", code)
	}

	addr := &Address{}
	err = json.Unmarshal(body, addr)
	if err != nil {
		return nil, err
	}

	return addr, nil
}

// Transfer moves funds between accounts. Use this when tipping users.
func (c *ApiKeyClient) Transfer(from, to string, amount *Balance) (*Transaction, error) {

	if !(amount.Currency == CurrencyBTC || amount.Currency == CurrencyUSD) {
		return nil, fmt.Errorf("invalid currency type: %s", amount.Currency)
	}

	params := map[string]string{
		"type":        "transfer",
		"to":          to,
		"amount":      fmt.Sprintf("%.8f", amount.Amount),
		"currency":    amount.Currency,
		"description": "cointip transfer",
	}
	code, body, err := c.Request("POST", fmt.Sprintf("accounts/%s/transactions", from), params)
	if err != nil {
		return nil, err
	}

	if code != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code %d", code)
	}

	tx := &Transaction{}
	err = json.Unmarshal(body, tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// Withdraw sends funds from an account id to an external address, letting users pull funds from their tipjar.
func (c *ApiKeyClient) Withdraw(from, to string, amount *Balance) (*Transaction, error) {

	if !(amount.Currency == CurrencyBTC || amount.Currency == CurrencyUSD) {
		return nil, fmt.Errorf("invalid currency type: %s", amount.Currency)
	}

	params := map[string]string{
		"type":        "send",
		"to":          to,
		"amount":      fmt.Sprintf("%.8f", amount.Amount),
		"currency":    amount.Currency,
		"description": "cointip withdraw",
	}
	code, body, err := c.Request("POST", fmt.Sprintf("accounts/%s/transactions", from), params)
	if err != nil {
		return nil, err
	}

	if code != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code %d", code)
	}

	tx := &Transaction{}
	err = json.Unmarshal(body, tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// GetTransaction returns a given transaction by id.
func (c *ApiKeyClient) GetTransaction(id, txID string) (*Transaction, error) {

	code, body, err := c.Request("GET", fmt.Sprintf("accounts/%s/transactions/%s", id, txID), nil)
	if err != nil {
		return nil, err
	}

	if code != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", code)
	}

	tx := &Transaction{}
	err = json.Unmarshal(body, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
