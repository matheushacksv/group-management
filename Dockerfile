FROM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/api .
# same goose version used in development, runs the migrations before the API starts
RUN CGO_ENABLED=0 GOBIN=/out go install github.com/pressly/goose/v3/cmd/goose@v3.29.0

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /out/api /out/goose /usr/local/bin/
COPY app/db/migrations /migrations

ENV PORT=8088
EXPOSE 8088

# ponytail: password goes inside the URL, avoid "@ : / #" in POSTGRES_PASSWORD
CMD ["sh", "-c", "goose -dir /migrations postgres \"postgresql://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB?sslmode=disable\" up && exec api"]
