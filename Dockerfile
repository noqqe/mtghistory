FROM golang:1.22-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /go/src/app
COPY templates /go/src/app/templates
COPY assets /go/src/app/assets
COPY go.mod /go/src/app/go.mod
COPY go.sum /go/src/app/go.sum
COPY .git /go/src/app/.git
COPY main.go /go/src/app/main.go

# build
#RUN go get -v ./...
RUN go build -v .

# copy
FROM scratch
WORKDIR /go/src/app
COPY --from=builder /go/src/app/ /go/src/app/
COPY templates /go/src/app/templates
COPY assets /go/src/app/assets

# run
EXPOSE 8080
CMD [ "/go/src/app/mtghistory" ]

