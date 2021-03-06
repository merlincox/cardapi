# cardapi

This repo contains an implementation of the `aws-api-gateway-deploy` project which deploys an API simulating card 
transactions using a MySql database (not deployed).

The CloudFormation template

* Creates the lambda and passes an environment to it
* Creates the API Gateway with two sample endpoints linked to the lambda
* Creates a subdomain mapped to the API

Altogether, the deployment script enables you to create to `https://{your-sub-domain}.{your-domain}/status` endpoint,
 an example environment for the API (in the supplied example git tag, branch, and commit information and a 
'platform' variable). Redirection of HTTP to HTTPS and CloudWatch logging are automatically supplied by the gateway.

### Prerequisites

* A domain with a hosted-zone record in AWS Route 53
* A SSL certificate for that domain in the AWS Certificate Manager. Note that this certificate has to be in the 
`us-east-1` AWS region because it is deployed to CloudFront.
* A publically accessible MySQL database (this doesn't have to be hosted on AWS)
* The AWS command line interface `aws` installed and suitably set up with credentials for your AWS account
* `go` installed
* `glide` (a go dependency manager) installed
* `git` installed
* `jq` installed (`jq` is a very useful command-line tool for manipulating JSON. See https://stedolan.github.io/jq.)
* Optionally, the `generate` tool from https://github.com/a-h/generate to generate GoLang struct definitions from the API JSON schema
* Optionally, the `mysql` command-line tool installed to generate the table structure in the database

### Deployment

The deployment script usage is:
 
 `./deploy.sh subdomain_base domain [platform]`

`platform` is intended to be 'test', 'stage' or 'test' ('test' is the default)

'-test' and '-stage' are appended to the `subdomain_base`

Examples: 

`./deploy.sh my-api my-domain.com` will deploy an API at `https://my-api-test.my-domain.com`

`./deploy.sh my-api my-domain.com stage` will deploy an API at `https://my-api-stage.my-domain.com`

`./deploy.sh my-api my-domain.com live` will deploy an API at `https://my-api.my-domain.com`

Uncommited code cannot be deployed, and live deploys have these additional checks:

* code must be on the master branch
* code must be sync with the remote origin
* code must be exactly on a tag of the form 

Lastly, deploys to the live platform will present a confirmation prompt.

Note that the first time a stack is created there will be a significant delay before the subdomain is available due to 
propagation but subsequent updates should be quite fast.

### Exporting Swagger JSON and models

The API definition YAML includes a Swagger definition for the API.

This can be exported using the `export.sh` script.

In addition a schema-generator executable can be created from here: https://github.com/a-h/generate

If this is added to the system path, the `export.sh` will also generate Go structs for the API and optionally replace 
the models/api.go file if that is out of sync with the API. (Therefore any additional models which do not feature 
directly in the API should be placed in the models/models.go file, as well as any functions attached to API models).

### MySQL

MySQL connection details are defined in a `mysql.sh` script which is git-ignored. A `mysql.sh.example` file shows what needs to be defined.
`rebuild_db.sh` can be used to build the tables, sourcing connection details from the same file.

## The Card API

Note that for simplicity codes such as authorisation codes are merely autoincremental ids.

### Endpoints

| Endpoint  | Method | Body or Parameter | Description |
| ------------- | ------------- | ------------- | ------------- |
| `/status` | GET  | (none) | Returns status data about the API, including the platform deployed to and the Git branch, release and commit deployed from |
| `/customers` | GET | (none) | Returns the list of customers|
| `/vendors` | GET | (none) | Returns the list of vendors|
| `/card/{id}` | GET | id of the card | Returns data about a card identified by id, including movements such as top-ups, payments and refunds|
| `/authorisation/{id}` | GET | id of the authorisation | Returns data about a payment authorisation identified by id, including movements such as captures, reversals and refunds|
| `/customer/{id}` | GET | id of the customer | Returns data about customer by id, including cards held |
| `/vendor/{id}` | GET | id of the vendor | Returns data about a vendor identified by id, including authorisations|
| `/customer` | POST | customer object, with or without an id| Adds or updates a customer, which is returned |
| `/vendor` | POST | vendor object, with or without an id | Adds or updates a vendor, which is returned |
| `/card` | POST | customer object with an id | Adds a card to a customer. Returns the card. |
| `/authorise` | POST | Code request object with card id, vendor id, amount and description | Request to authorise a payment, returning an authorisation code |
| `/capture` | POST | Code request object with authorisation id and amount | Request to capture all or part of an authorised payment, returning a capture code |
| `/reverse` | POST | Code request object with authorisation id, amount and description | Request to reverse all or part of an authorised payment, returning a reversal code. Cannot be applied to captured payments. |
| `/refund` | POST | Code request object with authorisation id, amount and description | Request to refund all or part of an authorised and captured payment, returning a reversal code. Cannot be applied to uncaptured payments. |

### Models

Some of the endpoints require JSON-encoded models in the body of the POST.

For `/authorise`, `/top-up`, `/capture`, `/refund`, `/reverse` the model to use is a "code request" with these fields:

| Field  | Type | Required For | Description |
| ------------- | ------------- | ------------- | ------------- |
| `amount`    |  integer  |  (all) | Amount of payment etc in £0.01|
| `authorisationId` | integer | `/capture`, `/refund`, `/reverse` | Value returned from `/authorise` |
| `cardId`  |        integer  | `/authorise`, `/top-up` | Id of card
| `description` |    string | `/authorise`, `/top-up`, `/refund`, `/reverse` | Description of transaction |
| `vendorId` |       integer  | `/authorise` | Id of vendor |

The other models represent (highly simplified) vendors and customers.

For `/customer` and `/card` the customer model:

| Field  | Type | Notes |
| ------------- | ------------- | -------------
| `fullname`    |  string  |  |
| `id` | id | If present and non-zero, the POST `/customer` endpoint will attempt to update rather than create. Required for the POST `/card` endpoint. |

For `/vendor`, the vendor model:

| Field  | Type | Notes |
| ------------- | ------------- | -------------
| `vendorName`    |  string  |  |
| `id` | id | If present and non-zero, the POST `/vendor` endpoint will attempt to update rather than create. |







