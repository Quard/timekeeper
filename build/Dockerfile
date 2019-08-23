FROM golang:1.12-alpine as build
WORKDIR /app
COPY . .
RUN apk add git
ENV GO111MODULE on
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -v -o /timekeeper cmd/timekeeper/main.go

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=build /timekeeper /timekeeper
CMD /timekeeper
