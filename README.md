# Gateway Startup

HTTP/gRPC gateway для сервисов Startup.

## Безопасная регистрация

> Это целевая схема. До её полной реализации API может отличаться от описания ниже.

Аккаунт нельзя создавать с заданными именем и паролем до подтверждения email. Иначе злоумышленник может начать регистрацию с чужим email и своим паролем, а владелец email может случайно активировать такой аккаунт.

Регистрация состоит из трёх этапов.

### 1. Запрос подтверждения email

```http
POST /auth/register/request
Content-Type: application/json
```

```json
{
  "email": "user@example.com"
}
```

Сервис создаёт одноразовую verification-сессию и отправляет на email код или ссылку. Имя и пароль на этом этапе не принимаются.

### 2. Подтверждение email

```http
POST /auth/register/confirm
Content-Type: application/json
```

```json
{
  "email": "user@example.com",
  "code": "123456"
}
```

После успешной проверки сервер выдаёт короткоживущий `signup_token`:

```json
{
  "signup_token": "eyJ..."
}
```

`signup_token` — это JWT, который только разрешает завершить регистрацию. Он не является access-токеном.

Минимальный набор claims:

```json
{
  "sub": "user@example.com",
  "purpose": "signup",
  "jti": "unique-token-id",
  "iat": 1710000000,
  "exp": 1710000900
}
```

### 3. Завершение регистрации

```http
POST /auth/register/complete
Authorization: Bearer <signup_token>
Content-Type: application/json
```

```json
{
  "name": "Weit User",
  "password": "secret123"
}
```

Сервер берёт email только из подписанного JWT. Email из JSON-тела не принимается. Без корректного `signup_token` запрос завершается ошибкой `401 Unauthorized` или `403 Forbidden`.

После успешного создания аккаунта signup-сессия погашается, а клиент получает обычные access- и refresh-токены.

## Требования безопасности

- `signup_token` действует 10–15 минут.
- JWT содержит `purpose: "signup"` и не принимается middleware для access-токенов.
- Для signup-токенов используются отдельные `issuer`, `audience` и желательно отдельный ключ.
- `jti` хранится в Redis с TTL, например: `signup:<jti> -> email`.
- Завершение регистрации и погашение `jti` выполняются атомарно. Повторное использование `jti` запрещено.
- На нормализованный email создаётся уникальный индекс в БД.
- После создания аккаунта все оставшиеся verification- и signup-сессии для email аннулируются.
- На запрос письма и проверку кода настраивается rate limit.
- API возвращает одинаковый нейтральный ответ для существующего и неизвестного email.

## Типы токенов

| Токен | Назначение |
| --- | --- |
| Verification token/code | Доказывает доступ к email |
| Signup token | Разрешает однократно завершить регистрацию |
| Access token | Даёт доступ к защищённому API |
| Refresh token | Обновляет access-токен |

