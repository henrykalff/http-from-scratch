FROM golang:1.27 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/httpserver ./cmd/httpserver/main.go

FROM scratch 

COPY --from=build /bin/httpserver /bin/httpserver
EXPOSE 8080
CMD ["/bin/httpserver"]