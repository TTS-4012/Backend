FROM golang:1.21.3
WORKDIR /src

COPY go.mod .
RUN go mod download

COPY . .
RUN go build -o /bin/ocontest ./main.go
CMD ["/bin/ocontest"]