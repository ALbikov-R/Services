# Используйте официальный образ Go
FROM golang:latest

# Копирование исходного кода в контейнер
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

# Копирование остального кода
COPY . .

# Компиляция приложения
RUN go build -o main .

# Экспорт порта
EXPOSE 8080

# Команда для запуска приложения
CMD ["./main"]
