# BUILD STAGE
FROM golang:1.14 AS builder

ADD . /webhook-server
WORKDIR /webhook-server
RUN CGO_ENABLED=0 go build -o /output/server -v ./

# DEPLOY STAGE
FROM alpine:latest

ARG SLACK_WEBHOOK_LINK

ENV SLACK_WEBHOOK_LINK="${SLACK_WEBHOOK_LINK}"

COPY --from=builder /output/server /

RUN addgroup -S litmus && adduser -S -G litmus 1001 
USER 1001

CMD ["./server"]

EXPOSE 8081

