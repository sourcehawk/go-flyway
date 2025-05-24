ARG FLYWAY_VERSION=11.8.2

FROM golang:1.24.3 AS go-flyway

WORKDIR /build

COPY internal internal
COPY main.go .
COPY go.mod .
COPY go.sum .

RUN go mod download
RUN go build -o go-flyway main.go
RUN chmod +x go-flyway

FROM flyway/flyway:${FLYWAY_VERSION} AS runner

WORKDIR /app

COPY --from=go-flyway /build/go-flyway /app/go-flyway

ENTRYPOINT [ "/app/go-flyway" ]
# CMD [ "--config", "/path/to/config.yml" ]