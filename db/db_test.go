package db

import (
	"testing"
	"fmt"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/merlincox/cardapi/utils"
)

func testWrapper(t *testing.T, callback func(*testing.T, sqlmock.Sqlmock, Dbi)) {

	mockDb, expecter, _ := sqlmock.New()
	dbi, _ := NewDbi(mockDb)
	defer dbi.Close()

	callback(t, expecter, dbi)

	utils.AssertNoError(t, "Calling ExpectationsWereMet", expecter.ExpectationsWereMet())
}

func TestGetVendors(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "fullname"}).
			AddRow(int64(1001), "a shop").
			AddRow(int64(2002), "a pub")

		expecter.ExpectPrepare(QUERY_GET_VENDORS).ExpectQuery().WillReturnRows(expected)

		vs, apiErr := dbi.GetVendors()

		utils.AssertNoError(t, "Calling GetVendors", apiErr)
		utils.AssertEquals(t, "Size of GetVendors result", 2, len(vs))
		utils.AssertEquals(t, "Fullname for GetVendors result[0]", "a shop", vs[0].Fullname)
		utils.AssertEquals(t, "Id for GetVendors result[0]", 1001, vs[0].Id)
	})
}

func TestGetCustomers(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "fullname"}).
			AddRow(int64(1001), "Fred Bloggs").
			AddRow(int64(2002), "Jane Doe")

		expecter.ExpectPrepare(QUERY_GET_CUSTOMERS).ExpectQuery().WillReturnRows(expected)

		vs, apiErr := dbi.GetCustomers()

		utils.AssertNoError(t, "Calling GetCustomers", apiErr)
		utils.AssertEquals(t, "Size of GetCustomers result", 2, len(vs))
		utils.AssertEquals(t, "Fullname for GetCustomers result[0]", "Fred Bloggs", vs[0].Fullname)
		utils.AssertEquals(t, "Id for GetCustomers result[0]", 1001, vs[0].Id)
	})
}

func TestGetCustomer(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "fullname"}).
			AddRow(int64(1001), "Fred Bloggs")

		expecter.ExpectPrepare(QUERY_GET_CUSTOMER).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		v, apiErr := dbi.GetCustomer(1001)

		utils.AssertNoError(t, "Calling GetCustomer", apiErr)
		utils.AssertEquals(t, "Fullname for GetCustomer result", "Fred Bloggs", v.Fullname)
		utils.AssertEquals(t, "Id for GetCustomer result", 1001, v.Id)
	})
}

func TestGetCustomerNotFound(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "fullname"})

		expecter.ExpectPrepare(QUERY_GET_CUSTOMER).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		_, apiErr := dbi.GetCustomer(1001)

		utils.AssertEquals(t, "Return status for calling GetCustomer with a bad id", 404, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling GetCustomer with bad id 1001", fmt.Sprintf(MESSAGE_GET_CUSTOMER_BAD_ID, 1001), apiErr.Error())
	})
}

func TestGetVendor(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "fullname"}).
			AddRow(int64(1001), "Coffee Shop")

		expecter.ExpectPrepare(QUERY_GET_VENDOR).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		v, apiErr := dbi.GetVendor(1001)

		utils.AssertNoError(t, "Calling GetVendor", apiErr)
		utils.AssertEquals(t, "Fullname for GetVendor result", "Coffee Shop", v.Fullname)
		utils.AssertEquals(t, "Id for GetVendor result", 1001, v.Id)
	})
}

func TestGetVendorNotFound(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "fullname"})

		expecter.ExpectPrepare(QUERY_GET_VENDOR).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		_, apiErr := dbi.GetVendor(1001)

		utils.AssertEquals(t, "Return status for calling GetVendor with a bad id", 404, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling GetVendor with bad id 1001", fmt.Sprintf(MESSAGE_GET_VENDOR_BAD_ID, 1001), apiErr.Error())
	})
}

func TestGetCard(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "balance", "available", "tc"}).
			AddRow(int64(1001), 12676, 12089, "2019-01-24 01:00:10")

		expecter.ExpectPrepare(QUERY_GET_CARD).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		v, apiErr := dbi.GetCard(1001)

		utils.AssertNoError(t, "Calling GetCard", apiErr)
		utils.AssertEquals(t, "Balance for GetCard result", 12676, v.Balance)
		utils.AssertEquals(t, "Available for GetCard result", 12089, v.Available)
	})
}

func TestGetCardNotFound(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "balance", "available", "tc"})

		expecter.ExpectPrepare(QUERY_GET_CARD).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		_, apiErr := dbi.GetCard(1001)

		utils.AssertEquals(t, "Return status for calling GetCard with a bad id", 404, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling GetCard with bad id 1001", fmt.Sprintf(MESSAGE_GET_CARD_BAD_ID, 1001), apiErr.Error())
	})
}

func TestAddVendor(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewResult(1001, 1)

		// Not sure why square brackets are needed here..
		expecter.ExpectPrepare("[" + QUERY_ADD_VENDOR + "]").ExpectExec().WithArgs("coffee shop").WillReturnResult(expected)

		v, apiErr := dbi.AddVendor("coffee shop")

		utils.AssertNoError(t, "Calling AddVendor", apiErr)
		utils.AssertEquals(t, "Fullname for AddVendor result", "coffee shop", v.Fullname)
		utils.AssertEquals(t, "Id for AddVendor result", 1001, v.Id)
	})
}

func TestAddCustomer(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewResult(1001, 1)

		// Not sure why square brackets are needed here..
		expecter.ExpectPrepare("[" + QUERY_ADD_CUSTOMER + "]").ExpectExec().WithArgs("Fred Bloggs").WillReturnResult(expected)

		v, apiErr := dbi.AddCustomer("Fred Bloggs")

		utils.AssertNoError(t, "Calling AddCustomer", apiErr)
		utils.AssertEquals(t, "Fullname for AddCustomer result", "Fred Bloggs", v.Fullname)
		utils.AssertEquals(t, "Id for AddCustomer result", 1001, v.Id)
	})
}

func TestAddCardOK(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewResult(1001, 1)

		// Not sure why square brackets are needed here..
		expecter.ExpectPrepare("[" + QUERY_ADD_CARD + "]").ExpectExec().WithArgs(1099).WillReturnResult(expected)

		c, apiErr := dbi.AddCard(1099)

		utils.AssertNoError(t, "Calling AddCard", apiErr)
		utils.AssertEquals(t, "CustomerId for AddCard result", 1099, c.CustomerId)
		utils.AssertEquals(t, "Id for AddCard result", 1001, c.Id)
	})
}

func TestAddCardNotFound(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		err := fmt.Errorf(ERROR_FOREIGN_KEY)

		// Not sure why square brackets are needed here..
		expecter.ExpectPrepare("[" + QUERY_ADD_CARD + "]").ExpectExec().WithArgs(1099).WillReturnError(err)

		_, apiErr := dbi.AddCard(1099)

		utils.AssertEquals(t, "Return status for calling AddCard with a bad customerId", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling AddCard with bad customerId 1099", fmt.Sprintf(MESSAGE_ADD_CARD_BAD_ID, 1099), apiErr.Error())
	})
}

func TestAuthoriseOK(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "fullname"}).
			AddRow(int64(1001), "Coffee Shop")

		expecter.ExpectPrepare(QUERY_GET_VENDOR).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		expected = sqlmock.NewRows([]string{"id", "balance", "available", "tc"}).
			AddRow(int64(100001), 12676, 12089, "2019-01-24 01:00:10")

		expecter.ExpectPrepare(QUERY_GET_CARD).ExpectQuery().WithArgs(100001).WillReturnRows(expected)

		expecter.ExpectBegin()

		expectedR := sqlmock.NewResult(0, 1)

		// Not sure why square brackets are needed here..
		expecter.ExpectPrepare("[" + QUERY_AUTHORISE + "]")
		// This duplication seems to be necessary for tx.Stmt(..)
		expecter.ExpectPrepare("[" + QUERY_AUTHORISE + "]").ExpectExec().WithArgs(210, 100001, 210).WillReturnResult(expectedR)

		expectedR = sqlmock.NewResult(1009, 1)

		// Not sure why square brackets are needed here..
		expecter.ExpectPrepare("[" + QUERY_ADD_AUTHORISATION + "]")
		expecter.ExpectPrepare("[" + QUERY_ADD_AUTHORISATION + "]").ExpectExec().WithArgs(100001, 1001, 210, "Coffee").WillReturnResult(expectedR)

		expecter.ExpectCommit()

		aid, apiErr := dbi.Authorise(100001, 1001, 210, "Coffee")

		utils.AssertNoError(t, "Calling Authorise", apiErr)

		utils.AssertEquals(t, "Authorisation id", 1009, aid)
	})
}


func TestAuthoriseBadAmount(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		aid, apiErr := dbi.Authorise(100001, 1001, 0, "Coffee")

		utils.AssertEquals(t, "Return status for calling Authorise with an invalid amount", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling Authorise with an invalid amount £0.00", fmt.Sprintf(MESSAGE_INVALID_AMOUNT, "Authorise", float32(0)), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling Authorise with an invalid amount", -1, aid)
	})
}

func TestAuthoriseBadVendor(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "fullname"})

		expecter.ExpectPrepare(QUERY_GET_VENDOR).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		aid, apiErr := dbi.Authorise(100001, 1001, 210, "Coffee")

		utils.AssertEquals(t, "Return status for calling Authorise with a bad vendorId", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling Authorise with a bad vendorId 1001", fmt.Sprintf(MESSAGE_INVALID_VENDOR, "Authorise", 1001), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling Authorise with a bad vendorId", -1, aid)
	})
}

func TestAuthoriseBadCard(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "fullname"}).
			AddRow(int64(1001), "Coffee Shop")

		expecter.ExpectPrepare(QUERY_GET_VENDOR).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		expected = sqlmock.NewRows([]string{"id", "balance", "available", "tc"})

		expecter.ExpectPrepare(QUERY_GET_CARD).ExpectQuery().WithArgs(100001).WillReturnRows(expected)

		aid, apiErr := dbi.Authorise(100001, 1001, 210, "Coffee")

		utils.AssertEquals(t, "Return status for calling Authorise with a bad cardId", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling Authorise with a bad cardId 100001", fmt.Sprintf(MESSAGE_INVALID_CARD, "Authorise", 100001), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling Authorise with a bad cardId", -1, aid)
	})
}

func TestAuthoriseInsufficientFunds(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "fullname"}).
			AddRow(int64(1001), "Coffee Shop")

		expecter.ExpectPrepare(QUERY_GET_VENDOR).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		expected = sqlmock.NewRows([]string{"id", "balance", "available", "tc"}).
		AddRow(int64(100001), 12676, 0, "2019-01-24 01:00:10")

		expecter.ExpectPrepare(QUERY_GET_CARD).ExpectQuery().WithArgs(100001).WillReturnRows(expected)

		aid, apiErr := dbi.Authorise(100001, 1001, 210, "Coffee")

		utils.AssertEquals(t, "Return status for calling Authorise with insufficient funds", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling Authorise with insufficient funds 100001", fmt.Sprintf(MESSAGE_AUTHORISE_INSUFFICIENT_AVAILABLE, 2.1, 0.0), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling Authorise with insufficient funds", -1, aid)
	})
}

// Edge case??
func TestAuthoriseInsufficientFunds2(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "fullname"}).
			AddRow(int64(1001), "Coffee Shop")

		expecter.ExpectPrepare(QUERY_GET_VENDOR).ExpectQuery().WithArgs(1001).WillReturnRows(expected)

		expected = sqlmock.NewRows([]string{"id", "balance", "available", "tc"}).
			AddRow(int64(100001), 12676, 211, "2019-01-24 01:00:10")

		expecter.ExpectPrepare(QUERY_GET_CARD).ExpectQuery().WithArgs(100001).WillReturnRows(expected)

		expecter.ExpectBegin()

		expectedR := sqlmock.NewResult(0, 0)

		// Not sure why square brackets are needed here..
		expecter.ExpectPrepare("[" + QUERY_AUTHORISE + "]")
		// This duplication seems to be necessary for tx.Stmt(..)
		expecter.ExpectPrepare("[" + QUERY_AUTHORISE + "]").ExpectExec().WithArgs(210, 100001, 210).WillReturnResult(expectedR)

		expecter.ExpectRollback()

		aid, apiErr := dbi.Authorise(100001, 1001, 210, "Coffee")

		//utils.AssertEquals(t, "Return status for calling Authorise with insufficient funds", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling Authorise with insufficient funds for £2.10", fmt.Sprintf(MESSAGE_AUTHORISE_INSUFFICIENT_AVAILABLE_FOR, 2.1), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling Authorise with insufficient funds", -1, aid)
	})
}

func TestTopUpOK(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "balance", "available", "tc"}).
			AddRow(int64(100001), 12676, 12089, "2019-01-24 01:00:10")

		expecter.ExpectPrepare(QUERY_GET_CARD).ExpectQuery().WithArgs(100001).WillReturnRows(expected)

		expecter.ExpectBegin()

		expectedR := sqlmock.NewResult(0, 1)

		// Not sure why square brackets are needed here..
		expecter.ExpectPrepare("[" + QUERY_UPDATE_CARD + "]")
		// This duplication seems to be necessary for tx.Stmt(..)
		expecter.ExpectPrepare("[" + QUERY_UPDATE_CARD + "]").ExpectExec().WithArgs(2000, 2000, 100001).WillReturnResult(expectedR)

		expectedR = sqlmock.NewResult(1009, 1)

		// Not sure why square brackets are needed here..
		expecter.ExpectPrepare("[" + QUERY_ADD_MOVEMENT + "]")
		expecter.ExpectPrepare("[" + QUERY_ADD_MOVEMENT + "]").ExpectExec().WithArgs(100001, 2000, "Transfer from Bank", "TOP-UP").WillReturnResult(expectedR)

		expecter.ExpectCommit()

		aid, apiErr := dbi.TopUp(100001, 2000, "Transfer from Bank")

		utils.AssertNoError(t, "Calling TopUp", apiErr)

		utils.AssertEquals(t, "Top-up movement id", 1009, aid)
	})
}

func TestTopUpBadAmount(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		aid, apiErr := dbi.TopUp(100001, 0, "Bank Transfer")

		utils.AssertEquals(t, "Return status for calling TopUp with an invalid amount", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling TopUp with an invalid amount £0.00", fmt.Sprintf(MESSAGE_INVALID_AMOUNT, "TopUp", float32(0)), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling TopUp with an invalid amount", -1, aid)
	})
}

func TestTopUpBadCard(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "balance", "available", "tc"})

		expecter.ExpectPrepare(QUERY_GET_CARD).ExpectQuery().WithArgs(100001).WillReturnRows(expected)

		aid, apiErr := dbi.TopUp(100001, 2000, "Transfer from Bank")

		utils.AssertEquals(t, "Return status for calling TopUp with a invalid card", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling TopUp with an invalid card 100001", fmt.Sprintf(MESSAGE_INVALID_CARD, "TopUp", 100001), apiErr.Error())
		utils.AssertEquals(t, "Return status for calling TopUp with an invalid card", -1, aid)
	})
}