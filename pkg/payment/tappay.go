package payment

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"

	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/utils"
)

// Pay could send payment body for different url, e.g. ByPrime or ByToken
func Pay(url string, payload []byte) (resp []byte, err error) {

	_, body, err := utils.HTTPRequest("POST", url,
		map[string]string{
			"x-api-key": config.Config.PaymentService.PartnerKey,
		}, payload)

	if err != nil {
		log.Printf("Charge error:%v\n", err)
		return resp, err
	}

	return body, err
}

// PrimePayload holds informations for requesting Tappay Pay By Prime API
type PrimePayload struct {
	PartnerKey string `json:"partner_key,omitempty" db:"partner_key"`
	MerchantID string `json:"merchant_id,omitempty" db:"merchant_id"`
	Prime      string `json:"prime,omitempty" db:"prime"`
	Currency   string `json:"currency,omitempty" db:"currency"`
	Details    string `json:"details,omitempty" db:"detail"`
	Remember   bool   `json:"remember,omitempty"`
	Amount     int    `json:"amount,omitempty"`
	Cardholder struct {
		PhoneNumber string `json:"phone_number,omitempty" db:"phone_number"`
		Name        string `json:"name,omitempty" db:"name"`
		Email       string `json:"email,omitempty" db:"email"`
	} `json:"cardholder,omitempty" db:"cardholder"`
}

// PrimeResp is the response from Prime API
type PrimeResp struct {
	Status      int    `json:"status"`
	Message     string `json:"msg"`
	RecTradeID  string `json:"rec_trade_id"`
	BankCode    string `json:"bank_result_code"`
	BankMessage string `json:"bank_result_msg"`
	CardSecret  struct {
		CardToken string `json:"card_token"`
		CardKey   string `json:"card_key"`
	} `json:"card_secret"`
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
	Details  string `json:"details"`
}

type PayByPrime struct {
	Payload PrimePayload
	Resp    PrimeResp
}

// Pay passes the payload with prime url to Pay()
func (p *PayByPrime) Pay() (err error) {

	p.Payload.PartnerKey = config.Config.PaymentService.PartnerKey
	p.Payload.MerchantID = config.Config.PaymentService.MerchantID

	payload, _ := json.Marshal(p)
	r, err := Pay(config.Config.PaymentService.PrimeURL, payload)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(r, &p.Resp); err != nil {
		return err
	}

	// token = make(map[string]interface{})
	// // Parse for pay by token
	// if p.Remember {
	// 	token["card_key"] = prime.CardSecret.CardKey
	// 	token["card_token"] = prime.CardSecret.CardToken
	// 	token["currency"] = prime.Currency
	// 	token["details"] = prime.Details
	// }
	return nil
}

func (p *PayByPrime) Token() (Provider, error) {
	var result = PayByCardToken{}
	fmt.Printf("PayByPrime RESP:%v\n", p.Resp)
	result.Payload.CardKey = p.Resp.CardSecret.CardKey
	result.Payload.CardToken = p.Resp.CardSecret.CardToken
	result.Payload.Currency = p.Resp.Currency
	result.Payload.Details = p.Resp.Details
	result.Payload.Amount = p.Resp.Amount

	return &result, nil
}

// Value converts the value in PayByPrime to JSON in []byte and could be stored to db
func (p *PayByPrime) Value() (driver.Value, error) {
	return json.Marshal(p.Payload)
}

// Scan the data from database and stores in PayByPrime
func (p *PayByPrime) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &p.Payload)
}

type TokenPayload struct {
	CardKey    string `json:"card_key,omitempty"`
	CardToken  string `json:"card_token,omitempty"`
	PartnerKey string `json:"partner_key,omitempty"`
	MerchantID string `json:"merchant_id,omitempty"`
	Amount     int    `json:"amount,omitempty"`
	Currency   string `json:"currency,omitempty"`
	Details    string `json:"details,omitempty"`
}

type TokenResp struct {
	Status      int    `json:"status"`
	Message     string `json:"msg"`
	RecTradeID  string `json:"rec_trade_id"`
	BankCode    string `json:"bank_result_code"`
	BankMessage string `json:"bank_result_msg"`
	Amount      int    `json:"amount"`
	Currency    string `json:"currency"`
	Details     string `json:"details"`
}

// PayByCardToken holds informations for requesting Tappay Pay By Card Token API
type PayByCardToken struct {
	Payload TokenPayload
	Resp    TokenResp
}

func (p *PayByCardToken) Pay() (err error) {

	p.Payload.PartnerKey = config.Config.PaymentService.PartnerKey
	p.Payload.MerchantID = config.Config.PaymentService.MerchantID

	reqBody, _ := json.Marshal(p)
	r, err := Pay(config.Config.PaymentService.TokenURL, reqBody)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(r, &p.Resp); err != nil {
		return err
	}
	return nil
}

func (p *PayByCardToken) Token() (Provider, error) {
	return p, nil
}

// Value converts the value in PayByCardToken to JSON in []byte and could be stored to db
func (p *PayByCardToken) Value() (driver.Value, error) {
	return json.Marshal(p.Payload)
}

// Scan the data from database and stores in PayByCardToken	
func (p *PayByCardToken) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &p.Payload)
}
