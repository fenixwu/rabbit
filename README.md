# Rabbit

參考 [RabbitMQ](http://www.rabbitmq.com/) 和 [程序猿成长计划](https://github.com/mylxsw/growing-up) 完成。

## Demo

1. Docker 執行後可以使用 http://localhost:8084 介面查看 RabbitMQ Server。
1. Receive 進行監聽任務隊列中的訊息。
1. Log 監聽錯誤隊列中的訊息。
1. Send 送出 2 筆訊息，其中一筆模擬錯誤。
1. Receive 接收到訊息。正常訊息會直接印出；錯誤訊息被送到重試隊列中重新推送，重試次數到達上限後會送到錯誤隊列中等待。
1. Log 接收到錯誤隊列中的訊息，直接取出訊息印出。