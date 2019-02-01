package front

import (
	"testing"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/merlincox/cardapi/models"
	"github.com/merlincox/cardapi/utils"
	"github.com/merlincox/cardapi/mocks"
)

func makeFront(t *testing.T) Front {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	return NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "v1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)
}

func TestStatusRoute(t *testing.T) {

	testFront := makeFront(t)

	expected := models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "v1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}

	Convey("When sending an request with the /status route", t, func() {

		request := events.APIGatewayProxyRequest{
			RequestContext: events.APIGatewayProxyRequestContext{
				ResourcePath: `/status`,
				HTTPMethod:   `GET`,
			},
		}

		Convey("Then it should return the status", func() {
			response, err := testFront.Handler(request)
			So(response.Body, ShouldEqual, utils.JsonStringify(expected))
			So(response.Headers["Access-Control-Allow-Origin"], ShouldEqual, "*")
			So(response.Headers["Cache-Control"], ShouldEqual, "max-age=123")
			So(response.StatusCode, ShouldEqual, 200)
			So(err, ShouldBeNil)
		})
	})
}

func testCalc(t *testing.T, val1, val2 float64, locale, result, op, fullop string) {

	testFront := makeFront(t)

	expected := models.CalculationResult{
		Locale: locale,
		Op:     fullop,
		Result: result,
		Val1:   val1,
		Val2:   val2,
	}

	Convey(fmt.Sprintf("When sending an request with the /calc route with %v operator", fullop), t, func() {

		request := events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{
				"val1": fmt.Sprintf("%v", val1),
				"val2": fmt.Sprintf("%v", val2),
			},
			PathParameters: map[string]string{
				"op": op,
			},
			Headers: map[string]string{
				"Accept-Language": locale,
			},
			RequestContext: events.APIGatewayProxyRequestContext{
				ResourcePath: `/calc/{op}`,
				HTTPMethod:   `GET`,
			},
		}

		Convey("Then it should return the correct result", func() {
			response, err := testFront.Handler(request)

			// Do not differentiate non-breaking spaces from ordinary spaces for testing purposes
			body := strings.Replace(response.Body, "\u00A0", " ", -1)

			So(body, ShouldEqual, utils.JsonStringify(expected))
			So(response.Headers["Access-Control-Allow-Origin"], ShouldEqual, "*")
			So(response.Headers["Cache-Control"], ShouldEqual, "max-age=123")
			So(response.StatusCode, ShouldEqual, 200)
			So(err, ShouldBeNil)
		})
	})

}

func TestCalcRouteAddEn(t *testing.T) {
	testCalc(t, 3.5, 2.25, "en-GB", "5.75", "add", "add")
}

func TestCalcRouteAddFr(t *testing.T) {
	testCalc(t, 3.5, 2.25, "fr-FR", "5,75", "add", "add")
}

func TestCalcRouteSubEn(t *testing.T) {
	testCalc(t, 3.5, 2.25, "en-GB", "1.25", "sub", "subtract")
}

func TestCalcRouteSubFr(t *testing.T) {
	testCalc(t, 3.5, 2.25, "fr-FR", "1,25", "sub", "subtract")
}

func TestCalcRouteMultEn(t *testing.T) {
	testCalc(t, 1.5, 7000, "en-GB", "10,500", "mul", "multiply")
}

func TestCalcRouteMultFr(t *testing.T) {
	testCalc(t, 1.5, 7000, "fr-FR", "10 500", "mul", "multiply")
}

func TestCalcRoutePowEn(t *testing.T) {
	testCalc(t, 2, 3, "en-GB", "8", "pow", "power")
}

func TestCalcRouteRootEn(t *testing.T) {

	testCalc(t, 16, 2, "en-GB", "4", "roo", "root")
}

func testCalcRouteBad(t *testing.T, val1, val2 float64, op, context, msg string) {

	testFront := makeFront(t)

	expected := models.ApiErrorBody{
		Message: msg,
		Code:    400,
	}

	Convey(context, t, func() {

		request := events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{
				"val1": fmt.Sprintf("%v", val1),
				"val2": fmt.Sprintf("%v", val2),
			},
			PathParameters: map[string]string{
				"op": op,
			},
			Headers: map[string]string{
				"Accept-Language": "fr",
			},
			RequestContext: events.APIGatewayProxyRequestContext{
				ResourcePath: `/calc/{op}`,
				HTTPMethod:   `GET`,
			},
		}

		Convey("Then it should return the correct error", func() {
			response, err := testFront.Handler(request)
			So(response.Body, ShouldEqual, utils.JsonStringify(expected))
			So(response.Headers["Access-Control-Allow-Origin"], ShouldEqual, "*")
			So(response.Headers["Cache-Control"], ShouldEqual, "max-age=123")
			So(response.StatusCode, ShouldEqual, 400)
			So(err, ShouldBeNil)
		})
	})
}

func TestCalcRouteBadOp(t *testing.T) {
	testCalcRouteBad(t, 1, 2, "bad", "When sending a request to the /calc route with a bad operator", "Unknown calc operation: bad")
}

func TestCalcRouteInf(t *testing.T) {
	testCalcRouteBad(t, 1, 0, "div", "When sending a request to the /calc route with inf result", "Out of limits: 1 divide 0")
}

func TestCalcRouteNegInf(t *testing.T) {
	testCalcRouteBad(t, -1, 0, "div", "When sending a request to the /calc route with negative inf result", "Out of limits: -1 divide 0")
}

func TestCalcRouteNaN(t *testing.T) {
	testCalcRouteBad(t, -1, 2, "root", "When sending a request to the /calc route with NaN result", "Out of limits: -1 root 2")
}

func TestCustomersRoute(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "v1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/customers`,
			HTTPMethod:   `GET`,
		},
	}

	c1 := models.Customer{
		Fullname: "Fred Bloggs",
		Id:       1001,
	}

	c2 := models.Customer{
		Fullname: "Jane Doe",
		Id:       1002,
	}

	cs := []models.Customer{c1, c2}

	expected := models.CustomerList{
		Items:  cs,
		Offset: 0,
		Total:  len(cs),
	}

	mockDbi.EXPECT().GetCustomers().Return(cs, nil).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from GetCustomers", response.Body, utils.JsonStringify(expected))
	utils.AssertEquals(t, "Http code from GetCustomers", response.StatusCode, 200)
}

func TestGetCustomerRoute(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "v1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/customer/{id}`,
			HTTPMethod:   `GET`,
		},
		PathParameters: map[string]string{
			"id": "1001",
		},
	}

	expected := models.Customer{
		Fullname: "Fred Bloggs",
		Id:       1001,
	}

	mockDbi.EXPECT().GetCustomer(1001).Return(expected, nil).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from GetCustomer", utils.JsonStringify(expected), response.Body)
	utils.AssertEquals(t, "Http code from GetCustomer", 200, response.StatusCode)
}

func TestGetCustomerRoute404(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "v1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/customer/{id}`,
			HTTPMethod:   `GET`,
		},
		PathParameters: map[string]string{
			"id": "1001",
		},
	}

	expected := models.ConstructApiError(404, "GetCustomer: no customer with id: 1001")

	mockDbi.EXPECT().GetCustomer(1001).Return(models.Customer{}, expected).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from GetCustomer with invalid id", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from GetCustomer with invalid id", 404, response.StatusCode)
}

func TestGetCustomerRouteMalformedId(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "v1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/customer/{id}`,
			HTTPMethod:   `GET`,
		},
		PathParameters: map[string]string{
			"id": "badid",
		},
	}

	expected := models.ConstructApiError(400, "GetCustomer: malformed id: badid")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from GetCustomer", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from GetCustomer", 400, response.StatusCode)
}

func TestAddCustomerRoute(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "v1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.Customer{
		Fullname: "Joe Bloggs",
	}

	expected := models.Customer{
		Fullname: "Joe Bloggs",
		Id: 1001,
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/customer`,
			HTTPMethod:   `POST`,
		},
		Body:            utils.JsonStringify(body),
	}

	mockDbi.EXPECT().AddCustomer(body.Fullname).Return(expected, nil).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from AddCustomer", utils.JsonStringify(expected), response.Body)
	utils.AssertEquals(t, "Http code from GetCustomer", 200, response.StatusCode)
}

//@TODO write tests for other card API routes
