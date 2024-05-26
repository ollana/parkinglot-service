# Parking Lot Service - Cloud Computing Course Exercise 1

[Assignment details](ASSIGNMENT-README.md#assignment) | [Solution](#solution) | [Deployment steps](#deployment-steps) 


## Solution 

The service is implemented in [Golang](https://go.dev/), and [Pulumi](https://www.pulumi.com/) is used to deploy the service to AWS Lambda. 
I assume the service will have idle time when no cars are entering or exiting the parking lot, I also expect high traffic during peak hours when many cars are expected to enter and leave the parking lot at the same time.

I chose a serverless solution for the following reasons:

- Cost Efficiency: we pay only for the compute time consumes. If the parking lot service has periods of idle time, we will not incur costs for the idle server time.

- Scalability: AWS Lambda automatically scales the application, so the parking lot service can handle a few requests per day or scale up to thousands per second without any need for manual intervention.

- No Server Management: Lambda eliminates the need to manage servers. AWS handles all the operational aspects, including provisioning, maintaining, and scaling the infrastructure.


Since the service is stateless, DynamoDB is used to store the parking lot data. TODO: Implement DynamoDB


## Deployment steps
### Option 1 - Using makefile 
#### Prerequisites
- [make](https://www.incredibuild.com/integrations/gnu-make) 
- [golang 1.20+](https://go.dev/doc/install) 
- [pulumi 3.0.1+](https://www.pulumi.com/docs/install/)
- aws credentials setup to `pulumi` profile

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

### Option 2 - Using docker image 
#### Prerequisites
- [docker](https://docs.docker.com/get-docker/) 
- [pulumi 3.0.1+](https://www.pulumi.com/docs/install/)
- aws credentials setup to `pulumi` profile

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


### Open questions:
- is it ok to use makefile for deployment?
- is it ok to assume that grader has golang and make already set up? if not, I can provide an image building all necessary files that has to be run before pulumi deployment
- is it ok to assume that once exit is called we no longer store the ticket details?
- what is the unit of parked time? can it be returned as unix duration?
