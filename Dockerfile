FROM alpine:latest
ARG BIN
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY $BIN app

ENTRYPOINT ["./app"]