# WebSocket-Go

Небольшой учебный проект на Go для практики с WebSocket.

Здесь реализован простой чат:
- Go HTTP-сервер раздаёт HTML-страницу
- клиент подключается по WebSocket
- сообщения рассылаются всем подключённым клиентам
- сервер добавляет к сообщению IP-адрес отправителя и время

## Стек

- Go
- Gorilla WebSocket
- Zap logger
- HTML + Bootstrap

## Что есть в проекте

- [cmd/server/main.go](d:\prog\pet-projects\WebSocket-Go\cmd\server\main.go) - точка входа
- [internal/wsserver/server.go](d:\prog\pet-projects\WebSocket-Go\internal\wsserver\server.go) - HTTP и WebSocket сервер
- [internal/wsserver/dto.go](d:\prog\pet-projects\WebSocket-Go\internal\wsserver\dto.go) - структура websocket-сообщения
- [internal/web/templates/html/index.html](d:\prog\pet-projects\WebSocket-Go\internal\web\templates\html\index.html) - клиентская страница чата

## Запуск

```bash
go run ./cmd/server
```

После запуска приложение будет доступно по адресу:

```txt
http://localhost:8080
```

Тестовый endpoint:

```txt
http://localhost:8080/test
```

## Как работает обмен сообщениями

Клиент отправляет JSON:

```json
{
  "message": "hello"
}
```

Сервер дополняет сообщение и рассылает всем клиентам JSON вида:

```json
{
  "address": "192.168.0.5",
  "message": "hello",
  "time": "15:04"
}
```

## Проверка с телефона

Если хочешь открыть чат с телефона:

1. Запусти сервер на Windows.
2. Подключи телефон и компьютер к одной Wi-Fi сети.
3. Открой в телефоне:

```txt
http://<IP_компьютера>:8080
```

Для WebSocket будет использоваться:

```txt
ws://<IP_компьютера>:8080/ws
```

Пример:

```txt
http://192.168.0.5:8080
```

## Зачем этот проект

Проект нужен для практики:
- WebSocket-подключений
- работы с goroutine и каналами
- broadcast-рассылки сообщений
- graceful shutdown сервера
