package front

import (
	"strconv"
	"fmt"
	"math"
	"encoding/json"

	"golang.org/x/text/message"
	"golang.org/x/text/language"

	"github.com/aws/aws-lambda-go/events"
	
	"github.com/merlincox/cardapi/models"
)

func (front Front) statusHandler(request events.APIGatewayProxyRequest) (interface{}, models.ApiError) {

	return front.status, nil
}

type codeRequestHandler func(request models.CodeRequest) (int, models.ApiError)

func (front Front) addCustomerHandler(request events.APIGatewayProxyRequest) (interface{}, models.ApiError) {

	c := models.Customer{}

	err := json.Unmarshal([]byte(request.Body), &c)

	if err != nil {
		return nil, models.ErrorWrap(err)
	}

	return front.dbi.AddCustomer(c.Fullname)
}

func (front Front) addCardHandler(request events.APIGatewayProxyRequest) (interface{}, models.ApiError) {

	c := models.Customer{}

	err := json.Unmarshal([]byte(request.Body), &c)

	if err != nil {
		return nil, models.ErrorWrap(err)
	}

	return front.dbi.AddCard(c.Id)
}

func (front Front) addVendorHandler(request events.APIGatewayProxyRequest) (interface{}, models.ApiError) {

	c := models.Vendor{}

	err := json.Unmarshal([]byte(request.Body), &c)

	if err != nil {
		return nil, models.ErrorWrap(err)
	}

	return front.dbi.AddVendor(c.VendorName)
}

func (front Front) codeRequestHandler(request events.APIGatewayProxyRequest) (interface{}, models.ApiError) {

	cr := models.CodeRequest{}

	err := json.Unmarshal([]byte(request.Body), &cr)

	if err != nil {
		return nil, models.ErrorWrap(err)
	}

	var subHandler codeRequestHandler

	switch request.RequestContext.ResourcePath {

	case "/top-up":
		subHandler = front.authoriseHandler

	case "/authorise":
		subHandler = front.authoriseHandler

	case "/capture":
		subHandler = front.captureHandler

	case "/refund":
		subHandler = front.refundHandler

	case "/revert":
		subHandler = front.reversalHandler

	default:
		return nil, models.ConstructApiError(400, "Unsupported code request route: %v", request.RequestContext.ResourcePath)
	}

	id, apiErr := subHandler(cr)

	if apiErr != nil {
		return nil, apiErr
	}

	return models.CodeResponse{
		Id: id,
	}, nil
}

func (front Front) authoriseHandler(cr models.CodeRequest) (int, models.ApiError) {

	if cr.VendorId < 1 || cr.CardId < 1 || cr.Amount < 1 || cr.Description == "" {
		return -1, models.ConstructApiError(400, "Malformed authorisation request: valid vendorId, cardId, amount, description required")
	}

	return front.dbi.Authorise(cr.CardId, cr.VendorId, cr.Amount, cr.Description)
}

func (front Front) captureHandler(cr models.CodeRequest) (int, models.ApiError) {

	if cr.AuthorisationId < 1 || cr.Amount < 1 {
		return -1, models.ConstructApiError(400, "Malformed capture request: valid authorisationId, amount required")
	}

	return front.dbi.Capture(cr.AuthorisationId, cr.Amount)
}

func (front Front) refundHandler(cr models.CodeRequest) (int, models.ApiError) {

	if cr.AuthorisationId < 1 || cr.Amount < 1 || cr.Description == "" {
		return -1, models.ConstructApiError(400, "Malformed refund request: valid authorisationId, amount, description required")
	}

	return front.dbi.Refund(cr.AuthorisationId, cr.Amount, cr.Description)
}

func (front Front) reversalHandler(cr models.CodeRequest) (int, models.ApiError) {

	if cr.AuthorisationId < 1 || cr.Amount < 1 || cr.Description == "" {
		return -1, models.ConstructApiError(400, "Malformed reversal request: valid authorisationId, amount, description required")
	}

	return front.dbi.Refund(cr.AuthorisationId, cr.Amount, cr.Description)
}

func (front Front) topUpHandler(cr models.CodeRequest) (int, models.ApiError) {

	if cr.CardId < 1 || cr.Amount < 1 || cr.Description == "" {
		return -1, models.ConstructApiError(400, "Malformed top-up request: valid cardId, amount, description required")
	}

	return front.dbi.TopUp(cr.CardId, cr.Amount, cr.Description)
}

func (front Front) getCardHandler(request events.APIGatewayProxyRequest) (interface{}, models.ApiError) {

	ids := request.PathParameters["id"]

	id, err := strconv.ParseInt(ids, 0, 0)

	if err != nil {
		return nil, models.ConstructApiError(400, "GetCard: malformed id: %v", ids)

	}

	return front.dbi.GetCard(int(id))
}

func (front Front) getVendorHandler(request events.APIGatewayProxyRequest) (interface{}, models.ApiError) {

	ids := request.PathParameters["id"]

	id, err := strconv.ParseInt(ids, 0, 0)

	if err != nil {
		return nil, models.ConstructApiError(400, "GetVendor: malformed id: %v", ids)

	}

	return front.dbi.GetVendor(int(id))
}

func (front Front) getVendorsHandler(request events.APIGatewayProxyRequest) (interface{}, models.ApiError) {

	//for now, offset and limit are ignored

	vendors, err := front.dbi.GetVendors()

	if err != nil {
		return nil, err
	}

	return models.VendorList{
		Items:  vendors,
		Offset: 0,
		Total:  len(vendors),
	}, nil
}

func (front Front) getCustomersHandler(request events.APIGatewayProxyRequest) (interface{}, models.ApiError) {

	//for now, offset and limit are ignored

	customers, err := front.dbi.GetCustomers()

	if err != nil {
		return nil, err
	}

	return models.CustomerList{
		Items:  customers,
		Offset: 0,
		Total:  len(customers),
	}, nil
}

func (front Front) getCustomerHandler(request events.APIGatewayProxyRequest) (interface{}, models.ApiError) {

	ids := request.PathParameters["id"]

	id, err := strconv.ParseInt(ids, 0, 0)

	if err != nil {
		return nil, models.ConstructApiError(400, "GetCustomer: malformed id: %v", ids)

	}

	return front.dbi.GetCustomer(int(id))
}

func (front Front) getAuthorisationHandler(request events.APIGatewayProxyRequest) (interface{}, models.ApiError) {

	ids := request.PathParameters["id"]

	id, err := strconv.ParseInt(ids, 0, 0)

	if err != nil {
		return nil, models.ConstructApiError(400, "GetAuthorisation: malformed id: %v", ids)

	}

	return front.dbi.GetAuthorisation(int(id))
}

//left for backwards compatibility: please ignore

func (front Front) calcHandler(request events.APIGatewayProxyRequest) (interface{}, models.ApiError) {

	var (
		result float64
		fullop string
	)

	locale, ok := request.Headers["Accept-Language"]

	p := message.NewPrinter(language.Make(locale))

	if !ok {
		locale = "undefined"
	}

	op := request.PathParameters["op"]

	val1, err := getFloatFromRequest(request, "val1")

	if err != nil {
		return nil, models.ConstructApiError(400, err.Error())
	}

	val2, err := getFloatFromRequest(request, "val2")

	if err != nil {
		return nil, models.ConstructApiError(400, err.Error())
	}

	switch op[0:3] {

	case "add":

		result = val1 + val2
		fullop = "add"

	case "sub":

		result = val1 - val2
		fullop = "subtract"

	case "mul":

		result = val1 * val2
		fullop = "multiply"

	case "div":

		result = val1 / val2
		fullop = "divide"

	case "pow":

		result = math.Pow(val1, val2)
		fullop = "power"

	case "roo":

		result = math.Pow(val1, 1/val2)
		fullop = "root"

	default:

		return nil, models.ConstructApiError(400, "Unknown calc operation: %v", op)
	}

	if math.IsNaN(result) || math.IsInf(result, 1) || math.IsInf(result, -1) {
		return nil, models.ConstructApiError(400, "Out of limits: %v %v %v", val1, fullop, val2)
	}

	return models.CalculationResult{
		Locale: locale,
		Op:     fullop,
		Val1:   val1,
		Val2:   val2,
		Result: p.Sprintf("%v", result),
	}, nil
}

func getFloatFromRequest(request events.APIGatewayProxyRequest, key string) (result float64, err error) {

	val, ok := request.QueryStringParameters[key]

	if ! ok {
		err = fmt.Errorf("Missing parameter %v", key)
		return
	}

	result, err = strconv.ParseFloat(val, 64)

	return
}
