# Drift

Один Go-бинарник, в который встроены Vue-приложение и две серверные страницы входа. Пользователями и сессиями управляет Ory Kratos.

## Как устроено

```
браузер ──► drift :8080 ─┬─ /setup, /login, /error, /logout   Go-шаблоны + htmx + _hyperscript + @ui-sdk/web
                         ├─ /.ory/kratos/public/*              обратный прокси в Kratos public (:4433)
                         ├─ /healthz, /readyz
                         └─ всё остальное                      Vue SPA (только с сессией Kratos)
drift ──► Kratos admin :4434   проверка «есть ли пользователи», создание суперпользователя
```

- **Нет ни одного пользователя**: любой адрес ведёт на `/setup`. Форма отправляется через htmx, Go создаёт identity через admin API Kratos с `metadata_public.role = "superuser"`. Регистрация в Kratos выключена (`selfservice.flows.registration.enabled: false`), поэтому других путей создать пользователя нет.
- **Пользователь есть**: `/login` рендерит login flow Kratos: email + пароль или passkey (отпечаток, Face ID, аппаратный ключ). Passkey также предлагается в автозаполнении поля email.
- **Passkey добавляют** после первого входа в SPA на странице «Ключи доступа» (`/settings`, settings flow Kratos).
- Kratos открыт на том же origin (`/.ory/kratos/public`), поэтому cookie, CSRF и origin для WebAuthn общие.
- Скрипты страниц входа (htmx, _hyperscript, веб-компоненты UI SDK) собирает Vite из `web/src/auth/main.ts`. CDN нигде не используется.

## Раскладка

| Путь | Что там |
| --- | --- |
| `app/` | Go-модуль: `cmd/server`, `internal/{auth,kratos,frontend,config,...}` |
| `app/internal/auth/templates` | layout + 2 страницы (`setup.html`, `login.html`) |
| `app/internal/frontend/dist` | результат `vite build`, встраивается через `go:embed` (в git не хранится) |
| `web/` | исходники Vue и бандла страниц входа (Vite) |
| `deploy/kratos` | `kratos.yml` и схема identity |

`web/` переносить не нужно: `go:embed` не видит родительские каталоги, поэтому Vite сам пишет сборку внутрь Go-модуля (`build.outDir`).

## Рабочее окно

SPA (`web/src`) — оболочка для работы с агентами: проекты и сессии в боковой панели, чат, файлы проекта.

- **Проекты** — папки внутри `DRIFT_WORKSPACE_ROOT` (по умолчанию `$HOME/work`). Удаление проекта удаляет только записи в БД, файлы остаются.
- **Боковая панель** — `Resizable` из ui-sdk: ширина и свёрнутое состояние сохраняются в `localStorage` (`ui-sdk-resizable:drift-shell`), сторона (слева/справа), тема и язык — через `createPreference` (`data-sidebar-side`, `data-color-scheme`, `data-locale` на `<html>`). Меняются в «Настройках».
- **Черновики** поля ввода хранятся на сервере (ключ `session:<id>` или `new:<project id>`), сохраняются через 400 мс после последнего нажатия и сразу при уходе со страницы. Другие устройства видят текст и пометку «печатает».
- **Файлы** — дерево `Tree` из ui-sdk с ленивой загрузкой папок и просмотр в CodeMirror 6 (только чтение, до 1 МБ).
- **Анимации** — токены `--dur-*`/`--ease-*` в `web/src/styles/motion.css`; раскрытие блоков через `grid-template-rows: 0fr → 1fr`, списки — глобальная директива `v-auto-animate` с теми же длительностями.
- Вкладки CLI, Git и Терминал и запуск агентов — следующий этап.

### API и realtime

JSON API под `/api` требует сессию Kratos (идентичность кешируется на 30 с). Имена по Google API design guide: `page_size`/`page_token` → `nextPageToken` (AIP-158), `createTime`/`updateTime` (AIP-142).

| Метод | Путь |
| --- | --- |
| `GET`, `POST` | `/api/projects` |
| `PATCH`, `DELETE` | `/api/projects/{id}` |
| `GET`, `POST` | `/api/projects/{id}/sessions` (POST создаёт сессию и первое сообщение) |
| `GET` | `/api/projects/{id}/tree?path=`, `/api/projects/{id}/file?path=` |
| `GET`, `PATCH`, `DELETE` | `/api/sessions/{id}` |
| `GET`, `POST` | `/api/sessions/{id}/messages` |
| `GET` | `/api/drafts`; `PUT`, `DELETE` `/api/drafts/{key}` |
| `GET` | `/api/events` — Server-Sent Events |

`/api/events` — одно соединение на вкладку, без брокера: события пользователя рассылаются из памяти процесса, последние 256 хранятся для переподключения по `Last-Event-ID`. Если пропущено больше или процесс перезапускался, приходит `resync`, и клиент перезапрашивает данные. Каждое изменение несёт `origin` — id устройства из заголовка `X-Drift-Device`, чтобы вкладка не применяла собственное эхо.

Контракт описан в `app/api/openapi.yaml`. Фронтенд не пишет запросы руками: Kubb (`web/kubb.config.ts`) генерирует из спеки типы, функции запросов и Vue Query хуки (query, mutation, infinite по `page_token`) в `web/src/gen`, транспорт — `web/src/shared/http/client.ts` на `URL`/`Request`/`Response`. После правки спеки: `pnpm generate` в `web/`.

Миграции из `app/db/migrations` применяются при старте бинарника (таблица `schema_migrations` совместима с `migrate` из taskfile). SQL-запросы — `app/db/queries`, код генерирует `task app:sqlc`.

## Запуск целиком в контейнерах

```sh
podman compose up -d --build   # или task stack
```

Открывать http://localhost:8080. `.env` не обязателен: без него берутся значения по умолчанию из `compose.yml` (для локального запуска). `Containerfile` собирает Vite, потом Go-бинарник с встроенным фронтендом, и кладёт его в alpine; фронт и бэк живут в одном контейнере `drift`.

## Разработка

```sh
cp .env.example .env        # смените пароли и секреты Kratos
task up                     # только Postgres + Kratos (БД kratos создаётся и мигрируется автоматически)
task web:install
task dev                    # Vite :5173 + Go (air, тег dev) :8080 → открывать http://localhost:8080
```

В dev-сборке (`-tags dev`) Go проксирует фронтенд на Vite вместе с HMR. Release:

```sh
task build                  # vite build → go build → bin/drift
```

Бинарнику нужны только переменные окружения из `.env.example` и запущенный Kratos.

### Проверки

Правила кода — в `CLAUDE.md` (из [dotconfig](https://github.com/e1berd/dotconfig)). Перед коммитом lefthook форматирует и линтует изменённые файлы `web/`. Вручную из `web/`: `pnpm format:check`, `pnpm lint`, `pnpm types:check`, `pnpm vitest run`.

## Продакшен

`deploy/kratos/kratos.yml` рассчитан на `http://localhost:8080` и запускается с `--dev`. Для боевого домена поменяйте `base_url`, все `ui_url`, `allowed_return_urls`, `passkey.config.rp.id/origins`, уберите `--dev` (нужен HTTPS) и задайте `DRIFT_PUBLIC_URL`.
