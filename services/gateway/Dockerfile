FROM golang:1.20.6-alpine3.18 as build

WORKDIR /go/src/github.com/harlow/go-micro-services

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go install -ldflags="-s -w" .

FROM gcr.io/distroless/static-debian11:nonroot

COPY --from=build /go/bin/ /app/

ENTRYPOINT ["/app/gateway"]
