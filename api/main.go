// This is the API lambda executable
package main

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/merlincox/cardapi/api/front"
	"github.com/merlincox/cardapi/db"
	"github.com/merlincox/cardapi/models"
	"github.com/merlincox/cardapi/utils"
)

const cacheTtlSeconds = 60

var handler func(request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error)

func main() {

	log.Printf("Starting %v API using Go %v\n", os.Getenv("RELEASE"), runtime.Version())
	log.Printf("Commit %v Timestamp %v\n", os.Getenv("COMMIT"), os.Getenv("TIMESTAMP"))

	dbi, apiErr := db.NewDbi(os.Getenv("MYSQLDSN"), nil)

	if apiErr != nil {

		log.Printf("Fatal database error: %v", apiErr.Error())

		handler = func(request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {

			body := apiErr.ErrorBody()

			response = events.APIGatewayProxyResponse{
				StatusCode: http.StatusServiceUnavailable,
				Body:       utils.JsonStringify(body),
			}

			return
		}

	} else {

		status := models.Status{
			Platform:  os.Getenv("PLATFORM"),
			Commit:    os.Getenv("COMMIT"),
			Branch:    os.Getenv("BRANCH"),
			Release:   os.Getenv("RELEASE"),
			Timestamp: time.Now().Format(time.RFC3339Nano),
		}

		handler = front.NewFront(dbi, status, cacheTtlSeconds).Handler
	}

	lambda.Start(handler)
}