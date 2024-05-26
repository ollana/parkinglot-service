SHELL := /bin/bash

unit_test:
	cd handler && go test -race -timeout=120s -v ./...

build:
	cd handler && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap

package: build
	cd handler &&  zip bootstrap.zip bootstrap

deploy: package
	pulumi up --yes

destroy:
	pulumi destroy --yes
