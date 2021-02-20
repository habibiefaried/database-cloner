FROM golang as builder
COPY . /app/
WORKDIR /app
RUN ls -lah && go build -ldflags "-linkmode external -extldflags -static" main.go

FROM ubuntu:focal
RUN apt update && apt install mysql-client postgresql-client-common postgresql-client mongodb-clients -y
COPY --from=builder /app/main /main
RUN chmod +x /main

ENTRYPOINT ["/main"]