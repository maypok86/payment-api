# payment-api

![example workflow](https://github.com/maypok86/payment-api/actions/workflows/test.yml/badge.svg)
[![codecov](https://codecov.io/gh/maypok86/payment-api/branch/main/graph/badge.svg?token=E8BJNC6388)](https://codecov.io/gh/maypok86/payment-api)
![Go Report](https://goreportcard.com/badge/github.com/maypok86/payment-api)
![License](https://img.shields.io/github/license/maypok86/payment-api)

# Решение

## Запуск

Чтобы запустить приложение, необходимо выполнить следующую команду в корне репозитория:

```bash
make up
```
Приложение запускается на порту 8080 по умолчанию.

Посмотреть логи запущенного приложения можно командой:

```bash
make logs
```

Остановить приложение можно командой:

```bash
make down
```

Если нужны дополнительные команды, то их описание можно посмотреть в Makefile, либо с помощью команды.

```bash
make help
```

## Общее описание

- Приложение написано на языке Go с использованием чистой архитектуры.
- Приложение разделено на слои `repository`, `domain`, `handler`.
- Для хранения данных используется PostgreSQL.
- Слои `domain` и `handler` покрыты unit-тестами.
- Для запуска приложения используется docker-compose.
- Приложение конфигурируется с помощью .env файла и переменных окружения.
- Баланс хранится и отдаётся в **копейках**, чтобы избежать ошибок округления.

Использованные библиотеки и фреймворки:
- `gin` - для реализации REST API.
- `pgxpool` - для работы с PostgreSQL.
- `squirrel` - для генерации SQL запросов.
- `envconfig` - для работы с переменными окружения.
- `zap` - для логирования.
- `goose` - для миграций.
- `testify` - для написания unit-тестов.
- `gomock` - для генерации моков.

Некоторые вопросы реализации и принятые решения описаны [тут](./docs/NOTES.md).

## Описание API

Для документации api написана [swagger](./api/swagger.yml) документация, а в README приведены чуть более подробное описание и curl запросы.

### Получние баланса по id пользователя

Пример запроса (заменить account_id на нужный id):
```bash
curl --request GET \
  --url http://localhost:8080/api/v1/balance/{account_id} \
  --header 'Content-Type: application/json'
```

### Пополнение баланса

Пример запроса:
```bash
curl --request POST \
  --url http://localhost:8080/api/v1/balance \
  --header 'Content-Type: application/json' \
  --data '{
  "account_id": 1,
  "amount": 100
}'
```

Пример ответа:
```json
{
  "balance": 100
}
```

Возвращается новый баланс пользователя, если пользователя с таким id не существует, то он создаётся c балансом равным переданному значению.

### Перевод денег

Пример запроса:
```bash
curl --request POST \
  --url http://localhost:8080/api/v1/balance/transfer \
  --header 'Content-Type: application/json' \
  --data '{
  "sender_id": 1,
  "receiver_id": 2,
  "amount": 100
}'
```

Пример ответа:
```json
{
  "sender_balance": 100,
  "receiver_balance": 100
}
```

Возвращаются обновлённые балансы отправителя и получателя.

### Создание заказа

Метод создаёт заказ и резервирует деньги пользователя для его оплаты.

Пример запроса:
```bash
curl --request POST \
  --url http://localhost:8080/api/v1/order/create \
  --header 'Content-Type: application/json' \
  --data '{
  "order_id": 1,
  "account_id": 1,
  "service_id": 1,
  "amount": 100
}'
```

Пример ответа:
```json
{
  "order": {
    "order_id": 1,
    "account_id": 1,
    "service_id": 1,
    "amount": 100,
    "is_paid": false,
    "is_cancelled": false,
    "created_at": "2019-08-24T14:15:22Z",
    "updated_at": "2019-08-24T14:15:22Z"
  },
  "balance": {
    "balance": 100
  }
}
```

Возвращается созданный заказ и обновлённый баланс пользователя.

### Оплата заказа

Метод списывает зарезервированные деньги и помечает заказ как оплаченный.

Пример запроса:
```bash
curl --request POST \
  --url http://localhost:8080/api/v1/order/pay \
  --header 'Content-Type: application/json' \
  --data '{
  "order_id": 1,
  "account_id": 1,
  "service_id": 1,
  "amount": 100
}'
```

Тела ответа у этого метода нет, только http status ответа.

### Отмена заказа

Метод отменяет заказ и возвращает зарезервированные деньги пользователю.

Пример запроса:
```bash
curl --request POST \
  --url http://localhost:8080/api/v1/order/cancel \
  --header 'Content-Type: application/json' \
  --data '{
  "order_id": 1,
  "account_id": 1,
  "service_id": 1,
  "amount": 100
}'
```

Пример ответа:
```json
{
  "balance": 100
}
```

Метод возвращает обновлённый баланс пользователя.

### Получить транзакции пользователя

Метод поддерживает пагинацию и сортировку. Параметры пагинации и сортировки передаются в query string.

- 0 <= limit <= 100, default = 10
- 0 <= offset, default = 0
- sort = date | sum, по умолчанию без сортировки
- direction = asc | desc, по умолчанию asc, если sort не задан, то игнорируется

Пример запроса:
```bash
curl --request GET \
  --url http://localhost:8080/api/v1/transaction/{account_id} \
  --header 'Content-Type: application/json'
```

Пример ответа:
```json
{
  "transacions": [
    {
      "transaction_id": 1,
      "type": "enrollment",
      "sender_id": 1,
      "receiver_id": 1,
      "amount": 100,
      "description": "Awesome description",
      "created_at": "2019-08-24T14:15:22Z"
    }
  ],
  "range": {
    "limit": 10,
    "offset": 0,
    "count": 1000
  }
}
```

count - число всего транзакций пользователя.

sender_id и receiver_id могут совпадать, тогда это значит, что пользователь не переводил деньги другому пользователю, а использовал остальные возможности потратить деньги :).

### Получить ссылку на отчёт для бухгалтерии

Пример запроса:
```bash
curl --request POST \
  --url http://localhost:8080/api/v1/report/link \
  --header 'Content-Type: application/json' \
  --data '{
  "month": 10,
  "year": 2022
}'
```

Пример ответа:
```json
{
  "link": "http://localhost:8080/api/v1/report?key=2022-10"
}
```

query параметр key - это строка, в которой закодированы месяц и год в формате год-месяц.

### Скачать отчёт для бухгалтерии

Пример запроса:
```bash
curl --request GET \
  --url http://localhost:8080/api/v1/report/?key=2022-10
```

Тут проблема с тем, что curl требует `/` после report, а в браузере это не так. Поэтому в браузере нужно открыть ссылку, которая пришла в ответе на запрос выше, либо отредактировать ссылку руками.

Пример ответа:
```csv
service_id,amount
1,2
2,40
```

