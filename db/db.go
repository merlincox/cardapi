package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"github.com/merlincox/cardapi/models"
	"log"
)

type Dbi interface {
	//AddCustomer(fullname string) (models.Customer, models.ApiError)
	//GetCustomer(id int) (models.Customer, models.ApiError)
	GetCustomers() ([]models.Customer, models.ApiError)
	//AddVendor(fullname string) (models.Vendor, models.ApiError)
	//GetVendor(id int) (models.Vendor, models.ApiError)
	GetVendors() ([]models.Vendor, models.ApiError)

	Ping() models.ApiError
	Close()
}

type dbGate struct{}

var (
	dbd   *dbGate
	dbx   *sql.DB
	stmts = make(map[string]*sql.Stmt, 10)
)

func NewDbi(injected *sql.DB) (Dbi, models.ApiError) {

	if dbd == nil {

		if injected == nil {

			s, err := sql.Open("mysql", "merlincox:merlincox@tcp(www.db4free.net:3306)/merlincox")

			if err != nil {
				return nil, models.ErrorWrap(err)
			}

			dbx = s

		} else {
			dbx = injected
		}

		dbd = &dbGate{}
	}

	return dbd, nil
}

func (d *dbGate) Ping() models.ApiError {

	err := dbx.Ping()

	if err != nil {
		return models.ErrorWrap(err)
	}

	return nil
}

func (d *dbGate) Close() {

	for qry, stmt := range stmts {
		log.Printf("Closing prepared query '%v'\n", qry)
		stmt.Close()
	}

	if dbx != nil {
		log.Printf("Closing connection\n")
		dbx.Close()
	}
}

func (d *dbGate) GetVendors() ([]models.Vendor, models.ApiError) {

	var (
		vs  []models.Vendor
		v   models.Vendor
		err error
	)

	qry := "SELECT id, fullname FROM vendors"

	_, already := stmts[qry]

	if !already {

		stmts[qry], err = dbx.Prepare(qry)

		if err != nil {
			return vs, models.ErrorWrap(err)
		}
	}

	rows, err := stmts[qry].Query()

	defer rows.Close()

	for rows.Next() {

		err := rows.Scan(&v.Id, &v.Fullname)

		if err != nil {
			return vs, models.ErrorWrap(err)
		}

		vs = append(vs, v)
	}

	err = rows.Err()

	if err != nil {
		return vs, models.ErrorWrap(err)
	}

	return vs, nil
}

func (d *dbGate) GetCustomers() ([]models.Customer, models.ApiError) {

	var (
		cs  []models.Customer
		c   models.Customer
		err error
	)

	qry := "SELECT id, fullname FROM customers"

	_, already := stmts[qry]

	if !already {

		stmts[qry], err = dbx.Prepare(qry)

		if err != nil {
			return cs, models.ErrorWrap(err)
		}
	}

	rows, err := stmts[qry].Query()

	defer rows.Close()

	for rows.Next() {

		err := rows.Scan(&c.Id, &c.Fullname)

		if err != nil {
			return cs, models.ErrorWrap(err)
		}

		cs = append(cs, c)
	}

	err = rows.Err()

	if err != nil {
		return cs, models.ErrorWrap(err)
	}

	return cs, nil
}
