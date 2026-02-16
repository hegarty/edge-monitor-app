FROM golang:1.22-alpine AS build

ARG TARGETARCH=amd64

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -o gateway-monitor

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=build /src/gateway-monitor /gateway-monitor
EXPOSE 9093
ENTRYPOINT ["/gateway-monitor"]
