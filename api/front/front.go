// The Front package routes HTTP requests to an appropriately routed handler and returns a response
// whose Body member is a JSON-encoded API object. In case of error it will be a JSON-encoded ApiErrorBody.
package front

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"

	"github.com/merlincox/cardapi/db"
	"github.com/merlincox/cardapi/models"
	"github.com/merlincox/cardapi/utils"
)

type Front struct {
	dbi         db.Dbi
	status      models.Status
	router      func(route string) innerHandler
	cacheMaxAge int
}

type innerHandler func(request events.APIGatewayProxyRequest) (interface{}, models.ApiError)

// NewFront creates a new Front object
func NewFront(dbi db.Dbi, status models.Status, cacheMaxAge int) Front {

	f := Front{
		dbi:         dbi,
		status:      status,
		cacheMaxAge: cacheMaxAge,
	}

	f.router = f.getHandlerForRoute

	return f
}

// Front.Handler takes an APIGatewayProxyRequest and returns an APIGatewayProxyResponse with an error which should be nil
//
// Any downstream panic should be recovered and wrapped into an ApiErrorBody, and the trace logged
func (front Front) Handler(request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {

	useCache := request.RequestContext.HTTPMethod == "GET"

	defer func() {

		if r := recover(); r != nil {
			log.Println(utils.JsonStack(r, debug.Stack()))
			apiErr := models.ConstructApiError(http.StatusInternalServerError, "Panic: %v", r)
			response = front.buildResponse(nil, apiErr, useCache)
		}

	}()

	route := getRoute(request)
	log.Println("Handling a request for %v.", route)

	data, apiErr := front.router(route)(request)
	response = front.buildResponse(data, apiErr, useCache)

	return
}

func (front *Front) getHandlerForRoute(route string) innerHandler {

	switch route {

	case "GET/status":
		return front.statusHandler

	case "GET/calc/{op}":
		return front.calcHandler

	case "POST/authorise",
		"POST/refund",
		"POST/reverse",
		"POST/top-up",
		"POST/capture":
		return front.codeRequestHandler

	case "POST/card":
		return front.addCardHandler

	case "POST/customer":
		return front.addCustomerHandler

	case "POST/vendor":
		return front.addVendorHandler

	case "GET/card/{id}":
		return front.getCardHandler

	case "GET/vendor/{id}":
		return front.getVendorHandler

	case "GET/customer/{id}":
		return front.getCustomerHandler

	case "GET/authorisation/{id}":
		return front.getAuthorisationHandler

	case "GET/vendors":
		return front.getVendorsHandler

	case "GET/customers":
		return front.getCustomersHandler
	}

	return front.unknownRouteHandler
}

func getRoute(request events.APIGatewayProxyRequest) string {

	return request.RequestContext.HTTPMethod + request.RequestContext.ResourcePath
}

func (front Front) unknownRouteHandler(request events.APIGatewayProxyRequest) (interface{}, models.ApiError) {

	return nil, models.ConstructApiError(http.StatusNotFound, "No such route as %v", getRoute(request))
}

func (front *Front) buildResponse(data interface{}, err models.ApiError, useCache bool) events.APIGatewayProxyResponse {

	var (
		body       string
		statusCode int
	)

	if err != nil {

		body = utils.JsonStringify(err.ErrorBody())
		statusCode = err.StatusCode()
		log.Printf("ERROR: Returning %v: %v", statusCode, err.Error())

	} else {

		body = utils.JsonStringify(data)
		statusCode = http.StatusOK
	}

	// handle unlikely case where json.Marshall fails for the data argument
	if body == "" {
		statusCode = http.StatusInternalServerError
		body = fmt.Sprintf(`{"message":"Unmarshallable data","code":%v}`, statusCode)
		log.Printf("ERROR: Returning %v: %v", statusCode, "Unmarshallable data")
	}

	cacheValue := "no-cache"

	if useCache {
		cacheValue = "max-age=" + strconv.Itoa(front.cacheMaxAge)
	}

	return events.APIGatewayProxyResponse{
		Body:       body,
		StatusCode: statusCode,
		Headers: map[string]string{
			"Cache-Control":               cacheValue,
			"Access-Control-Allow-Origin": "*",
			"X-Timestamp":                 time.Now().UTC().Format(time.RFC3339Nano),
		}}
}
