FROM golang
WORKDIR /go/src/github.com/polyverse-security/hermes
COPY . .
WORKDIR /go/src/github.com/polyverse-security/hermes/standalone
RUN go get -v ./...
RUN GOOS=linux CGO_ENABLED=0 go build

FROM scratch
EXPOSE 9091
COPY --from=0 /go/src/github.com/polyverse-security/hermes/standalone/standalone /
ENTRYPOINT ["/standalone"]
