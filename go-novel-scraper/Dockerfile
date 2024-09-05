FROM golang:1.22-alpine as builder

# 設置工作目錄
WORKDIR /app

# 將Go模組文件複製到容器中
COPY go.mod go.sum ./

# 下載所有依賴
RUN go mod download

COPY . .

RUN go build -o main main.go



FROM alpine:latest 

# 設置工作目錄
WORKDIR /app

COPY --from=builder /app/main .

COPY configs ./configs

COPY entrypoint.sh /usr/local/bin/

RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["entrypoint.sh"]


