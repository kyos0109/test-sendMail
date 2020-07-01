FROM golang:1.14-stretch AS base

WORKDIR /go/src/app

COPY sender.go .

RUN go get -d -v ./...

RUN go install -v ./...

FROM gcr.io/distroless/base

COPY --from=base /go/bin/app /usr/bin/test-sendMail

WORKDIR /app

ENTRYPOINT ["test-sendMail"]
