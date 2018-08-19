FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app/

COPY ./staging/builds/APP/job .

CMD ["./job"]