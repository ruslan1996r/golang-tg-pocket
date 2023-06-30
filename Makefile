.PHONY:
.SILENT: # Этот параметр отвечает за отображение логов в консоли

build:
	go build -o ./.bin/bot cmd/bot/main.go # будет билдить проект в скрытую папку .bin

# Для вызова "make run" в терминале
run: build
	./.bin/bot

build-image:
	docker build -t tg-giga .

start-container:
	docker run --name telegram-bot -p 80:80 --env-file .env tg-giga