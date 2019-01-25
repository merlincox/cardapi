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

// CardResponse: Card Request
type CardResponse struct {
	Amount      int `json:"amount"`
	CardId      int `json:"cardId"`
	Description string  `json:"description"`
	Id          int `json:"id"`
	RequestType string  `json:"requestType"`
	VendorId    int `json:"vendorId"`
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

type CardRequest struct {

	VendorId int `json:"vendorId"`
	CardId int `json:"cardId"`
	RequestType string `json:"requestType"`
	Amount int `json:"amount"`
	Description string `json:"description"`
}