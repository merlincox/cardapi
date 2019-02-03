// This is the API lambda executable
package main

import (
	"log"
	"os"
	"runtime"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/merlincox/cardapi/api/front"
	"github.com/merlincox/cardapi/db"
	"github.com/merlincox/cardapi/models"
)

const cacheTtlSeconds = 60

func main() {

	log.Printf("Starting %v API using Go %v\n", os.Getenv("RELEASE"), runtime.Version())
	log.Printf("Commit %v Timestamp %v\n", os.Getenv("COMMIT"), os.Getenv("TIMESTAMP"))

	status := models.Status{
		Platform:  os.Getenv("PLATFORM"),
		Commit:    os.Getenv("COMMIT"),
		Branch:    os.Getenv("BRANCH"),
		Release:   os.Getenv("RELEASE"),
		Timestamp: time.Now().Format(time.RFC3339Nano),
	}

	dbi, _ := db.NewDbi(nil)
	//@TODO handler db error

	lambda.Start(front.NewFront(dbi, status, cacheTtlSeconds).Handler)
}
