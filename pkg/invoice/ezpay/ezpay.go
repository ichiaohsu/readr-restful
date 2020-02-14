package ezpay

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/readr-media/readr-restful/config"
)

const DefaultCommentLength = 71

// InvoiceClient holds the infos to create invoice with ezPay
type InvoiceClient struct {
	Payload map[string]interface{}
}

// PKCS7Padding will add paddings to input bytearray
func PKCS7Padding(b []byte, blocksize int) ([]byte, error) {

	if blocksize <= 0 {
		return nil, errors.New("invalid blocksize")
	}
	if b == nil || len(b) == 0 {
		return nil, errors.New("invalid PKCS7 data (empty or not padded)")
	}
	n := blocksize - (len(b) % blocksize)
	pb := make([]byte, len(b)+n)
	copy(pb, b)
	copy(pb[len(b):], bytes.Repeat([]byte{byte(n)}, n))
	return pb, nil
}

// Create makes an invoice API call to ezpay
func (c *InvoiceClient) Create() (resp map[string]interface{}, err error) {

	dataURL := url.Values{}
	for k, v := range c.Payload {
		dataURL.Set(k, fmt.Sprintf("%v", v))
	}
	postdata := []byte(dataURL.Encode())
	key := []byte(config.Config.InvoiceService.Key)
	iv := []byte(config.Config.InvoiceService.IV)

	// encrypt PostData_ first with AES-CBC
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error creating new cipher for %s when create invoice:%v", key, err.Error())
	}
	// Add PKCS7Padding
	data, err := PKCS7Padding(postdata, aes.BlockSize)
	if err != nil {
		return nil, err
	}
	encrypter := cipher.NewCBCEncrypter(block, iv)
	encrypter.CryptBlocks(data, data)

	postURL := url.Values{}
	postURL.Set("MerchantID_", config.Config.InvoiceService.MerchantID)
	postURL.Set("PostData_", hex.EncodeToString(data))

	req, err := http.NewRequest("POST", config.Config.InvoiceService.URL, bytes.NewBufferString(postURL.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating http request for ezPay error:%s", err.Error())
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting ezPay API error: %s", err.Error())
	}
	defer r.Body.Close()
	// Parse response
	respBody, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return nil, fmt.Errorf("parsing response from ezPay error:%s", err.Error())
	}

	if status, ok := resp["Status"]; ok && status != "SUCCESS" {
		return nil, fmt.Errorf("create invoice error:%s", resp["Message"])
	}
	return resp, nil
}

func get(target map[string]interface{}, key string, defaultValue interface{}) (result interface{}) {
	if value, ok := target[key]; ok {
		switch value.(type) {
		case string:
			if value.(string) != "" {
				return value
			}
		}
		return value
	}
	return defaultValue
}

// iinterfaceSliceToa converts a []interface{} containing strings to a pure []string
func interfaceSliceToa(i []interface{}) (result []string) {
	for _, v := range i {
		result = append(result, v.(string))
	}
	return result
}

// interfaceSliceItoa converts each integer in []interface{} to a and put them into a []string
func interfaceSliceItoa(i []interface{}) (result []string) {
	for _, v := range i {
		result = append(result, strconv.Itoa(int(v.(float64))))
	}
	return result
}

// Validate check the data for InvoiceClient, fix missing fields
func (c *InvoiceClient) Validate() (err error) {

	var result = make(map[string]interface{}, 0)

	result["RespondType"] = get(c.Payload, "response_type", "JSON")
	result["TimeStamp"] = time.Now().Unix()
	result["MerchantOrderNo"] = time.Now().Format("20060102")
	result["Status"] = get(c.Payload, "status", "0")
	result["TaxType"] = get(c.Payload, "tax_type", "1")
	result["Category"] = get(c.Payload, "category", "B2C")
	result["LoveCode"] = get(c.Payload, "love_code", "")
	result["CarrierType"] = get(c.Payload, "carrier_type", "")
	result["CarrierNum"] = get(c.Payload, "carrier_num", "")
	result["BuyerName"] = get(c.Payload, "buyer_name", "")
	result["BuyerEmail"] = get(c.Payload, "buyer_email", "")

	result["ItemName"] = get(c.Payload, "item_name", []interface{}{})
	result["ItemCount"] = get(c.Payload, "item_count", []interface{}{})
	result["ItemPrice"] = get(c.Payload, "item_price", []interface{}{})
	result["ItemUnit"] = get(c.Payload, "item_unit", []interface{}{})

	result["TotalAmt"] = get(c.Payload, "amount", nil)
	if result["TotalAmt"] == nil {
		return errors.New("invalid amount")
	}

	if config.Config.InvoiceService.APIVersion == "" {
		result["Version"] = "1.4"
	} else {
		result["Version"] = config.Config.InvoiceService.APIVersion
	}

	switch result["TaxType"].(string) {
	case "2":
		result["TaxRate"] = 0
		result["TaxAmt"] = 0
		result["Amt"] = result["TotalAmt"]
		result["CustomsClearance"] = "1"
	case "3":
		result["TaxRate"] = 0
		result["TaxAmt"] = 0
		result["Amt"] = result["TotalAmt"]
	case "9":
		// TODO: validation
		// Temporarily fallthrough
		fallthrough
	case "1":
		fallthrough
	default:
		result["TaxRate"] = 5
		result["TaxAmt"] = math.Round(float64(result["TotalAmt"].(int)) * (float64(result["TaxRate"].(int)) / 100))
		result["Amt"] = float64(result["TotalAmt"].(int)) - result["TaxAmt"].(float64)
	}

	if result["Category"] == "B2B" {

		result["PrintFlag"] = "Y"
		result["BuyerUBN"] = get(c.Payload, "buyer_ubn", "-")
		result["BuyerAddress"] = get(c.Payload, "buyer_address", "-")

		var taxfreePrice = []int{}
		for _, v := range result["ItemPrice"].([]interface{}) {
			price := int(math.Round(v.(float64) / float64(1+result["TaxRate"].(int)/100)))
			taxfreePrice = append(taxfreePrice, price)
		}
		result["ItemPrice"] = taxfreePrice
		delete(result, "CarrierType")

	} else if result["Category"] == "B2C" {

		if result["LoveCode"] != "" {
			// check if LoveCode is a 3~7 digits int string
			if match, _ := regexp.MatchString("^[0-9]{3,7}$", result["LoveCode"].(string)); !match {
				delete(result, "CarrierType")
				result["PrintFlag"] = "Y"
			} else {
				result["CarrierType"] = ""
			}
		} else {

			switch result["CarrierType"] {
			case "0":
				fallthrough
			case "1":
				var checkString string
				if result["CarrierType"] == "0" {
					checkString = "^/[A-Z0-9+-.]{7}$"
				} else if result["CarrierType"] == "1" {
					checkString = "^/[A-Z0-9+-.]{7}$"
				}
				if match, _ := regexp.MatchString(checkString, result["CarrierNum"].(string)); !match {
					delete(result, "CarrierType")
					result["PrintFlag"] = "Y"
					result["Comment"] = fmt.Sprintf("Incorrect carrier num: %s", result["CarrierNum"])
				} else {
					result["CarrierNum"] = strings.TrimSpace(result["CarrierNum"].(string))
				}
			case "2":
				if buyerEmail, ok := result["BuyerEmail"]; buyerEmail == "" || !ok {
					return errors.New("empty buyer_email when carrier_type = 2")
				}
				result["CarrierNum"] = result["BuyerEmail"]
			default:
				delete(result, "CarrierType")
				result["PrintFlag"] = "Y"
			}
		}
	}
	if len(result["ItemCount"].([]interface{})) == len(result["ItemPrice"].([]interface{})) {
		var itemAmt = []int{}
		for i := range result["ItemCount"].([]interface{}) {
			count := int(result["ItemCount"].([]interface{})[i].(float64))
			price := int(result["ItemPrice"].([]interface{})[i].(float64))
			itemAmt = append(itemAmt, count*price)
		}
		result["ItemAmt"] = itemAmt
	}

	result["ItemName"] = strings.Join(interfaceSliceToa(result["ItemName"].([]interface{})), "|")
	result["ItemUnit"] = strings.Join(interfaceSliceToa(result["ItemUnit"].([]interface{})), "|")
	result["ItemPrice"] = strings.Join(interfaceSliceItoa(result["ItemPrice"].([]interface{})), "|")
	result["ItemCount"] = strings.Join(interfaceSliceItoa(result["ItemCount"].([]interface{})), "|")
	result["ItemAmt"] = strings.Join(func(i []int) (r []string) {
		for _, v := range i {
			r = append(r, strconv.Itoa(v))
		}
		return r
	}(result["ItemAmt"].([]int)), "|")

	// Trim comment messeage to allowed length
	if comment, ok := result["Comment"].(string); ok && len(comment) > DefaultCommentLength {
		result["Comment"] = func(s string, l int) string {
			result := []rune(s)
			if len(result) > l {
				result = result[:l]
			}
			return string(result)
		}(comment, DefaultCommentLength)
	}
	c.Payload = result
	return nil
}

func sliceStrToInt(target []string) (result []int, err error) {
	for _, s := range target {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return result, err
		}
		result = append(result, int(i))
	}
	return result, err
}

// Invoice holds the needed information for invoice in EZPay
type Invoice struct {
	RespondType      string `json:"RespondType,omitempty"`
	Version          string `json:"Version,omitempty"`
	TimeStamp        string `json:"TimeStamp,omitempty"`
	MerchantOrderNo  string `json:"MerchantOrderNo,omitempty"`
	Status           string `json:"Status,omitempty"`
	Category         string `json:"Category,omitempty"`
	BuyerName        string `json:"BuyerName,omitempty"`
	BuyerUBN         string `json:"BuyerUBN,omityempty"`
	BuyerAddress     string `json:"BuyerAddress,omitempty"`
	BuyerEmail       string `json:"BuyerEmail,omitempty"`
	CarrierType      string `json:"CarrierType,omitempty"`
	CarrierNum       string `json:"CarrierNum,omitempty"`
	LoveCode         string `json:"LoveCode,omitempty"`
	PrintFlag        string `json:"PrintFlag,omitempty"`
	TaxType          string `json:"TaxType,omitempty"`
	TaxRate          int    `json:"TaxRate,omitempty"`
	CustomsClearance string `json:"CustomsClearance,omitempty"`
	Amt              int    `json:"Amt,omitempty"`
	TaxAmt           int    `json:"TaxAmt,omitempty"`
	TotalAmt         int    `json:"TotalAmt,omitempty"`
	ItemName         string `json:"ItemName,omitempty"`
	ItemCount        string `json:"ItemCount,omitempty"`
	ItemUnit         string `json:"ItemUnit,omitempty"`
	ItemPrice        string `json:"ItemPrice,omitempty"`
	ItemAmt          string `json:"ItemAmt,omitempty"`
	Comment          string `json:"Comment,omitempty"`
}

// Create makes an invoice API call to ezpay
func (c *Invoice) Create() (resp map[string]interface{}, err error) {

	dataURL := url.Values{}

	var payload map[string]interface{}
	inter, _ := json.Marshal(c)
	json.Unmarshal(inter, &payload)

	for k, v := range payload {
		dataURL.Set(k, fmt.Sprintf("%v", v))
	}

	postdata := []byte(dataURL.Encode())
	key := []byte(config.Config.InvoiceService.Key)
	iv := []byte(config.Config.InvoiceService.IV)

	// encrypt PostData_ first with AES-CBC
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error creating new cipher for %s when create invoice:%v", key, err.Error())
	}
	// Add PKCS7Padding
	data, err := PKCS7Padding(postdata, aes.BlockSize)
	if err != nil {
		return nil, err
	}
	encrypter := cipher.NewCBCEncrypter(block, iv)
	encrypter.CryptBlocks(data, data)

	postURL := url.Values{}
	postURL.Set("MerchantID_", config.Config.InvoiceService.MerchantID)
	postURL.Set("PostData_", hex.EncodeToString(data))

	req, err := http.NewRequest("POST", config.Config.InvoiceService.URL, bytes.NewBufferString(postURL.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating http request for ezPay error:%s", err.Error())
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting ezPay API error: %s", err.Error())
	}
	defer r.Body.Close()
	// Parse response
	respBody, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return nil, fmt.Errorf("parsing response from ezPay error:%s", err.Error())
	}

	if status, ok := resp["Status"]; ok && status != "SUCCESS" {
		return nil, fmt.Errorf("create invoice error:%s", resp["Message"])
	}
	return resp, nil
}

func (c *Invoice) Validate() (err error) {

	if c.RespondType == "" {
		c.RespondType = "JSON"
	}
	c.TimeStamp = strconv.FormatInt(time.Now().Unix(), 10)
	if c.MerchantOrderNo == "" {
		c.MerchantOrderNo = time.Now().Format("20060102")
	}
	if c.Status == "" {
		c.Status = "0"
	}
	if c.TaxType == "" {
		c.TaxType = "1"
	}
	if c.Category == "" {
		c.Category = "B2C"
	}

	if c.TotalAmt == 0 {
		return errors.New("invalid amount")
	}
	itemPrice, err := sliceStrToInt(strings.Split(c.ItemPrice, "|"))
	if err != nil {
		return err
	}

	itemCount, err := sliceStrToInt(strings.Split(c.ItemCount, "|"))
	if err != nil {
		return err
	}
	var itemAmt []int

	if config.Config.InvoiceService.APIVersion == "" {
		c.Version = "1.4"
	} else {
		c.Version = config.Config.InvoiceService.APIVersion
	}
	switch c.TaxType {
	case "2":
		c.TaxRate = 0
		c.TaxAmt = 0
		c.Amt = c.TotalAmt
		c.CustomsClearance = "1"
	case "3":
		c.TaxRate = 0
		c.TaxAmt = 0
		c.Amt = c.TotalAmt
	case "9":
		// TODO: validation
		// Temporarily fallthrough
		fallthrough
	case "1":
		fallthrough
	default:
		c.TaxRate = 5
		c.TaxAmt = int(math.Round(float64(c.TotalAmt) * (float64(c.TaxRate) / 100)))
		c.Amt = c.TotalAmt - c.TaxAmt
	}

	if c.Category == "B2B" {

		c.PrintFlag = "Y"
		if c.BuyerUBN == "" {
			c.BuyerUBN = "-"
		}
		if c.BuyerAddress == "" {
			c.BuyerAddress = "-"
		}

		var taxfreePrice = []int{}
		for _, p := range itemPrice {
			price := int(math.Round(float64(p) / float64(1+c.TaxRate/100)))
			taxfreePrice = append(taxfreePrice, price)
		}
		itemPrice = taxfreePrice
		c.CarrierType = ""

	} else if c.Category == "B2C" {

		if c.LoveCode != "" {
			// check if LoveCode is a 3~7 digits int string
			if match, _ := regexp.MatchString("^[0-9]{3,7}$", c.LoveCode); !match {
				c.CarrierType = ""
				c.PrintFlag = "Y"
			} else {
				c.CarrierType = ""
			}
		} else {

			switch c.CarrierType {
			case "0":
				fallthrough
			case "1":
				var checkString string
				if c.CarrierType == "0" {
					checkString = "^/[A-Z0-9+-.]{7}$"
				} else if c.CarrierType == "1" {
					checkString = "^/[A-Z0-9+-.]{7}$"
				}
				if match, _ := regexp.MatchString(checkString, c.CarrierNum); !match {
					c.CarrierType = ""
					c.PrintFlag = "Y"
					c.Comment = fmt.Sprintf("Incorrect carrier num: %s", c.CarrierNum)
				} else {
					c.CarrierNum = strings.TrimSpace(c.CarrierNum)
				}
			case "2":
				if c.BuyerEmail == "" {
					return errors.New("empty buyer_email when carrier_type = 2")
				}
				c.CarrierNum = c.BuyerEmail
			default:
				c.CarrierType = ""
				c.PrintFlag = "Y"
			}
		}
	}

	if len(itemCount) == len(itemPrice) {
		// var itemAmt = []int{}
		for i := range itemCount {
			count := itemCount[i]
			price := itemPrice[i]
			itemAmt = append(itemAmt, count*price)
		}
	}

	c.ItemAmt = strings.Join(func(i []int) (r []string) {
		for _, v := range i {
			r = append(r, strconv.Itoa(v))
		}
		return r
	}(itemAmt), "|")

	// Trim comment messeage to allowed length
	if c.Comment != "" && len(c.Comment) > DefaultCommentLength {
		c.Comment = func(s string, l int) string {
			result := []rune(s)
			if len(result) > l {
				result = result[:l]
			}
			return string(result)
		}(c.Comment, DefaultCommentLength)
	}
	return nil
}

// Value turn the data in Invoice byte array for database
func (c *Invoice) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan the data from database and stores in Invoice
func (c *Invoice) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), c)
}
