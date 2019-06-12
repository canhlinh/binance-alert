FROM golang:1.11-alpine as builder
RUN apk --no-cache --update add git gcc musl-dev linux-headers bash

RUN wget -q https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh -O /tmp/wait-for-db.sh
RUN chmod u+x /tmp/wait-for-db.sh

RUN wget -q https://github.com/markbates/refresh/releases/download/v1.4.11/refresh_1.4.11_linux_amd64.tar.gz \
    && tar -xzf refresh_1.4.11_linux_amd64.tar.gz && mv refresh /usr/local/bin/refresh && chmod u+x /usr/local/bin/refresh

WORKDIR /opt/alert
COPY . .
RUN go build -v -o ./bin/alert .

FROM alpine:3.8 as dist
RUN apk --no-cache --update add ca-certificates tzdata bash

WORKDIR /opt/alert
COPY --from=builder /opt/auth/bin/auth ./bin/alert

CMD ["./bin/alert"]