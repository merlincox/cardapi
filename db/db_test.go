package db

import (
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"github.com/merlincox/cardapi/utils"
	"database/sql"
)

func testWrapper(t *testing.T, callback func(*testing.T, *sql.DB, sqlmock.Sqlmock, Dbi)) {

	mockDb, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDb.Close()

	dbi, _ := NewDbi(mockDb)

	callback(t, mockDb, mock, dbi)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetVendors(t *testing.T) {
	testWrapper(t, func(t *testing.T, mockDb *sql.DB, mock sqlmock.Sqlmock, dbi Dbi) {

		rows := sqlmock.NewRows([]string{"id", "fullname"}).
			AddRow(int64(1001), "a shop").
			AddRow(int64(2002), "a pub")

		mock.ExpectPrepare("SELECT id, fullname FROM vendors").ExpectQuery().WillReturnRows(rows)

		vs, apiErr := dbi.GetVendors()

		utils.AssertNoError(t, "Calling GetVendor", apiErr)
		utils.AssertEquals(t, "Size of GetVendor result", 2, len(vs))
		utils.AssertEquals(t, "Fullname for GetVendor result[0]", "a shop", vs[0].Fullname)
		utils.AssertEquals(t, "Id for GetVendor result[0]", 1001, vs[0].Id)
	})
}

func TestGetCustomers(t *testing.T) {
	testWrapper(t, func(t *testing.T, mockDb *sql.DB, mock sqlmock.Sqlmock, dbi Dbi) {

		rows := sqlmock.NewRows([]string{"id", "fullname"}).
			AddRow(int64(1001), "a shop").
			AddRow(int64(2002), "a pub")

		mock.ExpectPrepare("SELECT id, fullname FROM customers").ExpectQuery().WillReturnRows(rows)

		vs, apiErr := dbi.GetCustomers()

		utils.AssertNoError(t, "Calling GetCustomer", apiErr)
		utils.AssertEquals(t, "Size of GetCustomer result", 2, len(vs))
		utils.AssertEquals(t, "Fullname for GetCustomer result[0]", "a shop", vs[0].Fullname)
		utils.AssertEquals(t, "Id for GetCustomer result[0]", 1001, vs[0].Id)
	})
}

