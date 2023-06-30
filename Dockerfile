# Docker с двухэтапной сборкой
FROM golang:1.20-alpine3.18 AS builder

# Директория обязательно должна иметь имя модуля
COPY . /tg-giga/
WORKDIR /tg-giga/

# Аналог npm install
RUN go mod download
# Сбилдить main.go в папку bin/bot
RUN go build -o ./.bin/bot ./cmd/bot/main.go

FROM alpine:latest

WORKDIR /root/

# -from=0 - указывает на то, что нужно скопировать файлы из предыдущего этапа сборки
COPY --from=0 /tg-giga/.bin/bot .
COPY --from=0 /tg-giga/configs configs/

# Открыть порт 80
EXPOSE 80

# Вызвать cmd-файл
CMD ["./bot"]