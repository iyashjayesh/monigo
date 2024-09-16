FROM golang:latest as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o /monigo-app
FROM ubuntu:latest
RUN apt-get update && \
    apt-get install -y \
    graphviz \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /monigo-app /monigo-app
EXPOSE 8000 8080

ENTRYPOINT ["/monigo-app"]