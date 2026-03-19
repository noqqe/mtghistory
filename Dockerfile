# build go binary
FROM golang:1.26-alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR /go/src/app
COPY templates /go/src/app/templates
COPY assets /go/src/app/assets
COPY pkg /go/src/app/pkg
COPY cmd /go/src/app/cmd
COPY go.mod /go/src/app/go.mod
COPY go.sum /go/src/app/go.sum
COPY .git /go/src/app/.git
RUN go build -ldflags "-X github.com/noqqe/mtghistory/pkg/mtghistory.Version=`git describe --tags`" -v cmd/mtghistory/mtghistory.go

# copy
FROM scratch
WORKDIR /go/src/app
COPY --from=builder /go/src/app/ /go/src/app/
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /tmp /tmp
COPY templates /go/src/app/templates

# run
EXPOSE 8080
CMD [ "/go/src/app/mtghistory" ]

