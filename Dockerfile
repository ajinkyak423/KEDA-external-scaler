FROM golang:1.21.0 as builder

WORKDIR /src

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GO111MODULE=on go build -a -o external-scaler main.go


FROM alpine:latest

WORKDIR /

EXPOSE 6000

COPY --from=builder /src/external-scaler .

ENTRYPOINT ["/external-scaler"]