# Parking Lot Service - Cloud Computing Course Exercise 1

[Assignment details](ASSIGNMENT-README.md#assignment) | [Solution](#solution) | [Deployment steps](#deployment-steps) 


## Solution 

The service is implemented in [Golang](https://go.dev/), and [Pulumi](https://www.pulumi.com/) is used to deploy the service to AWS Lambda. 
I assume the service will have idle time when no cars are entering or exiting the parking lot, I also expect high traffic during peak hours when many cars are expected to enter and leave the parking lot at the same time.
Since the service is stateless, DynamoDB is used to store the parking lot data.

I chose a serverless solution for the following reasons:

- Cost Efficiency: we pay only for the compute time consumes. If the parking lot service has periods of idle time, we will not incur costs for the idle server time.

- Scalability: AWS Lambda automatically scales the application, so the parking lot service can handle a few requests per day or scale up to thousands per second without any need for manual intervention.

- No Server Management: Lambda eliminates the need to manage servers. AWS handles all the operational aspects, including provisioning, maintaining, and scaling the infrastructure.


Additional assumptions:
- The ticket if is a UUID. Example for response for `/entry`:
```json
{
    "ticketId": "1234"
}
```
- The exit ticket details have human-readable format for time and charge. Example for response for `/exit`: 
```json
{
    "licensePlate": "123-123-123",
    "parkedTime": "1h12m1s",
    "parkingLotId": "382",
    "charge": "$12.5"
}
```
- Charge is calculated based on 15 minutes increments rounded up.
- Ticket information is never deleted. After a car exits the parking lot, if exit is called again with the same ticket id, the service will return the same details as before.
- The service is not responsible for validating the license plate format, nor if the license plate is already in the parking lot.
- Possible race conditions are not handled. For example, if two requests are made with the same ticket id, the service will return the details of the first request.



## Deployment steps
Using makefile 
#### Prerequisites
- [make](https://www.incredibuild.com/integrations/gnu-make) 
- [golang 1.20+](https://go.dev/doc/install) 
- zip
- [pulumi 3.0.1+](https://www.pulumi.com/docs/install/)
- aws credentials setup to a profile named `pulumi`

#### Steps
1. login to pulumi
``` bash
pulumi login
```
2. deploy the service 
``` bash 
make deploy
```
3. destroy the service
``` bash
make destroy
```


### Alternative deployment option - using docker image 
This will build the necessary artifacts in a docker container and copy the artifacts to the host machine.

Then you can deploy the service using pulumi.

#### Prerequisites
- [docker](https://docs.docker.com/get-docker/) 
- [golang 1.20+](https://go.dev/doc/install)
- [pulumi 3.0.1+](https://www.pulumi.com/docs/install/)
- aws credentials setup to a profile named `pulumi` 

1. login to pulumi
``` bash
pulumi login
```
2. build and run the docker image to build necessary artifacts
``` bash
docker build -t parkinglot-service-builder .
docker create --name temp-container parkinglot-service-builder
docker cp temp-container:/app/bootstrap.zip .
docker rm temp-container
```
3. deploy the service 
``` bash
pulumi up
```
4. destroy the service
``` bash
pulumi destroy
```

### Service endpoint
The service url is outputted after the pulumi deployment.

For example:
```
Outputs:
    url: "https://xxxx.execute-api.us-west-2.amazonaws.com/stage/"
```
The service has the following endpoints:
- POST /entry
- POST /exit
as described in the [assignment details](ASSIGNMENT-README.md#assignment)