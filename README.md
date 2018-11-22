# Rabbit

參考 [RabbitMQ](http://www.rabbitmq.com/) 和 [程序猿成长计划](https://github.com/mylxsw/growing-up) 完成。

## Demo

### Docker

```bash
cd docker; docker-compose up
```

### Receive

```bash
cd receive; go run main.go
```

### Send

```bash
cd send; go run main.go
```

### Log

```bash
cd log; go run main.go
```

### 情境

1. Docker 執行後可以使用 http://localhost:8084 介面查看 RabbitMQ Server。
1. Receive 進行訂閱監聽訊息推送。
1. Send 送出 11 筆資訊，其中一筆資訊會是模擬逾時錯誤。
1. Receive 接收到一般訊息會直接印出。
1. 錯誤訊息會將任務送處重試隊列中處理，重試次數到達上限後會送入錯誤隊列中等待處理。
1. Log 會監聽錯誤隊列中的訊息，直接取出訊息印出，不刪除對列上的訊息(回傳 Ack 則是會刪除，見`send/main.go:61`)。