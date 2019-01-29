package models

import (
	"fmt"
	"net/http"
	"database/sql"
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

// Amount that can be captured or reversed
func (auth Authorisation) Capturable() int {
	return auth.Amount - (auth.Captured + auth.Reversed)
}

// Amount that can be refunded
func (auth Authorisation) Refundable() int {
	return auth.Captured - auth.Refunded
}

// Placeholder type which can receive null values in database scans
type NullableMovement struct {
	Id           sql.NullInt64
	ParentId     sql.NullInt64
	MovementType sql.NullString
	Amount       sql.NullInt64
	Description  sql.NullString
	Ts           sql.NullString
}

func (nm NullableMovement) Valid() bool {
	return nm.MovementType.Valid
}

func (nm NullableMovement) Movement() Movement {
	return Movement{
		Id:           int(nm.Id.Int64),
		CardId:       int(nm.ParentId.Int64),
		MovementType: nm.MovementType.String,
		Amount:       int(nm.Amount.Int64),
		Description:  nm.Description.String,
		Ts:           nm.Ts.String,
	}
}

func (nm NullableMovement) AuthMovement() AuthMovement {
	return AuthMovement{
		Id:              int(nm.Id.Int64),
		AuthorisationId: int(nm.ParentId.Int64),
		MovementType:    nm.MovementType.String,
		Amount:          int(nm.Amount.Int64),
		Description:     nm.Description.String,
		Ts:              nm.Ts.String,
	}
}

// Placeholder type which can receive null values in database scans
type NullableAuthorisation struct {
	Amount      sql.NullInt64
	Captured    sql.NullInt64
	CardId      sql.NullInt64
	Description sql.NullString
	Id          sql.NullInt64
	Refunded    sql.NullInt64
	Reversed    sql.NullInt64
	Ts          sql.NullString
	VendorId    sql.NullInt64
}

func (na NullableAuthorisation) Valid() bool {
	return na.Amount.Valid
}

func (na NullableAuthorisation) Authorisation() Authorisation {
	return Authorisation{
		Amount:      int(na.Amount.Int64),
		Captured:    int(na.Captured.Int64),
		CardId:      int(na.CardId.Int64),
		Description: na.Description.String,
		Id:          int(na.Id.Int64),
		Refunded:    int(na.Refunded.Int64),
		Reversed:    int(na.Reversed.Int64),
		Ts:          na.Ts.String,
		VendorId:    int(na.VendorId.Int64),
	}
}

// Placeholder type which can receive null values in database scans
type NullableCard struct {
	Available  sql.NullInt64
	Balance    sql.NullInt64
	CustomerId sql.NullInt64
	Id         sql.NullInt64
	Ts         sql.NullString
}

func (nc NullableCard) Valid() bool {
	return nc.Id.Valid
}

func (nc NullableCard) Card() Card {
	return Card{
		Available:  int(nc.Available.Int64),
		Balance:    int(nc.Balance.Int64),
		CustomerId: int(nc.CustomerId.Int64),
		Id:         int(nc.Id.Int64),
		Ts:         nc.Ts.String,
	}
}