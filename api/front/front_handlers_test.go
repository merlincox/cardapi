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
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)
}

func TestStatusRoute(t *testing.T) {

	testFront := makeFront(t)

	expected := models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
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

// customer

func TestCustomersRoute(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
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
		Release:   "1.0.1",
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
		Release:   "1.0.1",
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
		Release:   "1.0.1",
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

func TestAddOrUpdateCustomerRouteAdd(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.Customer{
		Fullname: "Joe Bloggs",
	}

	expected := models.Customer{
		Fullname: "Joe Bloggs",
		Id:       1001,
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/customer`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	mockDbi.EXPECT().AddOrUpdateCustomer(body).Return(expected, nil).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from AddOrUpdateCustomer", utils.JsonStringify(expected), response.Body)
	utils.AssertEquals(t, "Http code from GetCustomer", 200, response.StatusCode)
}

//vendor

func TestVendorsRoute(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/vendors`,
			HTTPMethod:   `GET`,
		},
	}

	c1 := models.Vendor{
		VendorName: "Coffee shop",
		Id:         1001,
	}

	c2 := models.Vendor{
		VendorName: "Pub",
		Id:         1002,
	}

	cs := []models.Vendor{c1, c2}

	expected := models.VendorList{
		Items:  cs,
		Offset: 0,
		Total:  len(cs),
	}

	mockDbi.EXPECT().GetVendors().Return(cs, nil).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from GetVendors", response.Body, utils.JsonStringify(expected))
	utils.AssertEquals(t, "Http code from GetVendors", response.StatusCode, 200)
}

func TestGetVendorRoute(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/vendor/{id}`,
			HTTPMethod:   `GET`,
		},
		PathParameters: map[string]string{
			"id": "1001",
		},
	}

	expected := models.Vendor{
		VendorName: "Coffee shop",
		Id:         1001,
	}

	mockDbi.EXPECT().GetVendor(1001).Return(expected, nil).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from GetVendor", utils.JsonStringify(expected), response.Body)
	utils.AssertEquals(t, "Http code from GetVendor", 200, response.StatusCode)
}

func TestGetVendorRoute404(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/vendor/{id}`,
			HTTPMethod:   `GET`,
		},
		PathParameters: map[string]string{
			"id": "1001",
		},
	}

	expected := models.ConstructApiError(404, "GetVendor: no vendor with id: 1001")

	mockDbi.EXPECT().GetVendor(1001).Return(models.Vendor{}, expected).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from GetVendor with invalid id", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from GetVendor with invalid id", 404, response.StatusCode)
}

func TestGetVendorRouteMalformedId(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/vendor/{id}`,
			HTTPMethod:   `GET`,
		},
		PathParameters: map[string]string{
			"id": "badid",
		},
	}

	expected := models.ConstructApiError(400, "GetVendor: malformed id: badid")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from GetVendor", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from GetVendor", 400, response.StatusCode)
}

func TestAddOrUpdateVendorRoute(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.Vendor{
		VendorName: "Coffee shop",
	}

	expected := models.Vendor{
		VendorName: "Coffee shop",
		Id:         1001,
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/vendor`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	mockDbi.EXPECT().AddOrUpdateVendor(body).Return(expected, nil).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from AddOrUpdateVendor", utils.JsonStringify(expected), response.Body)
	utils.AssertEquals(t, "Http code from GetVendor", 200, response.StatusCode)
}

// authorisation

func TestGetAuthorisationRoute(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/authorisation/{id}`,
			HTTPMethod:   `GET`,
		},
		PathParameters: map[string]string{
			"id": "1001",
		},
	}

	expected := models.Authorisation{
		CardId:      100001,
		Id:          1001,
		Captured:    0,
		Refunded:    0,
		Reversed:    0,
		Amount:      210,
		Description: "Cake",
	}

	mockDbi.EXPECT().GetAuthorisation(1001).Return(expected, nil).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from GetAuthorisation", utils.JsonStringify(expected), response.Body)
	utils.AssertEquals(t, "Http code from GetAuthorisation", 200, response.StatusCode)
}

func TestGetAuthorisationRoute404(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/authorisation/{id}`,
			HTTPMethod:   `GET`,
		},
		PathParameters: map[string]string{
			"id": "1001",
		},
	}

	expected := models.ConstructApiError(404, "GetAuthorisation: no authorisation with id: 1001")

	mockDbi.EXPECT().GetAuthorisation(1001).Return(models.Authorisation{}, expected).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from GetAuthorisation with invalid id", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from GetAuthorisation with invalid id", 404, response.StatusCode)
}

func TestGetAuthorisationRouteMalformedId(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/authorisation/{id}`,
			HTTPMethod:   `GET`,
		},
		PathParameters: map[string]string{
			"id": "badid",
		},
	}

	expected := models.ConstructApiError(400, "GetAuthorisation: malformed id: badid")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from GetAuthorisation", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from GetAuthorisation", 400, response.StatusCode)
}

// card

func TestGetCardRoute(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/card/{id}`,
			HTTPMethod:   `GET`,
		},
		PathParameters: map[string]string{
			"id": "100001",
		},
	}

	expected := models.Card{
		Id: 100001,
	}

	mockDbi.EXPECT().GetCard(100001).Return(expected, nil).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from GetCard", utils.JsonStringify(expected), response.Body)
	utils.AssertEquals(t, "Http code from GetCard", 200, response.StatusCode)
}

func TestGetCardRoute404(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/card/{id}`,
			HTTPMethod:   `GET`,
		},
		PathParameters: map[string]string{
			"id": "1001",
		},
	}

	expected := models.ConstructApiError(404, "GetCard: no card with id: 1001")

	mockDbi.EXPECT().GetCard(1001).Return(models.Card{}, expected).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from GetCard with invalid id", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from GetCard with invalid id", 404, response.StatusCode)
}

func TestGetCardRouteMalformedId(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/card/{id}`,
			HTTPMethod:   `GET`,
		},
		PathParameters: map[string]string{
			"id": "badid",
		},
	}

	expected := models.ConstructApiError(400, "GetCard: malformed id: badid")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from GetCard", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from GetCard", 400, response.StatusCode)
}

func TestAddCardRoute(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.Customer{
		Fullname: "Fred Bloggs",
		Id:       1001,
	}

	expected := models.Card{
		Id:         100001,
		CustomerId: 1001,
		Balance:    0,
		Available:  0,
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/card`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	mockDbi.EXPECT().AddCard(body.Id).Return(expected, nil).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from AddCard", utils.JsonStringify(expected), response.Body)
	utils.AssertEquals(t, "Http code from GetCard", 200, response.StatusCode)
}

// code requests

func TestTopUpRoute(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		Amount:      20000,
		CardId:      100001,
		Description: "Top-up from bank",
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/top-up`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.CodeResponse{
		Id: 10009,
	}

	mockDbi.EXPECT().TopUp(body.CardId, body.Amount, body.Description).Return(expected.Id, nil).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from TopUp", utils.JsonStringify(expected), response.Body)
	utils.AssertEquals(t, "Http code from TopUp", 200, response.StatusCode)
}

func TestTopUpRouteBad1(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		Amount:      20000,
		Description: "Top-up from bank",
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/top-up`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.ConstructApiError(400, "Malformed top-up request: valid cardId, amount, description required")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from TopUp with incomplete code request data", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from TopUp with incomplete code request data", 400, response.StatusCode)
}

func TestTopUpRouteBad2(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		CardId:      10001,
		Description: "Top-up from bank",
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/top-up`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.ConstructApiError(400, "Malformed top-up request: valid cardId, amount, description required")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from TopUp with incomplete code request data", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from TopUp with incomplete code request data", 400, response.StatusCode)
}

func TestTopUpRouteBad3(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		CardId: 10001,
		Amount: 20000,
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/top-up`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.ConstructApiError(400, "Malformed top-up request: valid cardId, amount, description required")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from TopUp with incomplete code request data", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from TopUp with incomplete code request data", 400, response.StatusCode)
}

func TestAuthoriseRoute(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		Amount:      200,
		CardId:      100001,
		VendorId:    1001,
		Description: "Coffee",
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/authorise`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.CodeResponse{
		Id: 10009,
	}

	mockDbi.EXPECT().Authorise(body.CardId, body.VendorId, body.Amount, body.Description).Return(expected.Id, nil).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from Authorise", utils.JsonStringify(expected), response.Body)
	utils.AssertEquals(t, "Http code from Authorise", 200, response.StatusCode)
}

func TestAuthoriseRouteBad1(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		Amount:      200,
		VendorId:    1002,
		Description: "Cake",
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/authorise`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.ConstructApiError(400, "Malformed authorisation request: valid vendorId, cardId, amount, description required")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from Authorise with incomplete code request data", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from Authorise with incomplete code request data", 400, response.StatusCode)
}

func TestAuthoriseRouteBad2(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		Amount:      200,
		CardId:      100001,
		Description: "Cake",
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/authorise`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.ConstructApiError(400, "Malformed authorisation request: valid vendorId, cardId, amount, description required")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from Authorise with incomplete code request data", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from Authorise with incomplete code request data", 400, response.StatusCode)
}

func TestAuthoriseRouteBad3(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		VendorId:    1002,
		CardId:      100001,
		Description: "Cake",
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/authorise`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.ConstructApiError(400, "Malformed authorisation request: valid vendorId, cardId, amount, description required")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from Authorise with incomplete code request data", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from Authorise with incomplete code request data", 400, response.StatusCode)
}

func TestAuthoriseRouteBad4(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		Amount:   200,
		VendorId: 1002,
		CardId:   100001,
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/authorise`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.ConstructApiError(400, "Malformed authorisation request: valid vendorId, cardId, amount, description required")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from Authorise with incomplete code request data", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from Authorise with incomplete code request data", 400, response.StatusCode)
}

func TestCaptureRoute(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		Amount:          200,
		AuthorisationId: 1001,
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/capture`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.CodeResponse{
		Id: 10009,
	}

	mockDbi.EXPECT().Capture(body.AuthorisationId, body.Amount).Return(expected.Id, nil).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from Capture", utils.JsonStringify(expected), response.Body)
	utils.AssertEquals(t, "Http code from Capture", 200, response.StatusCode)
}

func TestCaptureRouteBad1(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		Amount: 20000,
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/capture`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.ConstructApiError(400, "Malformed capture request: valid authorisationId, amount required")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from Capture with incomplete code request data", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from Capture with incomplete code request data", 400, response.StatusCode)
}

func TestCaptureRouteBad2(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		AuthorisationId: 1002,
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/capture`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.ConstructApiError(400, "Malformed capture request: valid authorisationId, amount required")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from Capture with incomplete code request data", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from Capture with incomplete code request data", 400, response.StatusCode)
}

func TestRefundRoute(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		Amount:          200,
		AuthorisationId: 1001,
		Description:     "Bad coffee",
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/refund`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.CodeResponse{
		Id: 10009,
	}

	mockDbi.EXPECT().Refund(body.AuthorisationId, body.Amount, body.Description).Return(expected.Id, nil).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from Refund", utils.JsonStringify(expected), response.Body)
	utils.AssertEquals(t, "Http code from Refund", 200, response.StatusCode)
}

func TestRefundRouteBad1(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		AuthorisationId: 1001,
		Description:     "Bad coffee",
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/refund`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.ConstructApiError(400, "Malformed refund request: valid authorisationId, amount, description required")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from Refund with incomplete code request data", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from Refund with incomplete code request data", 400, response.StatusCode)
}

func TestRefundRouteBad2(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		Amount:          200,
		Description:     "Bad coffee",
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/refund`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.ConstructApiError(400, "Malformed refund request: valid authorisationId, amount, description required")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from Refund with incomplete code request data", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from Refund with incomplete code request data", 400, response.StatusCode)
}

func TestRefundRouteBad3(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		Amount:          200,
		AuthorisationId: 1001,
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/refund`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.ConstructApiError(400, "Malformed refund request: valid authorisationId, amount, description required")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from Refund with incomplete code request data", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from Refund with incomplete code request data", 400, response.StatusCode)
}

func TestReverseRoute(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		Amount:          200,
		AuthorisationId: 1001,
		Description:     "Bad coffee",
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/reverse`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.CodeResponse{
		Id: 10009,
	}

	mockDbi.EXPECT().Reverse(body.AuthorisationId, body.Amount, body.Description).Return(expected.Id, nil).Times(1)

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from Reverse", utils.JsonStringify(expected), response.Body)
	utils.AssertEquals(t, "Http code from Reverse", 200, response.StatusCode)
}

func TestReverseRouteBad1(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		AuthorisationId: 1001,
		Description:     "Bad coffee",
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/reverse`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.ConstructApiError(400, "Malformed reversal request: valid authorisationId, amount, description required")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from Reverse with incomplete code request data", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from Reverse with incomplete code request data", 400, response.StatusCode)
}

func TestReverseRouteBad2(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		Amount:          200,
		Description:     "Bad coffee",
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/reverse`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.ConstructApiError(400, "Malformed reversal request: valid authorisationId, amount, description required")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from Reverse with incomplete code request data", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from Reverse with incomplete code request data", 400, response.StatusCode)
}

func TestReverseRouteBad3(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDbi := mocks.NewMockDbi(mockCtrl)

	testFront := NewFront(mockDbi, models.Status{
		Branch:    "testing",
		Platform:  "test",
		Commit:    "a00eaaf45694163c9b728a7b5668e3d510eb3eb0",
		Release:   "1.0.1",
		Timestamp: "2019-01-02T14:52:36.951375973Z",
	}, 123)

	body := models.CodeRequest{
		Amount:          200,
		AuthorisationId: 1001,
	}

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: `/reverse`,
			HTTPMethod:   `POST`,
		},
		Body: utils.JsonStringify(body),
	}

	expected := models.ConstructApiError(400, "Malformed reversal request: valid authorisationId, amount, description required")

	response, _ := testFront.Handler(request)

	utils.AssertEquals(t, "Data from Reverse with incomplete code request data", utils.JsonStringify(expected.ErrorBody()), response.Body)
	utils.AssertEquals(t, "Http code from Reverse with incomplete code request data", 400, response.StatusCode)
}
