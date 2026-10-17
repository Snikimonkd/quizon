# quizon

Сайт [КвизON](https://quiz-on.ru): регистрация команд на квизы в МГТУ им. Н.Э. Баумана. Капитаны записывают команду с телефона, организаторы создают игры, смотрят и удаляют участников и выгружают их в xlsx.

Один Go-бинарник без отдельного фронтенда:
- **сервер:** Go, [chi](https://github.com/go-chi/chi);
- **страницы:** `html/template`, [htmx](https://htmx.org) и [Pico.css](https://picocss.com), шаблоны и статика вшиты через `go:embed`;
- **база:** PostgreSQL 17, запросы на [go-jet](https://github.com/go-jet/jet), миграции на [goose](https://github.com/pressly/goose).

## Локальный запуск

Нужны Go (версия в `go.mod`) и Docker.

```sh
make bin-deps      # goose, jet и golangci-lint в ./bin
make postgres-up   # Postgres 17 в Docker на localhost:5432, без пароля
make migrate-up    # накатить миграции
make run           # http://localhost:8000
```

`make run` запускает приложение с `DEBUG=1`. В этом режиме:
- **адрес** `:8000` на всех интерфейсах, поэтому сайт можно открыть с телефона в той же сети по `http://<IP компьютера>:8000`; IP показывает `ipconfig getifaddr en0`;
- **логи** читаемым текстом и с уровнем debug.

### Организатор

Регистрации для организаторов нет, их добавляют прямо в базу. Пароль хранится как bcrypt-хеш:

```sh
hash=$(htpasswd -bnBC 12 "" 'пароль' | tr -d ':\n')
psql 'postgres://postgres@localhost:5432/postgres?sslmode=disable' \
  -c "INSERT INTO admins (login, password) VALUES ('admin', '$hash')"
```

Кука входа ставится с `Secure`, поэтому локально вход работает только на `http://localhost` и только в браузерах, которые считают localhost безопасным, например Chrome и Firefox. По IP-адресу (с телефона) войти не получится, для этого нужен https-туннель, например `cloudflared tunnel --url http://localhost:8000` и `make run DOMAIN=<адрес туннеля>`.

## Переменные окружения

| Переменная | Что это |
|---|---|
| `PG_DSN` | строка подключения к Postgres, обязательна |
| `DOMAIN` | домен для куки входа: `localhost` локально, `quiz-on.ru` на проде, обязательна: без неё приложение не стартует |
| `DEBUG` | задана с любым значением — локальная разработка: адрес `:8000` и текстовые логи. Без неё адрес `127.0.0.1:8000` и JSON-логи |

## Тесты, линтер, кодогенерация

```sh
make test       # нужен Postgres из make postgres-up с накатанными миграциями
make test-cov   # то же с отчётом о покрытии
make lint       # golangci-lint с .golangci.yml
make mocks      # перегенерировать моки после изменения интерфейсов
```

Тесты репозитория ходят в локальную базу. Остальные слои проверяются на моках [gomock](https://github.com/uber-go/mock): `delivery` на моке usecase, `usecase` на моке репозитория. Моки генерирует mockgen по директивам `//go:generate` рядом с интерфейсами и кладёт в соседний пакет `mock`, руками их не правь. Мок репозитория ссылается на типы usecase, поэтому тесты usecase на моках живут во внешнем пакете `usecase_test`: из самого `usecase` импорт `mock` замкнул бы цикл. mockgen подключён в `go.mod` через `tool`, отдельно его ставить не нужно. Данные команд в тестах случайные, их генерирует [gofakeit](https://github.com/brianvoe/gofakeit) через `testsupport.FakeRegistration`.

### Миграции

```sh
make create-migration NAME=add_something   # новый файл в migrations/
make migrate-up
make jet                                    # перегенерировать internal/generated после изменения схемы
```

Уже применённые миграции не правь. goose помнит их версии и второй раз не накатит, так что на проде изменение не появится. Любое изменение схемы оформляй новой миграцией.

## Устройство

```
cmd/                    точка входа
internal/app/delivery   HTTP-обработчики, шаблоны (templates/) и статика (static/)
internal/app/usecase    бизнес-логика: правила игр и анкет, регистрация, места и резерв, удаление команд, вход
internal/app/repository запросы к Postgres без бизнес-логики: каждый метод — один запрос
internal/app/*/mock     моки gomock для тестов, генерирует make mocks
internal/config         сервер, подключение к базе, журнал запросов
internal/pkg            cache (сессии в памяти), logger (slog), testsupport
internal/generated      код go-jet по схеме базы, руками не править
migrations              миграции goose
```

### Логи

Приложение пишет в stdout через `log/slog`. На проде это JSON по записи на строку:
- **каждый запрос** — одна запись `request` с методом, путём, статусом, временем и `request_id`;
- **ошибки обработчиков и паники** пишутся с тем же `request_id`, так что их легко связать с запросом.

```sh
docker logs quizon | jq 'select(.level == "ERROR")'
```

## Деплой

Всё делает пайплайн GitHub Actions (`.github/workflows/pipeline.yml`) на пуш в `main` и на теги `v*`:

1. **tests** и **lint** — параллельно.
2. **build** — образ `snikimonk/quizon` для linux/amd64 и linux/arm64 в Docker Hub с тегами `latest`, `sha-<коммит>` и версией из тега.
3. **migrate** (только `main`) — SSH-туннель к серверу и `goose up`.
4. **deploy** (только `main`) — по SSH `docker pull` и перезапуск контейнера.

`migrate` и `deploy` можно запустить и вручную через Run workflow. `deploy` при этом принимает тег образа, что удобно для отката на прошлую версию.

### Секреты и переменные репозитория

| Секрет | Что это |
|---|---|
| `DOCKERHUB_TOKEN` | токен Docker Hub с правом записи |
| `SSH_PRIVATE_KEY` | приватный ключ деплойного пользователя на сервере |
| `SSH_KNOWN_HOSTS` | `ssh-keyscan -H <сервер>` |
| `PG_DSN` | строка подключения к базе на сервере через `localhost:5432` |

| Переменная | Что это |
|---|---|
| `DOCKERHUB_USERNAME` | пользователь Docker Hub |
| `SSH_HOST`, `SSH_USER` | сервер и деплойный пользователь |
| `DOMAIN` | `quiz-on.ru` |

### Сервер

- **Postgres 17** работает прямо на сервере и слушает только localhost.
- **Контейнер** запускается с `--network host`. Его `localhost` — это сам сервер, поэтому одна и та же `PG_DSN` подходит и приложению, и туннелю миграций. Приложение без `DEBUG` слушает только `127.0.0.1:8000`, и снаружи порт закрыт.
- **[Caddy](https://caddyserver.com)** стоит спереди: сам получает и продлевает сертификат Let's Encrypt и перенаправляет http на https. `/etc/caddy/Caddyfile`:

  ```
  quiz-on.ru {
      encode zstd gzip
      reverse_proxy 127.0.0.1:8000
  }

  www.quiz-on.ru {
      redir https://quiz-on.ru{uri} permanent
  }
  ```

- **Файрвол (ufw)** открывает только 22, 80 и 443.
