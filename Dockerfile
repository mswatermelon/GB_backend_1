FROM golang:1.15 as modules
ADD ./go.mod /m/
RUN cd /m && go mod download

FROM golang:1.15 as builder

COPY --from=modules . ./

RUN mkdir -p /myapp
ADD .. /myapp
WORKDIR /myapp

# Добавляем непривилегированного пользователя
RUN useradd -u 10001 myapp

# Собираем бинарный файл
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
go build -o myapp ./file-server/main.go

FROM scratch

# Не забываем скопировать /etc/passwd с предыдущего стейджа
COPY --from=builder /etc/passwd /etc/passwd
USER myapp

COPY --from=builder /myapp /myapp

CMD ["./myapp/myapp"]

