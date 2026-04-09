# Go test task

## Запуск
docker compose --env-file ./config.env up -d  

## Тесты
go test ./cmd/api -v

## API endpoints
<table>
    <tr>
        <th>http метод</th>
        <th>api endpoint</th>
        <th>Описание</th>
    </tr>
    <tr>
        <td>GET</td>
        <td>/api/v1/healthcheck</td>
        <td>проверка доступности сервиса</td>
    </tr>
    <tr>
        <td>POST</td>
        <td>>/api/v1/wallet</td>
        <td>изменение баланса кошелька</td>
    </tr>
    <tr>
        <td>GET</td>
        <td>/api/v1/wallets/:uuid</td>
        <td>просмотр баланса кошелька</td>
    </tr>
    <tr>
        <td>GET</td>
        <td>/api/v1/wallets</td>
        <td>создание нового кошелька</td>
    </tr>
    <tr>
        <td>GET</td>
        <td>/api/v1/wallets/:uuid</td>
        <td>удаление кошелька</td>
    </tr>
</table>
