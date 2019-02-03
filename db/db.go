package db

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/go-sql-driver/mysql"

	"github.com/merlincox/cardapi/models"
)

const (
	QUERY_GET_CUSTOMERS = "SELECT id, fullname FROM customers"
	QUERY_GET_VENDORS   = "SELECT id, vendor_name, balance FROM vendors"

	QUERY_GET_VENDOR        = "SELECT id, vendor_name, balance FROM vendors WHERE id = ?"
	QUERY_GET_CARD          = "SELECT id, balance, available, ts FROM cards WHERE id = ?"
	QUERY_GET_AUTHORISATION = "SELECT id, amount, card_id, vendor_id, description, captured, reversed, refunded FROM authorisations WHERE id = ?"

	QUERY_GET_CARD_ALL = `SELECT c.id, c.balance, c.available, c.ts, m.id, m.amount, m.description, m.movement_type, m.ts
                            FROM cards c
                            LEFT OUTER JOIN movements m ON (m.card_id = c.id)
                            WHERE c.id = ?
                            ORDER BY m.ts`

	QUERY_GET_VENDOR_ALL = `SELECT v.id, v.vendor_name, v.balance, a.id, a.amount, a.card_id, a.description, a.captured, a.reversed, a.refunded, a.ts
                            FROM vendors v
                            LEFT OUTER JOIN authorisations a ON (a.vendor_id = v.id)
                            WHERE v.id = ?
                            ORDER BY a.ts`

	QUERY_GET_CUSTOMER_ALL = `SELECT cu.id, cu.fullname, c.id, c.balance, c.available, c.ts
                            FROM customers cu
                            LEFT OUTER JOIN cards c ON (c.customer_id = cu.id)
                            WHERE cu.id = ?
                            ORDER BY c.ts`

	QUERY_GET_AUTHORISATION_ALL = `SELECT a.id, a.amount, a.card_id, a.vendor_id, a.description, a.captured, a.reversed, a.refunded, m.id, m.amount, m.description, m.movement_type, m.ts
                            FROM authorisations a
                            LEFT OUTER JOIN auth_movements m ON (m.authorisation_id = a.id)
                            WHERE cu.id = ?
                            ORDER BY c.ts`

	QUERY_UPDATE_AUTH   = `UPDATE authorisations SET captured = captured + ?, refunded = refunded + ?, reversed = reversed + ? WHERE id = ?`
	QUERY_UPDATE_CARD   = `UPDATE cards SET balance = balance + ?, available = available + ? WHERE id = ?`
	QUERY_UPDATE_VENDOR = `UPDATE vendors SET balance = balance + ? WHERE id = ?`

	QUERY_UPDATE_VENDOR_DETAILS   = `UPDATE vendors SET vendor_name = ? WHERE id = ?`
	QUERY_UPDATE_CUSTOMER_DETAILS = `UPDATE customers SET fullname = ? WHERE id = ?`

	QUERY_ADD_VENDOR   = "INSERT INTO vendors (vendor_name) VALUES (?)"
	QUERY_ADD_CUSTOMER = "INSERT INTO customers (fullname) VALUES (?)"
	QUERY_ADD_CARD     = "INSERT INTO cards (customer_id) VALUES (?)"

	QUERY_ADD_AUTHORISATION = `INSERT INTO authorisations (card_id, vendor_id, amount, description) 
                               VALUES (?, ?, ?, ?)`

	QUERY_ADD_MOVEMENT = `INSERT INTO movements (card_id, amount, description, movement_type) 
                               VALUES (?, ?, ?, ?)`

	QUERY_ADD_AUTH_MOVEMENT = `INSERT INTO auth_movements (authorisation_id, amount, description, movement_type) 
                               VALUES (?, ?, ?, ?)`

	MYSQL_ERROR_FOREIGN_KEY = 1216

	MESSAGE_BAD_ID = "%v: no %v with id: %v"

	MESSAGE_INSUFFICIENT_AVAILABLE     = "%v: insufficient funds: £%.2f exceeds available £%.2f"
	MESSAGE_INSUFFICIENT_AVAILABLE_FOR = "%v: insufficient funds for amount £%.2f"

	MESSAGE_INVALID_ROW_UPDATE = "%v: invalid row update"
)

// Dbi interface for database operations
type Dbi interface {

	// GetCustomers returns an array of customers
	GetCustomers() ([]models.Customer, models.ApiError)
	// GetVendors returns an array of vendors
	GetVendors() ([]models.Vendor, models.ApiError)
	// Adds a customer using a customer object, or if an id already exists updates an existing customer

	// GetCustomer returns a customer object including associated cards
	GetCustomer(id int) (models.Customer, models.ApiError)
	// GetVendor returns a vendor object, including associated authorisations
	GetVendor(id int) (models.Vendor, models.ApiError)
	// GetCard returns a card object, including movements such as top-ups, payments, refunds
	GetCard(id int) (models.Card, models.ApiError)
	// GetAuthorisation returns an authorisation object, including associated movements such as captures, refunds, reversals etc
	GetAuthorisation(id int) (models.Authorisation, models.ApiError)

	AddOrUpdateCustomer(models.Customer) (models.Customer, models.ApiError)
	// AddOrUpdateVendor adds a vendor taking a vendor object, or if an id already exists updates an existing vendor
	AddOrUpdateVendor(models.Vendor) (models.Vendor, models.ApiError)
	// AddCard adds a card to a customer, taking a customer id and returning a card object
	AddCard(customerId int) (models.Card, models.ApiError)

	// TopUp simulates a top-up to a card and returns a top-up code
	TopUp(cardId, amount int, description string) (int, models.ApiError)
	// Authorise requests authorisation of a payment and returns an authorisation code
	Authorise(cardId, vendorId, amount int, description string) (int, models.ApiError)
	// Capture requests the capture of all or part of an authorised payment and returns a capture code
	Capture(authorisationId, amount int) (int, models.ApiError)
	// Refund requests a refund all or part of a captured payment and returns a refund code
	Refund(authorisationId, amount int, description string) (int, models.ApiError)
	// Reverse requests a reversal of all or part of a authorisation and returns a reversal code
	Reverse(authorisationId, amount int, description string) (int, models.ApiError)

	// Close closes prepared statements and the database connection
	Close()
}

type dbGate struct{}

var (
	dbd   *dbGate
	dbx   *sql.DB
	stmts = make(map[string]*sql.Stmt, 10)
	mutex sync.Mutex
)

// Returns a new Dbi singleton instance, retrieves the existing instance, or injects and returns one for testing
func NewDbi(mysqlDsn string, injected *sql.DB) (Dbi, models.ApiError) {

	mutex.Lock()
	defer mutex.Unlock()

	if dbx == nil {

		if injected == nil {

			s, err := sql.Open("mysql", mysqlDsn)

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

// Close close prepared statements and the database connection
func (d *dbGate) Close() {

	mutex.Lock()
	defer mutex.Unlock()

	for _, stmt := range stmts {
		stmt.Close()
	}

	if dbx != nil {
		dbx.Close()
	}

	stmts = make(map[string]*sql.Stmt)
	dbx = nil
}

// Retreive prepared query if it exists, or prepare the query and store it
func prepareQry(qry string) (err error) {

	mutex.Lock()
	defer mutex.Unlock()

	_, prepared := stmts[qry]

	if !prepared {

		stmts[qry], err = dbx.Prepare(qry)
	}

	return
}

type execResult struct {
	numRowsAffected int
	lastInsertedId  int
	apiErr          models.ApiError
	mysqlCode       uint16
}

// Reduce boiler plate by handling execution results in one place
func handleResults(result sql.Result, err error) execResult {

	if err == nil {

		affected, err := result.RowsAffected()

		if err == nil {

			lastId, err := result.LastInsertId()

			if err == nil {

				return execResult{
					numRowsAffected: int(affected),
					lastInsertedId:  int(lastId),
				}
			}
		}
	}

	res := execResult{
		apiErr: models.ErrorWrap(err),
	}

	if mysqlError, ok := err.(*mysql.MySQLError); ok {
		res.mysqlCode = mysqlError.Number
	}

	return res
}

// GetVendors returns an array of vendors
func (d *dbGate) GetVendors() ([]models.Vendor, models.ApiError) {

	var (
		vs  []models.Vendor
		v   models.Vendor
		err error
	)

	qry := QUERY_GET_VENDORS

	err = prepareQry(qry)

	if err != nil {
		return vs, models.ErrorWrap(err)
	}

	rows, err := stmts[qry].Query()

	if err != nil {
		return vs, models.ErrorWrap(err)
	}

	defer rows.Close()

	for rows.Next() {

		err := rows.Scan(&v.Id, &v.VendorName, &v.Balance)

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

// GetCustomers returns an array of customers
func (d *dbGate) GetCustomers() ([]models.Customer, models.ApiError) {

	var (
		cs  []models.Customer
		c   models.Customer
		err error
	)

	qry := QUERY_GET_CUSTOMERS

	err = prepareQry(qry)

	if err != nil {
		return cs, models.ErrorWrap(err)
	}

	rows, err := stmts[qry].Query()

	if err != nil {
		return cs, models.ErrorWrap(err)
	}

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

// GetCustomer returns a customer object including associated cards
func (d *dbGate) GetCustomer(id int) (models.Customer, models.ApiError) {

	var (
		cu  models.Customer
		c   models.NullableCard
		err error
	)

	qry := QUERY_GET_CUSTOMER_ALL

	err = prepareQry(qry)

	if err != nil {
		return cu, models.ErrorWrap(err)
	}

	rows, err := stmts[qry].Query(id)

	if err != nil {
		return cu, models.ErrorWrap(err)
	}

	defer rows.Close()

	for rows.Next() {

		//cu.id, cu.fullname, c.id, c.balance, c.available, c.ts
		err := rows.Scan(&cu.Id, &cu.Fullname, &c.Id, &c.Balance, &c.Available, &c.Ts)

		if err != nil {
			return cu, models.ErrorWrap(err)
		}

		if c.Valid() {
			c.CustomerId.Int64 = int64(id)
			cu.Cards = append(cu.Cards, c.Card())
		}
	}

	err = rows.Err()

	if err != nil {
		return cu, models.ErrorWrap(err)
	}

	if cu.Id == 0 {
		return cu, models.ConstructApiError(404, MESSAGE_BAD_ID, "GetCustomer", "customer", id)
	}

	return cu, nil
}

// GetVendor returns a vendor object, including associated authorisations
func (d *dbGate) GetVendor(id int) (models.Vendor, models.ApiError) {

	var (
		v   models.Vendor
		a   models.NullableAuthorisation
		err error
	)

	qry := QUERY_GET_VENDOR_ALL

	err = prepareQry(qry)

	if err != nil {
		return v, models.ErrorWrap(err)
	}

	rows, err := stmts[qry].Query(id)

	if err != nil {
		return v, models.ErrorWrap(err)
	}

	defer rows.Close()

	for rows.Next() {

		//v.id, v.vendor_name, v.balance, a.id, a.amount, a.card_id, a.description, a.captured, a.reversed, a.refunded, a.ts
		err := rows.Scan(&v.Id, &v.VendorName, &v.Balance, &a.Amount, &a.Id, &a.CardId, &a.Description, &a.Captured, &a.Reversed, &a.Refunded, &a.Ts)

		if err != nil {
			return v, models.ErrorWrap(err)
		}

		if a.Valid() {
			a.VendorId.Int64 = int64(id)
			v.Authorisations = append(v.Authorisations, a.Authorisation())
		}
	}

	err = rows.Err()

	if err != nil {
		return v, models.ErrorWrap(err)
	}

	if v.Id == 0 {
		return v, models.ConstructApiError(404, MESSAGE_BAD_ID, "GetVendor", "vendor", id)
	}

	return v, nil
}

func (d *dbGate) getVendor(id int) (models.Vendor, models.ApiError) {

	var (
		v   models.Vendor
		err error
	)

	qry := QUERY_GET_VENDOR

	err = prepareQry(qry)

	if err != nil {
		return v, models.ErrorWrap(err)
	}

	err = stmts[qry].QueryRow(id).Scan(&v.Id, &v.VendorName, &v.Balance)

	if err != nil {
		if err == sql.ErrNoRows {
			return v, models.ConstructApiError(404, MESSAGE_BAD_ID, "getVendor", "vendor", id)
		}
		return v, models.ErrorWrap(err)
	}

	return v, nil
}

func (d *dbGate) getAuthorisation(id int) (models.Authorisation, models.ApiError) {
	var (
		a   models.Authorisation
		err error
	)

	qry := QUERY_GET_AUTHORISATION

	err = prepareQry(qry)

	if err != nil {
		return a, models.ErrorWrap(err)
	}

	// id, amount, card_id, vendor_id, description, captured, reversed, refunded
	err = stmts[qry].QueryRow(id).Scan(&a.Id, &a.Amount, &a.CardId, &a.VendorId, &a.Description, &a.Captured, &a.Reversed, &a.Refunded)

	if err != nil {
		if err == sql.ErrNoRows {
			return a, models.ConstructApiError(404, MESSAGE_BAD_ID, "getAuthorisation", "authorisation", id)
		}
		return a, models.ErrorWrap(err)
	}

	return a, nil
}

// GetAuthorisation returns an authorisation object, including associated movements such as capture etc
func (d *dbGate) GetAuthorisation(id int) (models.Authorisation, models.ApiError) {
	var (
		a   models.Authorisation
		m   models.NullableMovement
		err error
	)

	qry := QUERY_GET_AUTHORISATION_ALL

	err = prepareQry(qry)

	if err != nil {
		return a, models.ErrorWrap(err)
	}

	rows, err := stmts[qry].Query(id)

	defer rows.Close()

	for rows.Next() {

		//a.id, a.amount, a.card_id, a.vendor_id, a.description, a.captured, a.reversed, a.refunded, m.amount, m.description, m.movement_type, m.ts
		err := rows.Scan(&a.Id, &a.Amount, &a.CardId, &a.VendorId, &a.Description, &a.Captured, &a.Reversed, &a.Refunded, &m.Id, &m.Amount, &m.Description, &m.MovementType, &m.Ts)

		if err != nil {
			return a, models.ErrorWrap(err)
		}

		if m.Valid() {
			m.ParentId.Int64 = int64(id)
			a.Movements = append(a.Movements, m.AuthMovement())
		}
	}

	err = rows.Err()

	if err != nil {
		return a, models.ErrorWrap(err)
	}

	if a.Id == 0 {
		return a, models.ConstructApiError(404, MESSAGE_BAD_ID, "GetAuthorisation", "authorisation", id)
	}

	return a, nil
}

func (d *dbGate) getCard(id int) (models.Card, models.ApiError) {

	var (
		c   models.Card
		err error
	)

	qry := QUERY_GET_CARD

	err = prepareQry(qry)

	if err != nil {
		return c, models.ErrorWrap(err)
	}

	err = stmts[qry].QueryRow(id).Scan(&c.Id, &c.Balance, &c.Available, &c.Ts)

	if err != nil {
		if err == sql.ErrNoRows {
			return c, models.ConstructApiError(404, MESSAGE_BAD_ID, "getCard", "card", id)
		}
		return c, models.ErrorWrap(err)
	}

	return c, nil
}

// GetCard  returns a card object, including movements such as top-ups, payments, refunds
func (d *dbGate) GetCard(id int) (models.Card, models.ApiError) {

	var (
		c   models.Card
		m   models.NullableMovement
		err error
	)

	qry := QUERY_GET_CARD_ALL

	err = prepareQry(qry)

	if err != nil {
		return c, models.ErrorWrap(err)
	}

	rows, err := stmts[qry].Query(id)

	defer rows.Close()

	for rows.Next() {

		//c.id, c.balance, c.available, c.ts, m.id, m.amount, m.description, m.movement_type, m.ts
		err := rows.Scan(&c.Id, &c.Balance, &c.Available, &c.Ts, &m.Id, &m.Amount, &m.Description, &m.MovementType, &m.Ts)

		if err != nil {
			return c, models.ErrorWrap(err)
		}

		if m.Valid() {
			m.ParentId.Int64 = int64(id)
			c.Movements = append(c.Movements, m.Movement())
		}
	}

	err = rows.Err()

	if err != nil {
		return c, models.ErrorWrap(err)
	}

	if c.Id == 0 {
		return c, models.ConstructApiError(404, MESSAGE_BAD_ID, "GetCard", "card", id)
	}

	return c, nil
}

// AddOrUpdateVendor adds a vendor taking a vendor object, or if an id already exists updates an existing vendor
func (d *dbGate) AddOrUpdateVendor(v models.Vendor) (models.Vendor, models.ApiError) {

	var (
		err error
	)

	qry := QUERY_ADD_VENDOR

	if v.Id > 0 {
		qry = QUERY_UPDATE_VENDOR_DETAILS
	}

	err = prepareQry(qry)

	if err != nil {
		return models.Vendor{}, models.ErrorWrap(err)
	}

	res := handleResults(stmts[qry].Exec(v.VendorName))

	if res.apiErr != nil {
		return models.Vendor{}, res.apiErr
	}

	if v.Id > 0 {

		if res.numRowsAffected == 0 {

			return models.Vendor{}, models.ConstructApiError(404, MESSAGE_BAD_ID, "AddOrUpdateVendor", "vendor", v.Id)
		}

	} else {

		v.Id = res.lastInsertedId
	}

	return v, nil
}

// AddOrUpdateCustomer adds a customer taking a customer object, or if an id already exists updates an existing customer
func (d *dbGate) AddOrUpdateCustomer(c models.Customer) (models.Customer, models.ApiError) {

	var (
		err error
	)

	qry := QUERY_ADD_CUSTOMER

	if c.Id > 0 {
		qry = QUERY_UPDATE_CUSTOMER_DETAILS
	}

	err = prepareQry(qry)

	if err != nil {
		return models.Customer{}, models.ErrorWrap(err)
	}

	res := handleResults(stmts[qry].Exec(c.Fullname))

	if res.apiErr != nil {
		return models.Customer{}, res.apiErr
	}

	if c.Id > 0 {
		if res.numRowsAffected == 0 {
			return models.Customer{}, models.ConstructApiError(404, MESSAGE_BAD_ID, "AddOrUpdateCustomer", "customer", c.Id)
		}
	} else {
		c.Id = res.lastInsertedId
	}

	return c, nil
}

// AddCard adds a card to a customer, taking a customer id and returning a card object
func (d *dbGate) AddCard(customerId int) (models.Card, models.ApiError) {

	var (
		c   models.Card
		err error
	)

	qry := QUERY_ADD_CARD

	err = prepareQry(qry)

	if err != nil {
		return c, models.ErrorWrap(err)
	}

	res := handleResults(stmts[qry].Exec(customerId))

	if res.apiErr != nil {

		if res.mysqlCode == MYSQL_ERROR_FOREIGN_KEY {
			return c, models.ConstructApiError(400, MESSAGE_BAD_ID, "AddCard", "customer", customerId)
		}

		return c, res.apiErr
	}

	c = models.Card{
		CustomerId: customerId,
		Id:         res.lastInsertedId,
	}

	return c, nil
}

// Authorise requests authorisation of a payment and returns an authorisation code
func (d *dbGate) Authorise(cardId, vendorId, amount int, description string) (int, models.ApiError) {

	_, apiErr := d.getVendor(vendorId)

	if apiErr != nil {

		if apiErr.StatusCode() == 500 {
			return -1, apiErr
		}

		return -1, models.ConstructApiError(400, MESSAGE_BAD_ID, "Authorise", "vendor", vendorId)
	}

	c, apiErr := d.getCard(cardId)

	if apiErr != nil {

		if apiErr.StatusCode() == 500 {
			return -1, apiErr
		}

		return -1, models.ConstructApiError(400, MESSAGE_BAD_ID, "Authorise", "card", cardId)
	}

	if c.Available < amount {
		return -1, models.ConstructApiError(400, MESSAGE_INSUFFICIENT_AVAILABLE, "Authorise", float32(amount)/100, float32(c.Available)/100)
	}

	tx, err := dbx.Begin()

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	defer tx.Rollback()

	qry := QUERY_UPDATE_CARD

	err = prepareQry(qry)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	res := handleResults(tx.Stmt(stmts[qry]).Exec(0, amount, cardId, amount))

	if err != nil {
		return -1, res.apiErr
	}

	//double check that exactly one row was updated

	if res.numRowsAffected != 1 {
		return -1, models.ConstructApiError(400, MESSAGE_INSUFFICIENT_AVAILABLE_FOR, "Authorise", float32(amount)/100)
	}

	qry = QUERY_ADD_AUTHORISATION

	err = prepareQry(qry)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	res = handleResults(tx.Stmt(stmts[qry]).Exec(cardId, vendorId, amount, description))

	if err != nil {
		return -1, res.apiErr
	}

	err = tx.Commit()

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	return res.lastInsertedId, nil
}

// TopUp simulates a top-up to a card and returns a top-up code
func (d *dbGate) TopUp(cardId, amount int, description string) (int, models.ApiError) {

	_, apiErr := d.getCard(cardId)

	if apiErr != nil {

		if apiErr.StatusCode() == 500 {
			return -1, apiErr
		}

		return -1, models.ConstructApiError(400, MESSAGE_BAD_ID, "TopUp", "card", cardId)
	}

	tx, err := dbx.Begin()

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	defer tx.Rollback()

	qry := QUERY_UPDATE_CARD

	err = prepareQry(qry)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	res := handleResults(tx.Stmt(stmts[qry]).Exec(amount, amount, cardId))

	if err != nil {
		return -1, res.apiErr
	}

	//double check that exactly one row was updated

	if res.numRowsAffected != 1 {
		return -1, models.ConstructApiError(500, MESSAGE_INVALID_ROW_UPDATE, "TopUp")
	}

	qry = QUERY_ADD_MOVEMENT

	err = prepareQry(qry)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	res = handleResults(tx.Stmt(stmts[qry]).Exec(cardId, amount, description, "TOP-UP"))

	if res.apiErr != nil {
		return -1, res.apiErr
	}

	err = tx.Commit()

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	return res.lastInsertedId, nil
}

// Capture requests the capture of all or part of an authorised payment and returns a capture code
func (d *dbGate) Capture(authorisationId, amount int) (int, models.ApiError) {

	auth, apiErr := d.getAuthorisation(authorisationId)

	if apiErr != nil {

		if apiErr.StatusCode() == 500 {
			return -1, apiErr
		}

		return -1, models.ConstructApiError(400, MESSAGE_BAD_ID, "Capture", "authorisation", authorisationId)
	}

	if amount > auth.Capturable() {
		return -1, models.ConstructApiError(400, MESSAGE_INSUFFICIENT_AVAILABLE, "Capture", float32(amount)/100, float32(auth.Capturable())/100)
	}

	tx, err := dbx.Begin()

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	defer tx.Rollback()

	qry := QUERY_UPDATE_CARD

	err = prepareQry(qry)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	res := handleResults(tx.Stmt(stmts[qry]).Exec(-amount, 0, auth.CardId))

	if res.apiErr != nil {
		return -1, res.apiErr
	}

	//double check that exactly one row was updated

	if res.numRowsAffected != 1 {
		return -1, models.ConstructApiError(500, MESSAGE_INVALID_ROW_UPDATE, "Capture")
	}

	qry = QUERY_UPDATE_AUTH

	err = prepareQry(qry)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	//captured = captured + ?, refunded = refunded + ?, reversed = reversed + ?
	res = handleResults(tx.Stmt(stmts[qry]).Exec(amount, 0, 0, auth.Id))

	if res.apiErr != nil {
		return -1, res.apiErr
	}

	//double check that exactly one row was updated

	if res.numRowsAffected != 1 {
		return -1, models.ConstructApiError(500, MESSAGE_INVALID_ROW_UPDATE, "Capture")
	}

	// this is only done in this simulation so that the effect of capturing is easily visible through a UI
	qry = QUERY_UPDATE_VENDOR

	err = prepareQry(qry)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	res = handleResults(tx.Stmt(stmts[qry]).Exec(amount, auth.VendorId))

	if res.apiErr != nil {
		return -1, res.apiErr
	}

	//double check that exactly one row was updated

	if res.numRowsAffected != 1 {
		return -1, models.ConstructApiError(500, MESSAGE_INVALID_ROW_UPDATE, "Capture")
	}

	qry = QUERY_ADD_MOVEMENT

	err = prepareQry(qry)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	res = handleResults(tx.Stmt(stmts[qry]).Exec(auth.CardId, -amount, auth.Description, "PURCHASE")) //? add original purchase date from auth.Ts

	if res.apiErr != nil {
		return -1, res.apiErr
	}

	qry = QUERY_ADD_AUTH_MOVEMENT

	err = prepareQry(qry)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	res = handleResults(tx.Stmt(stmts[qry]).Exec(auth.Id, amount, fmt.Sprintf("Capture of £%.2f", float32(amount)/100), "CAPTURE"))

	if res.apiErr != nil {
		return -1, res.apiErr
	}

	err = tx.Commit()

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	return res.lastInsertedId, nil
}

// Refund requests a refund all or part of a captured payment and returns a refund code
func (d *dbGate) Refund(authorisationId, amount int, description string) (int, models.ApiError) {

	auth, apiErr := d.getAuthorisation(authorisationId)

	if apiErr != nil {

		if apiErr.StatusCode() == 500 {
			return -1, apiErr
		}

		return -1, models.ConstructApiError(400, MESSAGE_BAD_ID, "Refund", "authorisation", authorisationId)
	}

	if amount > auth.Refundable() {
		return -1, models.ConstructApiError(400, MESSAGE_INSUFFICIENT_AVAILABLE, "Refund", float32(amount)/100, float32(auth.Refundable())/100)
	}

	tx, err := dbx.Begin()

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	defer tx.Rollback()

	qry := QUERY_UPDATE_CARD

	err = prepareQry(qry)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	res := handleResults(tx.Stmt(stmts[qry]).Exec(amount, amount, auth.CardId))

	if res.apiErr != nil {
		return -1, res.apiErr
	}

	//double check that exactly one row was updated

	if res.numRowsAffected != 1 {
		return -1, models.ConstructApiError(500, MESSAGE_INVALID_ROW_UPDATE, "Refund")
	}

	qry = QUERY_UPDATE_AUTH

	err = prepareQry(qry)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	//captured = captured + ?, refunded = refunded + ?, reversed = reversed + ?
	res = handleResults(tx.Stmt(stmts[qry]).Exec(0, amount, 0, auth.Id))

	if res.apiErr != nil {
		return -1, res.apiErr
	}

	//double check that exactly one row was updated

	if res.numRowsAffected != 1 {
		return -1, models.ConstructApiError(500, MESSAGE_INVALID_ROW_UPDATE, "Capture")
	}

	// this is only done in this simulation so that the effect of capturing is easily visible through a UI
	qry = QUERY_UPDATE_VENDOR

	err = prepareQry(qry)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	res = handleResults(tx.Stmt(stmts[qry]).Exec(-amount, auth.VendorId))

	if res.apiErr != nil {
		return -1, res.apiErr
	}

	//double check that exactly one row was updated

	if res.numRowsAffected != 1 {
		return -1, models.ConstructApiError(500, MESSAGE_INVALID_ROW_UPDATE, "Refund")
	}

	qry = QUERY_ADD_MOVEMENT

	err = prepareQry(qry)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	res = handleResults(tx.Stmt(stmts[qry]).Exec(auth.CardId, amount, description, "REFUND"))

	if res.apiErr != nil {
		return -1, res.apiErr
	}

	qry = QUERY_ADD_AUTH_MOVEMENT

	err = prepareQry(qry)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	res = handleResults(tx.Stmt(stmts[qry]).Exec(auth.Id, -amount, description, "REFUND"))

	if res.apiErr != nil {
		return -1, res.apiErr
	}

	err = tx.Commit()

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	return res.lastInsertedId, nil
}

// Reverse requests a reversal of all or part of a authorisation and returns a reversal code
func (d *dbGate) Reverse(authorisationId, amount int, description string) (int, models.ApiError) {

	auth, apiErr := d.getAuthorisation(authorisationId)

	if apiErr != nil {

		if apiErr.StatusCode() == 500 {
			return -1, apiErr
		}

		return -1, models.ConstructApiError(400, MESSAGE_BAD_ID, "Reverse", "authorisation", authorisationId)
	}

	if amount > auth.Capturable() {
		return -1, models.ConstructApiError(400, MESSAGE_INSUFFICIENT_AVAILABLE, "Reverse", float32(amount)/100, float32(auth.Capturable())/100)
	}

	tx, err := dbx.Begin()

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	defer tx.Rollback()

	qry := QUERY_UPDATE_CARD

	err = prepareQry(qry)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	res := handleResults(tx.Stmt(stmts[qry]).Exec(0, amount, auth.CardId))

	if res.apiErr != nil {
		return -1, res.apiErr
	}

	//double check that exactly one row was updated

	if res.numRowsAffected != 1 {
		return -1, models.ConstructApiError(500, MESSAGE_INVALID_ROW_UPDATE, "Reverse")
	}

	qry = QUERY_UPDATE_AUTH

	err = prepareQry(qry)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	//captured = captured + ?, refunded = refunded + ?, reversed = reversed + ?
	res = handleResults(tx.Stmt(stmts[qry]).Exec(0, 0, amount, auth.Id))

	if res.apiErr != nil {
		return -1, res.apiErr
	}

	//double check that exactly one row was updated

	if res.numRowsAffected != 1 {
		return -1, models.ConstructApiError(500, MESSAGE_INVALID_ROW_UPDATE, "Reverse")
	}

	qry = QUERY_ADD_AUTH_MOVEMENT

	err = prepareQry(qry)

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	res = handleResults(tx.Stmt(stmts[qry]).Exec(auth.Id, -amount, description, "REVERSAL"))

	if res.apiErr != nil {
		return -1, res.apiErr
	}

	err = tx.Commit()

	if err != nil {
		return -1, models.ErrorWrap(err)
	}

	return res.lastInsertedId, nil
}
