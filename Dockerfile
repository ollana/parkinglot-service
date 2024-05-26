# Start from the latest golang base image
FROM golang:1.20 as builder

# Set the current working directory inside the container
WORKDIR /app

COPY handler/go.mod handler/go.sum ./

RUN go mod download

COPY handler/. .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./main.go

FROM public.ecr.aws/lambda/provided:al2
COPY --from=builder /app/main /var/task/

CMD [ "main" ]
