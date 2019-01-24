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

		expecter.ExpectPrepare(QUERY_GET_CUSTOMER).ExpectQuery().WillReturnRows(expected)

		v, apiErr := dbi.GetCustomer(1001)

		utils.AssertNoError(t, "Calling GetCustomer", apiErr)
		utils.AssertEquals(t, "Fullname for GetCustomer result", "Fred Bloggs", v.Fullname)
		utils.AssertEquals(t, "Id for GetCustomer result", 1001, v.Id)
	})
}

func TestGetCustomerNotFound(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "fullname"})

		expecter.ExpectPrepare(QUERY_GET_CUSTOMER).ExpectQuery().WillReturnRows(expected)

		_, apiErr := dbi.GetCustomer(1001)

		utils.AssertEquals(t, "Return status for calling GetCustomer with a bad id", 404, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling GetCustomer with bad id 1001", fmt.Sprintf(MESSAGE_GET_CUSTOMER_BAD_ID, 1001), apiErr.Error())
	})
}

func TestGetVendor(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "fullname"}).
			AddRow(int64(1001), "Coffee Shop")

		expecter.ExpectPrepare(QUERY_GET_VENDOR).ExpectQuery().WillReturnRows(expected)

		v, apiErr := dbi.GetVendor(1001)

		utils.AssertNoError(t, "Calling GetVendor", apiErr)
		utils.AssertEquals(t, "Fullname for GetVendor result", "Coffee Shop", v.Fullname)
		utils.AssertEquals(t, "Id for GetVendor result", 1001, v.Id)
	})
}

func TestGetVendorNotFound(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "fullname"})

		expecter.ExpectPrepare(QUERY_GET_VENDOR).ExpectQuery().WillReturnRows(expected)

		_, apiErr := dbi.GetVendor(1001)

		utils.AssertEquals(t, "Return status for calling GetVendor with a bad id", 404, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling GetVendor with bad id 1001", fmt.Sprintf(MESSAGE_GET_VENDOR_BAD_ID, 1001), apiErr.Error())
	})
}

func TestGetCard(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "balance", "available"}).
			AddRow(int64(1001), 12676, 12089)

		expecter.ExpectPrepare(QUERY_GET_CARD).ExpectQuery().WillReturnRows(expected)

		v, apiErr := dbi.GetCard(1001)

		utils.AssertNoError(t, "Calling GetCard", apiErr)
		utils.AssertEquals(t, "Balance for GetCard result", 12676, v.Balance)
		utils.AssertEquals(t, "Available for GetCard result", 12089, v.Available)
	})
}

func TestGetCardNotFound(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "balance", "available"})

		expecter.ExpectPrepare(QUERY_GET_CARD).ExpectQuery().WillReturnRows(expected)

		_, apiErr := dbi.GetCard(1001)

		utils.AssertEquals(t, "Return status for calling GetCard with a bad id", 404, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling GetCard with bad id 1001", fmt.Sprintf(MESSAGE_GET_CARD_BAD_ID, 1001), apiErr.Error())
	})
}

func TestAddVendor(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewResult(1001, 1)

		// Not sure why square brackets are needed here..
		expecter.ExpectPrepare("[" + QUERY_ADD_VENDOR + "]").ExpectExec().WillReturnResult(expected)

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
		expecter.ExpectPrepare("[" + QUERY_ADD_CUSTOMER + "]").ExpectExec().WillReturnResult(expected)

		v, apiErr := dbi.AddCustomer("coffee shop")

		utils.AssertNoError(t, "Calling AddCustomer", apiErr)
		utils.AssertEquals(t, "Fullname for AddCustomer result", "coffee shop", v.Fullname)
		utils.AssertEquals(t, "Id for AddCustomer result", 1001, v.Id)
	})
}

func TestAddCardOK(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewResult(1001, 1)

		// Not sure why square brackets are needed here..
		expecter.ExpectPrepare("[" + QUERY_ADD_CARD + "]").ExpectExec().WillReturnResult(expected)

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
		expecter.ExpectPrepare("[" + QUERY_ADD_CARD + "]").ExpectExec().WillReturnError(err)

		_, apiErr := dbi.AddCard(1099)

		utils.AssertEquals(t, "Return status for calling AddCard with a bad customerId", 400, apiErr.StatusCode())
		utils.AssertEquals(t, "Return message for calling AddCard with bad customerId 1099", fmt.Sprintf(MESSAGE_ADD_CARD_BAD_ID, 1099), apiErr.Error())
	})
}

func TestAuthoriseOK(t *testing.T) {
	testWrapper(t, func(t *testing.T, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewRows([]string{"id", "fullname"}).
			AddRow(int64(1001), "Coffee Shop")

		expecter.ExpectPrepare(QUERY_GET_VENDOR).ExpectQuery().WillReturnRows(expected)

		expected = sqlmock.NewRows([]string{"id", "balance", "available"}).
			AddRow(int64(100001), 12676, 12089)

		expecter.ExpectPrepare(QUERY_GET_CARD).ExpectQuery().WillReturnRows(expected)

		expecter.ExpectBegin()

		expectedR := sqlmock.NewResult(0, 1)

		// Not sure why square brackets are needed here..
		expecter.ExpectPrepare("[" + QUERY_AUTHORISE + "]").ExpectExec().WithArgs(100001, 210).WillReturnResult(expectedR)

		expectedR = sqlmock.NewResult(1009, 1)

		// Not sure why square brackets are needed here..
		expecter.ExpectPrepare("[" + QUERY_ADD_AUTHORISATION + "]").ExpectExec().WillReturnResult(expectedR)

		expecter.ExpectCommit()

		aid, apiErr := dbi.Authorise(100001, 1001, 210, "Coffee")

		utils.AssertNoError(t, "Calling Authorise", apiErr)

		utils.AssertEquals(t, "Authorisation id", 1009, aid)

		//@TODO

	})
}