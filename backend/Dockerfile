FROM golang:1.22.4-alpine

WORKDIR /swallow

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

WORKDIR /swallow/cmd/swallow

RUN go build -o /swallow/bin/swallow .

EXPOSE 8080
ENTRYPOINT [ "/swallow/bin/swallow" ]