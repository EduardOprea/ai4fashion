FROM golang:1.17.5-alpine
ENV ENV_VAR_TEST="Valoare la misto"
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o main .
EXPOSE 8081
CMD ["/app/main"]