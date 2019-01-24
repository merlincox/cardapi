package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"github.com/merlincox/cardapi/models"
	"log"
)

const (
	DEFAULT_DATASOURCE_NAME = "merlincox:merlincox@tcp(www.db4free.net:3306)/merlincox"

	QUERY_GET_CUSTOMERS = "SELECT id, fullname FROM customers"
	QUERY_GET_CUSTOMER  = "SELECT id, fullname FROM customers WHERE id = ?"
	QUERY_GET_VENDORS   = "SELECT id, fullname FROM vendors"
	QUERY_GET_VENDOR    = "SELECT id, fullname FROM vendors WHERE id = ?"
	QUERY_GET_CARD      = "SELECT id, balance, available FROM cards WHERE id = ?"

	QUERY_ADD_VENDOR   = "INSERT INTO vendors (fullname) VALUES (?)"
	QUERY_ADD_CUSTOMER = "INSERT INTO customers (fullname) VALUES (?)"
	QUERY_ADD_CARD     = "INSERT INTO cards (customer_id) VALUES (?)"

	QUERY_AUTHORISE = `UPDATE cards SET available = available - ? WHERE id = ? AND available >= ?`

	QUERY_ADD_AUTHORISATION = `INSERT INTO authorisations (card_id, vendor_id, amount, description) 
                               VALUES (?, ?, ?, ?)`

	QUERY_TOPUP = `UPDATE cards SET balance = balance + ?, available = available + ? WHERE id = ?`

	QUERY_ADD_MOVEMENT = `INSERT INTO movements (card_id, amount, description, movement_date) 
                               VALUES (?, ?, ?, NOW())`

	ERROR_NOT_FOUND   = "sql: no rows in result set"
	ERROR_FOREIGN_KEY = "Error 1216: Cannot add or update a child row: a foreign key constraint fails"

	MESSAGE_GET_CUSTOMER_BAD_ID = "GetCustomer: no customer with id: %v"
	MESSAGE_GET_VENDOR_BAD_ID   = "GetVendor: no vendor with id: %v"
	MESSAGE_GET_CARD_BAD_ID     = "GetCard: no card with id: %v"
	MESSAGE_ADD_CARD_BAD_ID     = "AddCard: no customer with id: %v"

	MESSAGE_AUTHORISE_INVALID_AMOUNT         = "Authorise: amount must exceed 0"
	MESSAGE_AUTHORISE_INVALID_VENDOR         = "Authorise: no vendor with id: %v"
	MESSAGE_AUTHORISE_INVALID_CARD           = "Authorise: no card with id: %v"
	MESSAGE_AUTHORISE_INSUFFICIENT_AVAILABLE = "Authorise: insufficient funds: £%.2f exceeds available £%.2f"
)

type Dbi interface {
	AddCustomer(fullname string) (models.Customer, models.ApiError)
	GetCustomer(id int) (models.Customer, models.ApiError)
	GetCustomers() ([]models.Customer, models.ApiError)

	AddVendor(fullname string) (models.Vendor, models.ApiError)
	GetVendor(id int) (models.Vendor, models.ApiError)
	GetVendors() ([]models.Vendor, models.ApiError)

	AddCard(customerId int) (models.Card, models.ApiError)
	GetCard(id int) (models.Card, models.ApiError)

	//Refund(authorisationId, amount int, description string) (int, models.ApiError)
	//Capture(authorisationId, amount int) (int, models.ApiError)
	//TopUp(cardId, amount int, description string) (int, models.ApiError)
	Authorise(cardId, vendorId, amount int, description string) (int, models.ApiError)

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

	if dbx == nil {

		if injected == nil {

			s, err := sql.Open("mysql", DEFAULT_DATASOURCE_NAME)

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

	stmts = make(map[string]*sql.Stmt)
	dbx = nil
}

func (d *dbGate) GetVendors() ([]models.Vendor, models.ApiError) {

	var (
		vs  []models.Vendor
		v   models.Vendor
		err error
	)

	qry := QUERY_GET_VENDORS

	_, prepared := stmts[qry]

	if !prepared {

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

	qry := QUERY_GET_CUSTOMERS

	_, prepared := stmts[qry]

	if !prepared {

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

func (d *dbGate) GetCustomer(id int) (models.Customer, models.ApiError) {

	var (
		c   models.Customer
		err error
	)

	qry := QUERY_GET_CUSTOMER

	_, already := stmts[qry]

	if !already {

		stmts[qry], err = dbx.Prepare(qry)

		if err != nil {
			return c, models.ErrorWrap(err)
		}
	}

	err = stmts[qry].QueryRow(id).Scan(&c.Id, &c.Fullname)

	if err != nil {
		if err.Error() == ERROR_NOT_FOUND {
			return c, models.ConstructApiError(404, MESSAGE_GET_CUSTOMER_BAD_ID, id)
		}
		return c, models.ErrorWrap(err)
	}

	return c, nil
}

func (d *dbGate) GetVendor(id int) (models.Vendor, models.ApiError) {

	var (
		v   models.Vendor
		err error
	)

	qry := QUERY_GET_VENDOR

	_, already := stmts[qry]

	if !already {

		stmts[qry], err = dbx.Prepare(qry)

		if err != nil {
			return v, models.ErrorWrap(err)
		}
	}

	err = stmts[qry].QueryRow(id).Scan(&v.Id, &v.Fullname)

	if err != nil {
		if err.Error() == ERROR_NOT_FOUND {
			return v, models.ConstructApiError(404, MESSAGE_GET_VENDOR_BAD_ID, id)
		}
		return v, models.ErrorWrap(err)
	}

	return v, nil
}

func (d *dbGate) GetCard(id int) (models.Card, models.ApiError) {

	var (
		c   models.Card
		err error
	)

	qry := QUERY_GET_CARD

	_, already := stmts[qry]

	if !already {

		stmts[qry], err = dbx.Prepare(qry)

		if err != nil {
			return c, models.ErrorWrap(err)
		}
	}

	err = stmts[qry].QueryRow(id).Scan(&c.Id, &c.Balance, &c.Available)

	if err != nil {
		if err.Error() == ERROR_NOT_FOUND {
			return c, models.ConstructApiError(404, MESSAGE_GET_CARD_BAD_ID, id)
		}
		return c, models.ErrorWrap(err)
	}

	return c, nil
}

func (d *dbGate) AddVendor(fullname string) (models.Vendor, models.ApiError) {

	var (
		v   models.Vendor
		err error
	)

	qry := QUERY_ADD_VENDOR

	_, prepared := stmts[qry]

	if !prepared {

		stmts[qry], err = dbx.Prepare(qry)

		if err != nil {
			return v, models.ErrorWrap(err)
		}
	}

	res, err := stmts[qry].Exec(fullname)

	if err != nil {
		return v, models.ErrorWrap(err)
	}

	id64, err := res.LastInsertId()

	if err != nil {
		return v, models.ErrorWrap(err)
	}

	v = models.Vendor{
		Fullname: fullname,
		Id:       int(id64),
	}

	return v, nil
}

func (d *dbGate) AddCustomer(fullname string) (models.Customer, models.ApiError) {

	var (
		v   models.Customer
		err error
	)

	qry := QUERY_ADD_CUSTOMER

	_, prepared := stmts[qry]

	if !prepared {

		stmts[qry], err = dbx.Prepare(qry)

		if err != nil {
			return v, models.ErrorWrap(err)
		}
	}

	res, err := stmts[qry].Exec(fullname)

	if err != nil {
		return v, models.ErrorWrap(err)
	}

	id64, err := res.LastInsertId()

	if err != nil {
		return v, models.ErrorWrap(err)
	}

	v = models.Customer{
		Fullname: fullname,
		Id:       int(id64),
	}

	return v, nil
}

func (d *dbGate) AddCard(customerId int) (models.Card, models.ApiError) {

	var (
		v   models.Card
		err error
	)

	qry := QUERY_ADD_CARD

	_, prepared := stmts[qry]

	if !prepared {

		stmts[qry], err = dbx.Prepare(qry)

		if err != nil {
			return v, models.ErrorWrap(err)
		}
	}

	res, err := stmts[qry].Exec(customerId)

	if err != nil {
		if err.Error() == ERROR_FOREIGN_KEY {
			return v, models.ConstructApiError(400, MESSAGE_ADD_CARD_BAD_ID, customerId)
		}
		return v, models.ErrorWrap(err)
	}

	id64, err := res.LastInsertId()

	if err != nil {
		return v, models.ErrorWrap(err)
	}

	v = models.Card{
		CustomerId: customerId,
		Id:         int(id64),
	}

	return v, nil
}

func (d *dbGate) Authorise(cardId, vendorId, amount int, description string) (int, models.ApiError) {

	var (
		err error
	)

	if amount <= 0 {
		return -1, models.ConstructApiError(400, MESSAGE_AUTHORISE_INVALID_AMOUNT, amount/100)
	}

	_, apiErr := d.GetVendor(vendorId)

	if apiErr != nil {

		if apiErr.StatusCode() == 500 {
			return -1, apiErr
		}

		return -1, models.ConstructApiError(400, MESSAGE_AUTHORISE_INVALID_VENDOR, vendorId)
	}

	c, apiErr := d.GetCard(cardId)

	if apiErr != nil {

		if apiErr.StatusCode() == 500 {
			return -1, apiErr
		}

		return -1, models.ConstructApiError(400, MESSAGE_AUTHORISE_INVALID_CARD, cardId)
	}

	if c.Available < amount {
		return -1, models.ConstructApiError(400, MESSAGE_AUTHORISE_INSUFFICIENT_AVAILABLE, float32(amount)/100, float32(c.Available)/100)
	}

	tx, err := dbx.Begin()

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	defer tx.Rollback()

	qry := QUERY_AUTHORISE

	_, prepared := stmts[qry]

	if !prepared {

		stmts[qry], err = dbx.Prepare(qry)

		if err != nil {
			return -1, models.ErrorWrap(err)
		}
	}

	res, err := tx.Stmt(stmts[qry]).Exec(amount, cardId, amount)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	qry = QUERY_ADD_AUTHORISATION

	_, prepared = stmts[qry]

	if !prepared {

		stmts[qry], err = dbx.Prepare(qry)

		if err != nil {
			return -1, models.ErrorWrap(err)
		}
	}

	res, err = tx.Stmt(stmts[qry]).Exec(cardId, vendorId, amount, description)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	err = tx.Commit()

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	id64, err := res.LastInsertId()

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	return int(id64), nil
}

//func (d *dbGate) TopUp(cardId, amount int, description string) (models.Movement, models.ApiError) {
//
//	var (
//		m   models.Movement
//		err error
//	)
//
//	if amount <= 0 {
//		return m, models.ConstructApiError(400, MESSAGE_TOPUP_INVALID_AMOUNT, amount)
//	}
//
//	qry := QUERY_ADD_TOPUP
//
//	_, prepared := stmts[qry]
//
//	if !prepared {
//
//		stmts[qry], err = dbx.Prepare(qry)
//
//		if err != nil {
//			return m, models.ErrorWrap(err)
//		}
//	}
//
//	res, err := stmts[qry].Exec(amount, description, cardId, amount)
//
//	if err != nil {
//		if err.Error() == ERROR_FOREIGN_KEY {
//			return m, models.ConstructApiError(400, MESSAGE_ADD_CARD_BAD_ID, cardId)
//		}
//		return m, models.ErrorWrap(err)
//	}
//
//	id64, err := res.LastInsertId()
//
//	if err != nil {
//		return m, models.ErrorWrap(err)
//	}
//
//	m = models.Movement{
//		CardId: cardId,
//		MovementType: "Top-up",
//		Amount: amount,
//		Id:         int(id64),
//	}
//
//	return m, nil
//}
