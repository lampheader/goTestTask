Go test task

Запуск: 
docker compose --env-file ./config.env up -d

http.MethodGet,    "localhost:4000/api/v1/healthcheck"   -   проверка доступности сервиса
http.MethodPost,   "localhost:4000/api/v1/wallet"        -   изменение баланса кошелька
http.MethodGet,    "localhost:4000/api/v1/wallets/:uuid" -   просмотр баланса кошелька
http.MethodGet,    "localhost:4000/api/v1/wallets"       -   создание нового кошелька
http.MethodDelete, "localhost:4000/api/v1/wallets/:uuid" -   удаление кошелька

Тесты:
go test ./cmd/api -v
