FROM golang:1.22.1

WORKDIR /build

COPY . ./

WORKDIR ./cmd/agent

RUN go build -o agent main.go

CMD ["./agent"]