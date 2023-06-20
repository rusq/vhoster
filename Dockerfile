FROM golang:1.20-alpine as stage

RUN apk add --no-cache make

WORKDIR /build
COPY . .

RUN make gateway

FROM alpine:3.17

RUN apk add --no-cache ca-certificates

COPY --from=stage /build/gateway /usr/local/bin/gateway

EXPOSE 8081

ENTRYPOINT [ "gateway" ]
