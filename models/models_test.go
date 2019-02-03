package models

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/merlincox/cardapi/utils"
)

func TestApiErrorBody(t *testing.T) {

	eb := ApiErrorBody{
		Message: "Testing testing 1 2 3",
		Code:    123,
	}

	tb := errBody{
		body: eb,
	}

	utils.AssertEquals(t, "ApiErrorBody error", "Testing testing 1 2 3", tb.Error())
	utils.AssertEquals(t, "ApiErrorBody code", 123, tb.StatusCode())
	utils.AssertEquals(t, "ApiErrorBody body", eb, tb.ErrorBody())
}

func TestErrorf(t *testing.T) {

	var err ApiError

	err = ConstructApiError(123, "Testing testing %v %v %v", 1, "2", 3)

	eb := ApiErrorBody{
		Message: "Testing testing 1 2 3",
		Code:    123,
	}

	utils.AssertEquals(t, "ConstructApiError string", "Testing testing 1 2 3", err.Error())
	utils.AssertEquals(t, "ConstructApiError code", 123, err.StatusCode())
	utils.AssertEquals(t, "ConstructApiError body", eb, err.ErrorBody())
}

func TestErrorWrap(t *testing.T) {

	innerErr := errors.New("I am an error")

	innerErr2 := ConstructApiError(123, "I am an %v", "API error")

	err := ErrorWrap(innerErr)

	err2 := ErrorWrap(innerErr2)

	errBody := ApiErrorBody{
		Message: "I am an error",
		Code:    500,
	}

	errBody2 := ApiErrorBody{
		Message: "I am an API error",
		Code:    123,
	}

	utils.AssertEquals(t, "Non API error string", "I am an error", err.Error())
	utils.AssertEquals(t, "Non API error code", 500, err.StatusCode())
	utils.AssertEquals(t, "Non API error body", errBody, err.ErrorBody())
	utils.AssertEquals(t, "API error string", "I am an API error", err2.Error())
	utils.AssertEquals(t, "API error code", 123, err2.StatusCode())
	utils.AssertEquals(t, "API error body", errBody2, err2.ErrorBody())
}

func TestAuthorisation_Capturable(t *testing.T) {

	a := Authorisation{
		Amount:   100,
		Captured: 0,
		Reversed: 0,
		Refunded: 0,
	}

	utils.AssertEquals(t, "Capturable", 100, a.Capturable())

	a.Captured = 20

	utils.AssertEquals(t, "Capturable", 80, a.Capturable())

	a.Reversed = 25

	utils.AssertEquals(t, "Capturable", 55, a.Capturable())
}

func TestAuthorisation_Refundable(t *testing.T) {

	a := Authorisation{
		Amount:   100,
		Captured: 0,
		Reversed: 0,
		Refunded: 0,
	}

	utils.AssertEquals(t, "Refundable", 0, a.Refundable())

	a.Captured = 50

	utils.AssertEquals(t, "Refundable", 50, a.Refundable())

	a.Refunded = 25

	utils.AssertEquals(t, "Refundable", 25, a.Refundable())
}

func TestNullableAuthorisation_Valid(t *testing.T) {

	na := NullableAuthorisation{}

	utils.AssertEquals(t, "NullableAuthorisation Valid", false, na.Valid())

	na.Id.Valid = true

	utils.AssertEquals(t, "NullableAuthorisation Valid", true, na.Valid())

}

func TestNullableMovement_Valid(t *testing.T) {

	na := NullableMovement{}

	utils.AssertEquals(t, "NullableMovement Valid", false, na.Valid())

	na.Id.Valid = true

	utils.AssertEquals(t, "NullableMovement Valid", true, na.Valid())

}

func TestNullableCard_Valid(t *testing.T) {

	na := NullableCard{}

	utils.AssertEquals(t, "NullableCard Valid", false, na.Valid())

	na.Id.Valid = true

	utils.AssertEquals(t, "NullableCard Valid", true, na.Valid())
}

func TestNullableAuthorisation_Authorisation(t *testing.T) {

	na := NullableAuthorisation{
		Amount:      sql.NullInt64{Int64: 1000},
		Captured:    sql.NullInt64{Int64: 100},
		CardId:      sql.NullInt64{Int64: 100001},
		Description: sql.NullString{String: "Testing"},
		Id:          sql.NullInt64{Int64: 1001},
		Refunded:    sql.NullInt64{Int64: 99},
		Reversed:    sql.NullInt64{Int64: 98},
		Ts:          sql.NullString{String: "fake"},
		VendorId:    sql.NullInt64{Int64: 1002},
	}

	expected := Authorisation{
		Amount:      1000,
		Captured:    100,
		CardId:      100001,
		Description: "Testing",
		Id:          1001,
		Refunded:    99,
		Reversed:    98,
		Ts:          "fake",
		VendorId:    1002,
	}

	utils.AssertEquals(t, "NullableAuthorisation Authorisation.Amount", expected.Amount, na.Authorisation().Amount)
	utils.AssertEquals(t, "NullableAuthorisation Authorisation.Captured", expected.Captured, na.Authorisation().Captured)
	utils.AssertEquals(t, "NullableAuthorisation Authorisation.CardId", expected.CardId, na.Authorisation().CardId)
	utils.AssertEquals(t, "NullableAuthorisation Authorisation.Description", expected.Description, na.Authorisation().Description)
	utils.AssertEquals(t, "NullableAuthorisation Authorisation.Id", expected.Id, na.Authorisation().Id)
	utils.AssertEquals(t, "NullableAuthorisation Authorisation.Refunded", expected.Refunded, na.Authorisation().Refunded)
	utils.AssertEquals(t, "NullableAuthorisation Authorisation.Reversed", expected.Reversed, na.Authorisation().Reversed)
	utils.AssertEquals(t, "NullableAuthorisation Authorisation.Ts", expected.Ts, na.Authorisation().Ts)
	utils.AssertEquals(t, "NullableAuthorisation Authorisation.VendorId", expected.VendorId, na.Authorisation().VendorId)
}

func TestNullableMovement_Movement(t *testing.T) {
	nm := NullableMovement{
		Id:           sql.NullInt64{Int64: 1001},
		ParentId:     sql.NullInt64{Int64: 10001},
		MovementType: sql.NullString{String: "TEST"},
		Amount:       sql.NullInt64{Int64: 100},
		Description:  sql.NullString{String: "TEST2"},
		Ts:           sql.NullString{String: "fake"},
	}

	expected := Movement{
		Amount:       100,
		CardId:       10001,
		Description:  "TEST2",
		Id:           1001,
		MovementType: "TEST",
		Ts:           "fake",
	}

	utils.AssertEquals(t, "NullableMovement Movement", expected, nm.Movement())
}

func TestNullableMovement_AuthMovement(t *testing.T) {
	nm := NullableMovement{
		Id:           sql.NullInt64{Int64: 1001},
		ParentId:     sql.NullInt64{Int64: 10001},
		MovementType: sql.NullString{String: "TEST"},
		Amount:       sql.NullInt64{Int64: 100},
		Description:  sql.NullString{String: "TEST2"},
		Ts:           sql.NullString{String: "fake"},
	}

	expected := AuthMovement{
		Amount:          100,
		AuthorisationId: 10001,
		Description:     "TEST2",
		MovementType:    "TEST",
		Id:              1001,
		Ts:              "fake",
	}

	utils.AssertEquals(t, "NullableMovement AuthMovement", expected, nm.AuthMovement())
}

func TestNullableCard_Card(t *testing.T) {
	nc := NullableCard{
		Available:  sql.NullInt64{Int64: 1000},
		Balance:    sql.NullInt64{Int64: 2000},
		CustomerId: sql.NullInt64{Int64: 1001},
		Id:         sql.NullInt64{Int64: 100001},
		Ts:         sql.NullString{String: "fake"},
	}

	expected := Card{
		Available:  1000,
		Balance:    2000,
		CustomerId: 1001,
		Id:         100001,
		Ts:         "fake",
	}

	utils.AssertEquals(t, "NullableCard Card.Available", expected.Available, nc.Card().Available)
	utils.AssertEquals(t, "NullableCard Card.Balance", expected.Balance, nc.Card().Balance)
	utils.AssertEquals(t, "NullableCard Card.CustomerId", expected.CustomerId, nc.Card().CustomerId)
	utils.AssertEquals(t, "NullableCard Card.Id", expected.Id, nc.Card().Id)
	utils.AssertEquals(t, "NullableCard Card.Ts", expected.Ts, nc.Card().Ts)
}
