// Code generated by schema-generator. DO NOT EDIT.

package models

// CalculationResult: Calculation Result
type CalculationResult struct {
	Locale string  `json:"locale"`
	Op     string  `json:"op"`
	Result string  `json:"result"`
	Val1   float64 `json:"val1"`
	Val2   float64 `json:"val2"`
}

// Empty: (No description)
type Empty struct {
}

// Status: API status information
type Status struct {
	Branch    string `json:"branch"`
	Commit    string `json:"commit"`
	Platform  string `json:"platform"`
	Release   string `json:"release"`
	Timestamp string `json:"timestamp"`
}

type Customer struct {

	Fullname string `json:"fullname"`
	Id int `json:"id"`
}

type Vendor struct {

	Fullname string `json:"fullname"`
	Id int `json:"id"`
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

type CardRequest struct {

	VendorId int `json:"vendorId"`
	CardId int `json:"cardId"`
	RequestType string `json:"requestType"`
	Amount int `json:"amount"`
	Description string `json:"description"`
}