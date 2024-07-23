FROM golang:1.22.1

WORKDIR /build

COPY . ./

WORKDIR ./cmd/orchestrator

RUN go build -o orchestrator main.go

EXPOSE 8080
CMD ["./orchestrator"]