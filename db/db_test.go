package db

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/merlincox/cardapi/models"
	"github.com/merlincox/cardapi/utils"
)

var (
	// RE special chars which may occur in queries
	reSpecial = []string{
		"?",
		".",
		"(",
		")",
		"+",
	}
)

func badIdMessage(methodName, objectName string, id int) string {
	return fmt.Sprintf(MESSAGE_BAD_ID, methodName, objectName, id)
}

//NB ExpectPrepare takes regular expressions, so query constants may need to be escaped
func esc(src string) string {

	for _, s := range reSpecial {
		src = strings.Replace(src, s, "\\"+s, -1)
	}

	return src
}

func testWrapper(t *testing.T, callback func(*testing.T, sqlmock.Sqlmock, Dbi)) {

	mockDb, expecter, _ := sqlmock.New()
	dbi, _ := NewDbi(mockDb)
	defer dbi.Close()

	callback(t, expecter, dbi)

	utils.AssertNoError(t, "Calling ExpectationsWereMet", expecter.ExpectationsWereMet())
}

func TestGetVendors(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "vendor_name", "balance"}).
			AddRow(int64(1001), "a shop", 1234).
			AddRow(int64(2002), "a pub", 999)

		expecter.ExpectPrepare(esc(QUERY_GET_VENDORS)).ExpectQuery().WillReturnRows(expected)

		vs, apiErr := dbi.GetVendors()

		utils.AssertNoError(t, "Calling GetVendors", apiErr)
		utils.AssertEquals(t, "Size of GetVendors result", 2, len(vs))
		utils.AssertEquals(t, "VendorName for GetVendors result[0]", "a shop", vs[0].VendorName)
		utils.AssertEquals(t, "Id for GetVendors result[0]", 1001, vs[0].Id)
		utils.AssertEquals(t, "Balance for GetVendors result[0]", 1234, vs[0].Balance)
	})
}

func TestGetCustomers(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "fullname"}).
			AddRow(int64(1001), "Fred Bloggs").
			AddRow(int64(2002), "Jane Doe")

		expecter.ExpectPrepare(esc(QUERY_GET_CUSTOMERS)).ExpectQuery().WillReturnRows(expected)

		vs, apiErr := dbi.GetCustomers()

		utils.AssertNoError(t, "Calling GetCustomers", apiErr)
		utils.AssertEquals(t, "Size of GetCustomers result", 2, len(vs))
		utils.AssertEquals(t, "VendorName for GetCustomers result[0]", "Fred Bloggs", vs[0].Fullname)
		utils.AssertEquals(t, "Id for GetCustomers result[0]", 1001, vs[0].Id)
	})
}

func TestGetCustomer(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		//cu.id, cu.fullname, c.id, c.balance, c.available, c.ts
		expected := sqlmock.NewRows([]string{"cu.id", "cu.fullname", "c.id", "c.balance", "c.available", "c.ts"}).
			AddRow(int64(1001), "Fred Bloggs", 1001, 456, 0, "2019-01-24 01:00:10")

		expecter.ExpectPrepare(esc(QUERY_GET_CUSTOMER_ALL)).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		c, apiErr := dbi.GetCustomer(1001)

		utils.AssertNoError(t, "Calling GetCustomer", apiErr)
		utils.AssertEquals(t, "Fullname for GetCustomer result", "Fred Bloggs", c.Fullname)
		utils.AssertEquals(t, "Id for GetCustomer result", 1001, c.Id)
		utils.AssertEquals(t, "len(Cards) for GetCustomer result", 1, len(c.Cards))
	})
}

func TestGetCustomerNotFound(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"cu.id", "cu.fullname", "c.id", "c.balance", "c.available", "c.ts"})

		expecter.ExpectPrepare(esc(QUERY_GET_CUSTOMER_ALL)).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		_, apiErr := dbi.GetCustomer(1001)

		utils.AssertEquals(t, "Return status for calling GetCustomer with a bad id", 404, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling GetCustomer with bad id 1001", badIdMessage("GetCustomer", "customer", 1001), apiErr.Error())
	})
}

func TestGetAuthorisation(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		//a.id, a.amount, a.card_id, a.vendor_id, a.description, a.captured, a.reversed, a.refunded, m.amount, m.description, m.movement_type, m.ts
		expected := sqlmock.NewRows([]string{"a.id", "a.amount", "a.card_id", "a.vendor_id", "a.description", "a.captured", "a.reversed", "a.refunded", "m.id", "m.amount", "m.description", "m.movement_type", "m.ts"}).
			AddRow(int64(1001), 250, 100001, 1002, "cake", 0, 250, 0, 1009, 250, "cake bad", "REVERSAL", "2019-01-24 01:00:10")

		expecter.ExpectPrepare(esc(QUERY_GET_AUTHORISATION_ALL)).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		a, apiErr := dbi.GetAuthorisation(1001)

		utils.AssertNoError(t, "Calling GetAuthorisation", apiErr)
		utils.AssertEquals(t, "Amount for GetAuthorisation result", 250, a.Amount)
		utils.AssertEquals(t, "Id for GetAuthorisation result", 1001, a.Id)
		utils.AssertEquals(t, "len(Movements) for GetAuthorisation result", 1, len(a.Movements))
	})
}

func TestGetAuthorisationNotFound(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"a.id", "a.amount", "a.card_id", "a.vendor_id", "a.description", "a.captured", "a.reversed", "a.refunded", "m.id", "m.amount", "m.description", "m.movement_type", "m.ts"})

		expecter.ExpectPrepare(esc(QUERY_GET_AUTHORISATION_ALL)).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		_, apiErr := dbi.GetAuthorisation(1001)

		utils.AssertEquals(t, "Return status for calling GetAuthorisation with a bad id", 404, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling GetAuthorisation with bad id 1001", badIdMessage("GetAuthorisation", "authorisation", 1001), apiErr.Error())
	})
}

func TestGetVendor(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		//v.id, v.vendor_name, v.balance, a.id, a.amount, a.card_id, a.description, a.captured, a.reversed, a.refunded, a.ts
		expected := sqlmock.NewRows([]string{"v.id", "v.vendor_name", "v.balance", "a.id", "a.amount", "a.card_id", "a.description", "a.captured", "a.reversed", "a.refunded", "a.ts"}).
			AddRow(int64(1001), "Coffee Shop", 0, 99, 210, 10001, "Cake", 0, 0, 0, "2019-01-24 01:00:10").
			AddRow(int64(1001), "Coffee Shop", 0, 99, 150, 10001, "Coffee", 0, 0, 0, "2019-01-24 01:00:10")

		expecter.ExpectPrepare(esc(QUERY_GET_VENDOR_ALL)).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		v, apiErr := dbi.GetVendor(1001)

		utils.AssertNoError(t, "Calling GetVendor", apiErr)
		utils.AssertEquals(t, "VendorName for GetVendor result", "Coffee Shop", v.VendorName)
		utils.AssertEquals(t, "Id for GetVendor result", 1001, v.Id)
		utils.AssertEquals(t, "len(Authorisations) for GetVendor result", 2, len(v.Authorisations))
		utils.AssertEquals(t, "Authorisations[0].Description for GetVendor result", "Cake", v.Authorisations[0].Description)
	})
}

func TestGetVendorNotFound(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"v.id", "v.vendor_name", "v.balance", "a.amount", "a.card_id", "a.description", "a.captured", "a.reversed", "a.refunded"})

		expecter.ExpectPrepare(esc(QUERY_GET_VENDOR_ALL)).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		_, apiErr := dbi.GetVendor(1001)

		utils.AssertEquals(t, "Return status for calling GetVendor with a bad id", 404, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling GetVendor with bad id 1001", badIdMessage("GetVendor", "vendor", 1001), apiErr.Error())
	})
}

func TestGetCard(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		//  c.id, c.balance, c.available, c.ts, m.amount, m.description, m.movement_type, m.ts
		expected := sqlmock.NewRows([]string{"c.id", "c.balance", "c.available", "c.tc", "m.id", "m.amount", "m.description", "m.movement_type", "m.ts"}).
			AddRow(int64(1001), 12676, 12089, "2019-01-24 01:00:10", 1001, 95, "Cake", "PURCHASE", "2019-01-24 01:00:10")

		expecter.ExpectPrepare(esc(QUERY_GET_CARD_ALL)).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		v, apiErr := dbi.GetCard(1001)

		utils.AssertNoError(t, "Calling GetCard", apiErr)
		utils.AssertEquals(t, "Balance for GetCard result", 12676, v.Balance)
		utils.AssertEquals(t, "Available for GetCard result", 12089, v.Available)
		utils.AssertEquals(t, "len(Movements) for GetCard result", 1, len(v.Movements))
		utils.AssertEquals(t, "Movements[0].Description for GetCard result", "Cake", v.Movements[0].Description)
	})
}

func TestGetCardNotFound(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"c.id", "c.balance", "c.available", "c.tc", "m.amount", "m.description", "m.movement_type", "m.ts"})

		expecter.ExpectPrepare(esc(QUERY_GET_CARD_ALL)).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		_, apiErr := dbi.GetCard(1001)

		utils.AssertEquals(t, "Return status for calling GetCard with a bad id", 404, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling GetCard with bad id 1001", badIdMessage("GetCard", "card", 1001), apiErr.Error())
	})
}

func TestAddVendor(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		v := models.Vendor{
			VendorName: "coffee shop",
		}

		expected := sqlmock.NewResult(1001, 1)

		expecter.ExpectPrepare(esc(QUERY_ADD_VENDOR)).ExpectExec().WithArgs("coffee shop").WillReturnResult(expected)

		v, apiErr := dbi.AddOrUpdateVendor(v)

		utils.AssertNoError(t, "Calling AddOrUpdateVendor without an id", apiErr)
		utils.AssertEquals(t, "VendorName for AddOrUpdateVendor without an id", "coffee shop", v.VendorName)
		utils.AssertEquals(t, "Id for AddOrUpdateVendor without an id", 1001, v.Id)
	})
}

func TestUpdateVendor(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		v := models.Vendor{
			VendorName: "coffee shop",
			Id:         1002,
		}

		expected := sqlmock.NewResult(0, 1)

		expecter.ExpectPrepare(esc(QUERY_UPDATE_VENDOR_DETAILS)).ExpectExec().WithArgs("coffee shop").WillReturnResult(expected)

		v, apiErr := dbi.AddOrUpdateVendor(v)

		utils.AssertNoError(t, "Calling AddOrUpdateVendor with an id", apiErr)
		utils.AssertEquals(t, "VendorName for AddOrUpdateVendor with an id", "coffee shop", v.VendorName)
		utils.AssertEquals(t, "Id for AddOrUpdateVendor with an id of 1002", 1002, v.Id)
	})
}

func TestAddCustomer(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		c := models.Customer{
			Fullname: "Fred Bloggs",
		}

		expected := sqlmock.NewResult(1001, 1)

		expecter.ExpectPrepare(esc(QUERY_ADD_CUSTOMER)).ExpectExec().WithArgs("Fred Bloggs").WillReturnResult(expected)

		c, apiErr := dbi.AddOrUpdateCustomer(c)

		utils.AssertNoError(t, "Calling AddOrUpdateCustomer without an id", apiErr)
		utils.AssertEquals(t, "VendorName for AddOrUpdateCustomer without an id", "Fred Bloggs", c.Fullname)
		utils.AssertEquals(t, "Id for AddOrUpdateCustomer without an id", 1001, c.Id)
	})
}

func TestUpdateCustomer(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		c := models.Customer{
			Fullname: "Fred Bloggs",
			Id:       1001,
		}

		expected := sqlmock.NewResult(0, 1)

		expecter.ExpectPrepare(esc(QUERY_UPDATE_CUSTOMER_DETAILS)).ExpectExec().WithArgs("Fred Bloggs").WillReturnResult(expected)

		c, apiErr := dbi.AddOrUpdateCustomer(c)

		utils.AssertNoError(t, "Calling AddOrUpdateCustomer with an id", apiErr)
		utils.AssertEquals(t, "VendorName for AddOrUpdateCustomer with an id", "Fred Bloggs", c.Fullname)
		utils.AssertEquals(t, "Id for AddOrUpdateCustomer without an id of 1001", 1001, c.Id)
	})
}

func TestAddCardOK(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewResult(1001, 1)

		expecter.ExpectPrepare(esc(QUERY_ADD_CARD)).ExpectExec().WithArgs(1099).WillReturnResult(expected)

		c, apiErr := dbi.AddCard(1099)

		utils.AssertNoError(t, "Calling AddCard", apiErr)
		utils.AssertEquals(t, "CustomerId for AddCard result", 1099, c.CustomerId)
		utils.AssertEquals(t, "Id for AddCard result", 1001, c.Id)
	})
}

func TestAddCardNotFound(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		err := &mysql.MySQLError{
			Number:  MYSQL_ERROR_FOREIGN_KEY,
			Message: "(Foreign key violation)",
		}

		expecter.ExpectPrepare(esc(QUERY_ADD_CARD)).ExpectExec().WithArgs(1099).WillReturnError(err)

		_, apiErr := dbi.AddCard(1099)

		utils.AssertEquals(t, "Return status for calling AddCard with a bad customerId", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling AddCard with bad customerId 1099", badIdMessage("AddCard", "customer", 1099), apiErr.Error())
	})
}

func TestAuthoriseOK(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "vendor_name", "balance"}).
			AddRow(int64(1001), "Coffee Shop", 999)

		expecter.ExpectPrepare(esc(QUERY_GET_VENDOR)).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		expected = sqlmock.NewRows([]string{"id", "balance", "available", "tc"}).
			AddRow(int64(100001), 12676, 12089, "2019-01-24 01:00:10")

		expecter.ExpectPrepare(esc(QUERY_GET_CARD)).ExpectQuery().WithArgs(100001).WillReturnRows(expected)

		expecter.ExpectBegin()

		expectedR := sqlmock.NewResult(0, 1)

		expecter.ExpectPrepare(esc(QUERY_UPDATE_CARD))
		// This duplication seems to be necessary for tx.Stmt(..)
		expecter.ExpectPrepare(esc(QUERY_UPDATE_CARD)).ExpectExec().WithArgs(0, 210, 100001, 210).WillReturnResult(expectedR)

		expectedR = sqlmock.NewResult(1009, 1)

		expecter.ExpectPrepare(esc(QUERY_ADD_AUTHORISATION))
		expecter.ExpectPrepare(esc(QUERY_ADD_AUTHORISATION)).ExpectExec().WithArgs(100001, 1001, 210, "Coffee").WillReturnResult(expectedR)

		expecter.ExpectCommit()

		aid, apiErr := dbi.Authorise(100001, 1001, 210, "Coffee")

		utils.AssertNoError(t, "Calling Authorise", apiErr)

		utils.AssertEquals(t, "Authorisation id", 1009, aid)
	})
}

func TestAuthoriseBadVendor(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "vendor_name", "balance"})

		expecter.ExpectPrepare(esc(QUERY_GET_VENDOR)).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		aid, apiErr := dbi.Authorise(100001, 1001, 210, "Coffee")

		utils.AssertEquals(t, "Return status for calling Authorise with a bad vendorId", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling Authorise with a bad vendorId 1001", badIdMessage("Authorise", "vendor", 1001), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling Authorise with a bad vendorId", -1, aid)
	})
}

func TestAuthoriseBadCard(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "vendor_name", "balance"}).
			AddRow(int64(1001), "Coffee Shop", 999)

		expecter.ExpectPrepare(esc(QUERY_GET_VENDOR)).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		expected = sqlmock.NewRows([]string{"id", "balance", "available", "tc"})

		expecter.ExpectPrepare(esc(QUERY_GET_CARD)).ExpectQuery().WithArgs(100001).WillReturnRows(expected)

		aid, apiErr := dbi.Authorise(100001, 1001, 210, "Coffee")

		utils.AssertEquals(t, "Return status for calling Authorise with a bad cardId", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling Authorise with a bad cardId 100001", badIdMessage("Authorise", "card", 100001), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling Authorise with a bad cardId", -1, aid)
	})
}

func TestAuthoriseInsufficientFunds(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "vendor_name", "balance"}).
			AddRow(int64(1001), "Coffee Shop", 999)

		expecter.ExpectPrepare(esc(QUERY_GET_VENDOR)).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		expected = sqlmock.NewRows([]string{"id", "balance", "available", "tc"}).
			AddRow(int64(100001), 12676, 0, "2019-01-24 01:00:10")

		expecter.ExpectPrepare(esc(QUERY_GET_CARD)).ExpectQuery().WithArgs(100001).WillReturnRows(expected)

		aid, apiErr := dbi.Authorise(100001, 1001, 210, "Coffee")

		utils.AssertEquals(t, "Return status for calling Authorise with insufficient funds", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling Authorise with insufficient funds 100001", fmt.Sprintf(MESSAGE_INSUFFICIENT_AVAILABLE, "Authorise", 2.1, 0.0), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling Authorise with insufficient funds", -1, aid)
	})
}

// Edge case??
func TestAuthoriseInsufficientFunds2(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "vendor_name", "balance"}).
			AddRow(int64(1001), "Coffee Shop", 999)

		expecter.ExpectPrepare(esc(QUERY_GET_VENDOR)).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		expected = sqlmock.NewRows([]string{"id", "balance", "available", "tc"}).
			AddRow(int64(100001), 12676, 211, "2019-01-24 01:00:10")

		expecter.ExpectPrepare(esc(QUERY_GET_CARD)).ExpectQuery().WithArgs(100001).WillReturnRows(expected)

		expecter.ExpectBegin()

		expectedR := sqlmock.NewResult(0, 0)

		expecter.ExpectPrepare(esc(QUERY_UPDATE_CARD))
		// This duplication seems to be necessary for tx.Stmt(..)
		expecter.ExpectPrepare(esc(QUERY_UPDATE_CARD)).ExpectExec().WithArgs(0, 210, 100001, 210).WillReturnResult(expectedR)

		expecter.ExpectRollback()

		aid, apiErr := dbi.Authorise(100001, 1001, 210, "Coffee")

		utils.AssertEquals(t, "Return status for calling Authorise with insufficient funds", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling Authorise with insufficient funds for £2.10", fmt.Sprintf(MESSAGE_INSUFFICIENT_AVAILABLE_FOR, "Authorise", 2.1), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling Authorise with insufficient funds", -1, aid)
	})
}

func TestTopUpOK(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "balance", "available", "tc"}).
			AddRow(int64(100001), 12676, 12089, "2019-01-24 01:00:10")

		expecter.ExpectPrepare(esc(QUERY_GET_CARD)).ExpectQuery().WithArgs(100001).WillReturnRows(expected)

		expecter.ExpectBegin()

		expectedR := sqlmock.NewResult(0, 1)

		expecter.ExpectPrepare(esc(QUERY_UPDATE_CARD))
		// This duplication seems to be necessary for tx.Stmt(..)
		expecter.ExpectPrepare(esc(QUERY_UPDATE_CARD)).ExpectExec().WithArgs(2000, 2000, 100001).WillReturnResult(expectedR)

		expectedR = sqlmock.NewResult(1009, 1)

		expecter.ExpectPrepare(esc(QUERY_ADD_MOVEMENT))
		expecter.ExpectPrepare(esc(QUERY_ADD_MOVEMENT)).ExpectExec().WithArgs(100001, 2000, "Transfer from Bank", "TOP-UP").WillReturnResult(expectedR)

		expecter.ExpectCommit()

		aid, apiErr := dbi.TopUp(100001, 2000, "Transfer from Bank")

		utils.AssertNoError(t, "Calling TopUp", apiErr)

		utils.AssertEquals(t, "Top-up movement id", 1009, aid)
	})
}

func TestTopUpBadCard(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "balance", "available", "tc"})

		expecter.ExpectPrepare(esc(QUERY_GET_CARD)).ExpectQuery().WithArgs(100001).WillReturnRows(expected)

		aid, apiErr := dbi.TopUp(100001, 2000, "Transfer from Bank")

		utils.AssertEquals(t, "Return status for calling TopUp with a invalid card", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling TopUp with an invalid card 100001", badIdMessage("TopUp", "card", 100001), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling TopUp with an invalid card", -1, aid)
	})
}

func TestCaptureOK(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		//id, amount, card_id, vendor_id, description, captured, reversed, refunded
		expected := sqlmock.NewRows([]string{"id", "amount", "card_id", "vendor_id", "description", "captured", "reversed", "refunded"}).
			AddRow(int64(1005), 250, 100001, 1002, "Coffee", 0, 0, 0)

		expecter.ExpectPrepare(esc(QUERY_GET_AUTHORISATION)).ExpectQuery().WithArgs(1005).WillReturnRows(expected)

		expecter.ExpectBegin()

		expectedR := sqlmock.NewResult(0, 1)

		expecter.ExpectPrepare(esc(QUERY_UPDATE_CARD))
		// This duplication seems to be necessary for tx.Stmt(..)
		expecter.ExpectPrepare(esc(QUERY_UPDATE_CARD)).ExpectExec().WithArgs(-250, 0, 100001).WillReturnResult(expectedR)

		expectedR = sqlmock.NewResult(0, 1)

		expecter.ExpectPrepare(esc(QUERY_UPDATE_AUTH))
		expecter.ExpectPrepare(esc(QUERY_UPDATE_AUTH)).ExpectExec().WithArgs(250, 0, 0, 1005).WillReturnResult(expectedR)

		expectedR = sqlmock.NewResult(0, 1)

		expecter.ExpectPrepare(esc(QUERY_UPDATE_VENDOR))
		expecter.ExpectPrepare(esc(QUERY_UPDATE_VENDOR)).ExpectExec().WithArgs(250, 1002).WillReturnResult(expectedR)

		expectedR = sqlmock.NewResult(1009, 1)

		expecter.ExpectPrepare(esc(QUERY_ADD_MOVEMENT))
		expecter.ExpectPrepare(esc(QUERY_ADD_MOVEMENT)).ExpectExec().WithArgs(100001, -250, "Coffee", "PURCHASE").WillReturnResult(expectedR)

		expectedR = sqlmock.NewResult(1009, 1)
		//auth.Id, -amount, description, "REVERSAL"

		expecter.ExpectPrepare(esc(QUERY_ADD_AUTH_MOVEMENT))
		expecter.ExpectPrepare(esc(QUERY_ADD_AUTH_MOVEMENT)).ExpectExec().WithArgs(1005, 250, "Capture of £2.50", "CAPTURE").WillReturnResult(expectedR)

		expecter.ExpectCommit()

		aid, apiErr := dbi.Capture(1005, 250)

		utils.AssertNoError(t, "Calling Capture", apiErr)

		utils.AssertEquals(t, "Capture id", 1009, aid)
	})
}

func TestCaptureBadId(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		//id, amount, card_id, vendor_id, description, captured, reversed, refunded
		expected := sqlmock.NewRows([]string{"id", "amount", "card_id", "vendor_id", "description", "captured", "reversed", "refunded"})

		expecter.ExpectPrepare(esc(QUERY_GET_AUTHORISATION)).ExpectQuery().WithArgs(1005).WillReturnRows(expected)

		aid, apiErr := dbi.Capture(1005, 250)

		utils.AssertEquals(t, "Return status for calling Capture with bad id", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling Capture with bad id", fmt.Sprintf(MESSAGE_BAD_ID, "Capture", "authorisation", 1005), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling Capture with insufficient uncaptured funds", -1, aid)
	})
}

func TestCaptureInsufficient(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		//id, amount, card_id, vendor_id, description, captured, reversed, refunded
		expected := sqlmock.NewRows([]string{"id", "amount", "card_id", "vendor_id", "description", "captured", "reversed", "refunded"}).
			AddRow(int64(1005), 250, 100001, 1002, "Coffee", 250, 0, 0)

		expecter.ExpectPrepare(esc(QUERY_GET_AUTHORISATION)).ExpectQuery().WithArgs(1005).WillReturnRows(expected)

		aid, apiErr := dbi.Capture(1005, 250)

		utils.AssertEquals(t, "Return status for calling Capture with insufficient uncaptured funds", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling Capture with insufficient uncaptured funds for £2.50", fmt.Sprintf(MESSAGE_INSUFFICIENT_AVAILABLE, "Capture", 2.5, 0.0), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling Capture with insufficient uncaptured funds", -1, aid)
	})
}

func TestRefundOK(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		//id, amount, card_id, vendor_id, description, refundd, reversed, refunded
		expected := sqlmock.NewRows([]string{"id", "amount", "card_id", "vendor_id", "description", "captured", "reversed", "refunded"}).
			AddRow(int64(1005), 250, 100001, 1002, "Coffee", 250, 0, 0)

		expecter.ExpectPrepare(esc(QUERY_GET_AUTHORISATION)).ExpectQuery().WithArgs(1005).WillReturnRows(expected)

		expecter.ExpectBegin()

		expectedR := sqlmock.NewResult(0, 1)

		expecter.ExpectPrepare(esc(QUERY_UPDATE_CARD))
		// This duplication seems to be necessary for tx.Stmt(..)
		expecter.ExpectPrepare(esc(QUERY_UPDATE_CARD)).ExpectExec().WithArgs(250, 250, 100001).WillReturnResult(expectedR)

		expectedR = sqlmock.NewResult(0, 1)

		expecter.ExpectPrepare(esc(QUERY_UPDATE_AUTH))
		expecter.ExpectPrepare(esc(QUERY_UPDATE_AUTH)).ExpectExec().WithArgs(0, 250, 0, 1005).WillReturnResult(expectedR)

		expectedR = sqlmock.NewResult(0, 1)

		expecter.ExpectPrepare(esc(QUERY_UPDATE_VENDOR))
		expecter.ExpectPrepare(esc(QUERY_UPDATE_VENDOR)).ExpectExec().WithArgs(-250, 1002).WillReturnResult(expectedR)

		expectedR = sqlmock.NewResult(1009, 1)

		expecter.ExpectPrepare(esc(QUERY_ADD_MOVEMENT))
		expecter.ExpectPrepare(esc(QUERY_ADD_MOVEMENT)).ExpectExec().WithArgs(100001, 250, "Bad coffee", "REFUND").WillReturnResult(expectedR)

		expectedR = sqlmock.NewResult(1009, 1)
		//auth.Id, -amount, description, "REVERSAL"

		expecter.ExpectPrepare(esc(QUERY_ADD_AUTH_MOVEMENT))
		expecter.ExpectPrepare(esc(QUERY_ADD_AUTH_MOVEMENT)).ExpectExec().WithArgs(1005, -250, "Bad coffee", "REFUND").WillReturnResult(expectedR)

		expecter.ExpectCommit()

		aid, apiErr := dbi.Refund(1005, 250, "Bad coffee")

		utils.AssertNoError(t, "Calling Refund", apiErr)

		utils.AssertEquals(t, "Refund id", 1009, aid)
	})
}

func TestRefundBadId(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		//id, amount, card_id, vendor_id, description, refundd, reversed, refunded
		expected := sqlmock.NewRows([]string{"id", "amount", "card_id", "vendor_id", "description", "captured", "reversed", "refunded"})

		expecter.ExpectPrepare(esc(QUERY_GET_AUTHORISATION)).ExpectQuery().WithArgs(1005).WillReturnRows(expected)

		aid, apiErr := dbi.Refund(1005, 250, "Bad coffee")

		utils.AssertEquals(t, "Return status for calling Refund with bad id", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling Refund with bad id", fmt.Sprintf(MESSAGE_BAD_ID, "Refund", "authorisation", 1005), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling Refund with insufficient unrefundd funds", -1, aid)
	})
}

func TestRefundInsufficient(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		//id, amount, card_id, vendor_id, description, refundd, reversed, refunded
		expected := sqlmock.NewRows([]string{"id", "amount", "card_id", "vendor_id", "description", "captured", "reversed", "refunded"}).
			AddRow(int64(1005), 250, 100001, 1002, "Coffee", 0, 0, 0)

		expecter.ExpectPrepare(esc(QUERY_GET_AUTHORISATION)).ExpectQuery().WithArgs(1005).WillReturnRows(expected)

		aid, apiErr := dbi.Refund(1005, 250, "Bad coffee")

		utils.AssertEquals(t, "Return status for calling Refund with insufficient unrefundd funds", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling Refund with insufficient captured funds for £2.50", fmt.Sprintf(MESSAGE_INSUFFICIENT_AVAILABLE, "Refund", 2.5, 0.0), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling Refund with insufficient captured funds", -1, aid)
	})
}

func TestReverseOK(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		//id, amount, card_id, vendor_id, description, reversed, reversed, refunded
		expected := sqlmock.NewRows([]string{"id", "amount", "card_id", "vendor_id", "description", "captured", "reversed", "refunded"}).
			AddRow(int64(1005), 250, 100001, 1002, "Coffee", 0, 0, 0)

		expecter.ExpectPrepare(esc(QUERY_GET_AUTHORISATION)).ExpectQuery().WithArgs(1005).WillReturnRows(expected)

		expecter.ExpectBegin()

		expectedR := sqlmock.NewResult(0, 1)

		expecter.ExpectPrepare(esc(QUERY_UPDATE_CARD))
		// This duplication seems to be necessary for tx.Stmt(..)
		expecter.ExpectPrepare(esc(QUERY_UPDATE_CARD)).ExpectExec().WithArgs(0, 250, 100001).WillReturnResult(expectedR)

		expectedR = sqlmock.NewResult(0, 1)

		expecter.ExpectPrepare(esc(QUERY_UPDATE_AUTH))
		expecter.ExpectPrepare(esc(QUERY_UPDATE_AUTH)).ExpectExec().WithArgs(0, 0, 250, 1005).WillReturnResult(expectedR)

		expectedR = sqlmock.NewResult(1009, 1)
		//auth.Id, -amount, description, "REVERSAL"

		expecter.ExpectPrepare(esc(QUERY_ADD_AUTH_MOVEMENT))
		expecter.ExpectPrepare(esc(QUERY_ADD_AUTH_MOVEMENT)).ExpectExec().WithArgs(1005, -250, "Bad coffee", "REVERSAL").WillReturnResult(expectedR)

		expecter.ExpectCommit()

		aid, apiErr := dbi.Reverse(1005, 250, "Bad coffee")

		utils.AssertNoError(t, "Calling Reverse", apiErr)

		utils.AssertEquals(t, "Reversal id", 1009, aid)
	})
}

func TestReverseBadId(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		//id, amount, card_id, vendor_id, description, reversed, reversed, refunded
		expected := sqlmock.NewRows([]string{"id", "amount", "card_id", "vendor_id", "description", "captured", "reversed", "refunded"})

		expecter.ExpectPrepare(esc(QUERY_GET_AUTHORISATION)).ExpectQuery().WithArgs(1005).WillReturnRows(expected)

		aid, apiErr := dbi.Reverse(1005, 250, "Bad coffee")

		utils.AssertEquals(t, "Return status for calling Reverse with bad id", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling Reverse with bad id", fmt.Sprintf(MESSAGE_BAD_ID, "Reverse", "authorisation", 1005), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling Reverse with insufficient unreversed funds", -1, aid)
	})
}

func TestReverseInsufficient(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		//id, amount, card_id, vendor_id, description, reversed, reversed, refunded
		expected := sqlmock.NewRows([]string{"id", "amount", "card_id", "vendor_id", "description", "captured", "reversed", "refunded"}).
			AddRow(int64(1005), 250, 100001, 1002, "Coffee", 250, 0, 0)

		expecter.ExpectPrepare(esc(QUERY_GET_AUTHORISATION)).ExpectQuery().WithArgs(1005).WillReturnRows(expected)

		aid, apiErr := dbi.Reverse(1005, 250, "Bad coffee")

		utils.AssertEquals(t, "Return status for calling Reverse with insufficient unreversed funds", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling Reverse with insufficient captured funds for £2.50", fmt.Sprintf(MESSAGE_INSUFFICIENT_AVAILABLE, "Reverse", 2.5, 0.0), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling Reverse with insufficient captured funds", -1, aid)
	})
}
