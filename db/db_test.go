package db

import (
	"testing"
	"database/sql"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/merlincox/cardapi/utils"
)

func testWrapper(t *testing.T, callback func(*testing.T, *sql.DB, sqlmock.Sqlmock, Dbi)) {

	mockDb, expecter, err := sqlmock.New()

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	dbi, _ := NewDbi(mockDb)

	callback(t, mockDb, expecter, dbi)

	if err := expecter.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetVendors(t *testing.T) {
	testWrapper(t, func(t *testing.T, mockDb *sql.DB, expecter sqlmock.Sqlmock, dbi Dbi) {

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
	testWrapper(t, func(t *testing.T, mockDb *sql.DB, expecter sqlmock.Sqlmock, dbi Dbi) {

		//mockDb, expecter, err := sqlmock.New()
		//
		//if err != nil {
		//	t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		//}

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

func TestAddVendor(t *testing.T) {
	testWrapper(t, func(t *testing.T, mockDb *sql.DB, expecter sqlmock.Sqlmock, dbi Dbi) {

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
	testWrapper(t, func(t *testing.T, mockDb *sql.DB, expecter sqlmock.Sqlmock, dbi Dbi) {

		expected := sqlmock.NewResult(1001, 1)

		// Not sure why square brackets are needed here..
		expecter.ExpectPrepare("[" + QUERY_ADD_CUSTOMER + "]").ExpectExec().WillReturnResult(expected)

		v, apiErr := dbi.AddCustomer("coffee shop")

		utils.AssertNoError(t, "Calling AddCustomer", apiErr)
		utils.AssertEquals(t, "Fullname for AddCustomer result", "coffee shop", v.Fullname)
		utils.AssertEquals(t, "Id for AddCustomer result", 1001, v.Id)
	})
}
