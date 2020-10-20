# Fetch the dependencies
FROM golang:1.15-alpine AS builder

RUN apk add --update ca-certificates git gcc g++ libc-dev
WORKDIR /src/

ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY site24x7/ /src/site24x7/
COPY main.go /src/

RUN CGO_ENABLED=0 GOOS=linux go build


# Build the final image
FROM hashicorp/terraform:0.12.29

COPY --from=builder /src/terraform-provider-site24x7 /root/.terraform.d/plugins/
