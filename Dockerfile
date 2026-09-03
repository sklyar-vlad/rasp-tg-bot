# ЭТАП 1: Сборка (Builder)
FROM golang:1.22-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы модулей (для кэширования скачивания зависимостей)
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем статический бинарный файл (CGO_ENABLED=0 важно для Alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot-app .

# ЭТАП 2: Запуск (Runner)
FROM alpine:latest


WORKDIR /root/

# Копируем скомпилированный бинарник из этапа сборки
COPY --from=builder /app/bot-app .

# Команда для запуска бота
CMD ["./bot-app"]