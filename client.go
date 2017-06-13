package cointip

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

const apiEndpoint = "https://api.coinbase.com/v2/"
const apiVersion = "2017-05-17" // https://developers.coinbase.com/api/v2#versioning

const (
	CurrencyUSD = "USD"
	CurrencyBTC = "BTC"
	CurrencyETH = "ETH"
)

type ApiKeyClient struct {
	endpoint  string
	version   string
	apiKey    string
	apiSecret string
	client    http.Client
	debug     bool
}

type Response struct {
	Pagination json.RawMessage `json:"pagination"`
	Data       json.RawMessage `json:"data"`
}

type Price struct {
	Amount   float64 `json:"amount,string"`
	Currency string  `json:"currency"`
}

// APIKeyClient makes a coinbase client using API key auth.
func APIKeyClient(apiKey, apiSecret string) (*ApiKeyClient, error) {

	debug := os.Getenv("COINTIP_DEBUG") == "1"

	timeout := time.Duration(10 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	return &ApiKeyClient{
		endpoint:  apiEndpoint,
		version:   apiVersion,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		client:    client,
		debug:     debug,
	}, nil
}

// https://developers.coinbase.com/docs/wallet/api-key-authentication
func (c *ApiKeyClient) authenticate(req *http.Request, endpoint string, params []byte) {

	timestamp := fmt.Sprintf("%d", time.Now().UTC().Unix())
	message := timestamp + req.Method + req.URL.Path + string(params)

	req.Header.Set("CB-ACCESS-KEY", c.apiKey)

	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(message))

	signature := hex.EncodeToString(h.Sum(nil))

	req.Header.Set("CB-ACCESS-SIGN", signature)
	req.Header.Set("CB-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("CB-VERSION", c.version)
}

// Request makes an authenticated API request.
func (c *ApiKeyClient) Request(method string, path string, params interface{}) (int, []byte, error) {

	endpoint := c.endpoint + path

	jsonParams, err := json.Marshal(params)
	if err != nil {
		return 0, nil, err
	}

	request, err := http.NewRequest(method, endpoint, bytes.NewBuffer(jsonParams))
	if err != nil {
		return 0, nil, err
	}

	c.authenticate(request, endpoint, jsonParams)

	request.Header.Set("User-Agent", "Cointip/v1")
	request.Header.Set("Content-Type", "application/json")

	if c.debug {
		dump, _ := httputil.DumpRequest(request, true)
		fmt.Printf("%s\n\n", dump)
	}

	resp, err := c.client.Do(request)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	if c.debug {
		dump, _ := httputil.DumpResponse(resp, true)
		fmt.Printf("%s%s\n\n", dump, string(body))
	}

	response := &Response{}
	if len(body) > 0 {
		err = json.Unmarshal(body, response)
		if err != nil {
			return 0, nil, err
		}
	}

	// TODO: Pagination

	return resp.StatusCode, response.Data, nil
}

// Price gets the current spot price for a currency
func (c *ApiKeyClient) Price(from string, to string) (*Price, error) {
	code, body, err := c.Request("GET", fmt.Sprintf("prices/%s-%s/spot", from, to), nil)
	if err != nil {
		return nil, err
	}

	if code != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", code)
	}

	price := &Price{}
	err = json.Unmarshal(body, price)
	if err != nil {
		return nil, err
	}

	return price, nil
}