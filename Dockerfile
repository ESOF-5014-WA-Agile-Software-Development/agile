FROM golang:1.23.5 AS build

WORKDIR /src/app

COPY . .

RUN go build -o /bin/app

FROM scratch
COPY --from=build /bin/app /

COPY config.prod.toml .

EXPOSE 3000

CMD ["/app", "-c", "config.prod.toml"]