package models

import (
	"fmt"
	"net/http"
)

type ApiError interface {
	Error() string
	StatusCode() int
	ErrorBody() ApiErrorBody
}

type ApiErrorBody struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type errBody struct {
	body ApiErrorBody
}

func (err errBody) Error() string {
	return err.body.Message
}

func (err errBody) StatusCode() int {
	return err.body.Code
}

func (err errBody) ErrorBody() ApiErrorBody {
	return err.body
}

func ConstructApiError(code int, format string, a ...interface{}) ApiError {

	return errBody{
		body: ApiErrorBody{
			Message: fmt.Sprintf(format, a...),
			Code:    code,
		},
	}
}

func ErrorWrap(err error) ApiError {

	apiErr, ok := err.(ApiError)

	if ok {
		return apiErr
	}

	return errBody{
		body: ApiErrorBody{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		},
	}
}

type Customer struct {

	Fullname string `json:"fullname"`
	Id int `json:"id"`
}

type Vendor struct {

	VendorName string `json:"vendorName"`
	Id         int    `json:"id"`
	Balance    int    `json:balance`
}

type Card struct {

	Id int `json:"id"`
	Available int `json:"available"`
	Balance int `json:"balance"`
	CustomerId int `json:"customeriId"`
	Ts string `json:"ts"`
}

type Authorisation struct {

	Id int `json:"id"`
	CardId int `json:"cardId"`
	VendorId int `json:"vendorId"`
	Amount int `json:"amount"`
	Captured int `json:"captured"`
	Refunded int `json:"captured"`
	Reversed int `json:"captured"`
	Description string `json:"description"`
	Ts string `json:"ts"`
}

type Movement struct {

	Id int `json:"id"`
	CardId int `json:"cardId"`
	MovementType string `json:"movementType"`
	Amount int `json:"amount"`
	Description string `json:"description"`
	Ts string `json:"ts"`
}

// Amount that can be captured or reversed
func (auth Authorisation) Capturable() int {
	return auth.Amount - (auth.Captured + auth.Reversed)
}

// Amount that can be refunded
func (auth Authorisation) Refundable() int {
	return auth.Captured - auth.Refunded
}
