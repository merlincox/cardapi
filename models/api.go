// Code generated by schema-generator. DO NOT EDIT.

package models

// AuthMovement: Authorisation movement: capture, refund or reversal
type AuthMovement struct {
	Amount          int    `json:"amount"`
	AuthorisationId int    `json:"authorisationId"`
	Description     string `json:"description"`
	Id              int    `json:"id"`
	MovementType    string `json:"movementType"`
	Ts              string `json:"ts"`
}

// Authorisation: Authorisation
type Authorisation struct {
	Amount      int            `json:"amount"`
	Captured    int            `json:"captured"`
	CardId      int            `json:"cardId"`
	Description string         `json:"description"`
	Id          int            `json:"id"`
	Movements   []AuthMovement `json:"movements,omitempty"`
	Refunded    int            `json:"refunded"`
	Reversed    int            `json:"reversed"`
	Ts          string         `json:"ts"`
	VendorId    int            `json:"vendorId"`
}

// CalculationResult: Calculation Result
type CalculationResult struct {
	Locale string  `json:"locale"`
	Op     string  `json:"op"`
	Result string  `json:"result"`
	Val1   float64 `json:"val1"`
	Val2   float64 `json:"val2"`
}

// Card: Card
type Card struct {
	Available  int        `json:"available"`
	Balance    int        `json:"balance"`
	CustomerId int        `json:"customerId"`
	Id         int        `json:"id"`
	Movements  []Movement `json:"movements,omitempty"`
	Ts         string     `json:"ts"`
}

// CodeRequest: Request for a code
type CodeRequest struct {
	Amount          int    `json:"amount"`
	AuthorisationId int    `json:"authorisationId,omitempty"`
	CardId          int    `json:"cardId,omitempty"`
	Description     string `json:"description,omitempty"`
	VendorId        int    `json:"vendorId,omitempty"`
}

// CodeResponse: Response to a request for a code
type CodeResponse struct {
	Id int `json:"id"`
}

// Customer: Customer
type Customer struct {
	Cards    []Card `json:"cards,omitempty"`
	Fullname string `json:"fullname"`
	Id       int    `json:"id"`
}

// Empty: (No description)
type Empty struct {
}

// Movement: Card movement: top-up, purchase or refund
type Movement struct {
	Amount       int    `json:"amount"`
	CardId       int    `json:"cardId"`
	Description  string `json:"description"`
	Id           int    `json:"id"`
	MovementType string `json:"movementType"`
	Ts           string `json:"ts"`
}

// Status: API status information
type Status struct {
	Branch    string `json:"branch"`
	Commit    string `json:"commit"`
	Platform  string `json:"platform"`
	Release   string `json:"release"`
	Timestamp string `json:"timestamp"`
}

// Vendor: Vendor
type Vendor struct {
	Authorisations []Authorisation `json:"authorisations,omitempty"`
	Balance        int             `json:"balance,omitempty"`
	Id             int             `json:"id"`
	VendorName     string          `json:"vendorName"`
}
