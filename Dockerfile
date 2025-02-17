FROM golang:1.24 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -installsuffix 'static' ./cmd/processor

FROM gcr.io/distroless/static

WORKDIR /app

COPY --from=build /app/processor /app/

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT [ "/app/processor" ]
