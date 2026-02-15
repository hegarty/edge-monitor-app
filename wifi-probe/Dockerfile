FROM golang:1.22-alpine AS build

ARG TARGETARCH=amd64

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -o wifi-probe

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=build /src/wifi-probe /wifi-probe
EXPOSE 9090
ENTRYPOINT ["/wifi-probe"]
