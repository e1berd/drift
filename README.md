# Drift

Один Go-бинарник, в который встроены Vue-приложение и две серверные страницы входа. Пользователями и сессиями управляет Ory Kratos.

## Быстрый запуск

Скачать последний релиз под свою архитектуру и сразу запустить:

```sh
mkdir -p ~/drift/bin && cd ~/drift && curl -fsSL "https://github.com/e1berd/drift/releases/latest/download/drift_linux_$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/').gz" | gunzip > bin/drift && chmod +x bin/drift && DATABASE_URL='postgres://drift:drift@localhost:5432/drift?sslmode=disable' ./bin/drift
```

Открывать http://localhost:8080. Первый вход ведёт на `/setup`, там создаётся суперпользователь.

Перед запуском должны работать:

- **PostgreSQL**: база для Drift, строка подключения передаётся в `DATABASE_URL`. Миграции применяются при старте бинарника.
- **Ory Kratos** с конфигом из `deploy/kratos` (в нём выключена регистрация и задана схема identity). Public API нужен на `KRATOS_PUBLIC_URL` (по умолчанию `http://localhost:4433`), admin API — на `KRATOS_ADMIN_URL` (по умолчанию `http://localhost:4434`). Kratos нужна своя база в том же Postgres.

Проще всего поднять обоих из клона репозитория: `podman compose up -d postgres kratos`, или `task up` (подробнее в разделе «Разработка»). Остальные переменные окружения перечислены в `.env.example`. Когда Drift открыт не на localhost, задайте `DRIFT_PUBLIC_URL`. Бинарник лежит в `~/drift/bin`, папка доступна на запись, поэтому Drift сможет обновлять себя сам (см. «Обновления»).

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
- **Passkey необязателен.** `KRATOS_PASSKEY_ENABLED=false` выключает метод в Kratos. Кнопка входа и раздел в настройках показываются, только если Kratos прислал данные passkey, браузер поддерживает WebAuthn, страница открыта в secure context (HTTPS или localhost) и её домен совпадает с `rp.id` из `deploy/kratos/kratos.yml`; иначе их нет, а вход остаётся по паролю.
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

- **Проекты** — папки внутри `DRIFT_WORKSPACE_ROOT`. По умолчанию это `./workspace` в корне репозитория (рядом с `app/` и `web/`, в `.gitignore`); сервер создаёт папку при старте, если её нет. Без переменной бинарник берёт `workspace` в текущей рабочей папке, в образе — `/workspace`. `compose.yml` монтирует `./workspace` в `/workspace`, а данные Postgres — в `./data/postgres`; у обоих сервисов `userns_mode: keep-id`, поэтому файлы на хосте принадлежат вашему пользователю. Удаление проекта удаляет только записи в БД, файлы остаются.
- **Чаты без проекта** — «Новый чат» создаёт сессию без проекта (`POST /api/sessions`). Её агент работает в своей пустой папке `DRIFT_CHATS_DIR/<id сессии>` (по умолчанию `<рабочий каталог>/.chats`, поэтому она сохраняется вместе с `./workspace` и не видна в выборе папки проекта). Удаление чата удаляет и его папку.
- **Боковая панель** — `Resizable` из ui-sdk: ширина и свёрнутое состояние сохраняются в `localStorage` (`ui-sdk-resizable:drift-shell`), сторона (слева/справа), тема и язык — через `createPreference` (`data-sidebar-side`, `data-color-scheme`, `data-locale` на `<html>`). Меняются в «Настройках».
- **Черновики** поля ввода хранятся на сервере (ключ `session:<id>` или `new:<project id>`), сохраняются через 400 мс после последнего нажатия и сразу при уходе со страницы. Другие устройства видят текст и пометку «печатает».
- **Файлы** — дерево `Tree` из ui-sdk с ленивой загрузкой папок; кнопки «Новый файл» / «Новая папка» в шапке дерева (путь вида `src/app/x.ts`, недостающие папки создаются). Файл открывается в CodeMirror 6, правка сохраняется по Ctrl+S (до 1 МБ; больше — только чтение). Несохранённые правки живут в памяти вкладки по пути файла.
- **Терминал** — вкладка с xterm.js и интерактивным шеллом в папке сессии (`GET /api/sessions/{id}/terminal`, WebSocket: бинарные кадры — байты терминала, текстовые — JSON `TerminalControl`/`TerminalEvent`). Шелл живёт на сервере, а не во вкладке: перезагрузка страницы или второе окно подключаются к тому же шеллу и сначала получают последние 256 КБ вывода. Шелл без зрителей завершается через 30 минут, удаление сессии завершает его сразу; одновременно — до 8 шеллов. Шелл задаёт `DRIFT_SHELL`, по умолчанию первый найденный из fish, bash, zsh, sh. Окружение — безопасный минимум (`PATH`, `HOME`, `LANG`, прокси…), секреты сервера вроде `DATABASE_URL` в шелл не попадают.
- **ИИ-провайдеры** (`/settings`) — список подписок агентов, см. ниже.
- **Анимации** — токены `--dur-*`/`--ease-*` в `web/src/styles/motion.css`; раскрытие блоков через `grid-template-rows: 0fr → 1fr`, списки — глобальная директива `v-auto-animate` с теми же длительностями.
- **Git** — панель для папки сессии: ветка, upstream, ahead/behind, списки «Подготовлено» и «Изменения» с буквами статуса, diff выбранного файла (неотслеживаемый файл — целиком как добавленный, до 512 КБ), подготовка и отмена подготовки по файлу и всех сразу, отмена изменений с подтверждением (неотслеживаемый файл удаляется), коммит, переключение и создание веток, `pull --ff-only` и `push` (без upstream — `--set-upstream origin`), история коммитов. Папка, которая не является корнем репозитория, показывает кнопку `git init`; репозитории выше папки сессии не ищутся (`GIT_CEILING_DIRECTORIES`), поэтому чат внутри чужого checkout не видит его. Сервер запускает `git` CLI с безопасным окружением (как у терминала), `GIT_TERMINAL_PROMPT=0`, таймаутами 15 с (локальные команды) и 2 мин (pull/push); ошибки git приходят в тост как есть (HTTP 409).
- Вкладка CLI — следующий этап.

### Подписки агентов

Drift не ходит в API провайдеров сам, а запускает официальные CLI: `claude` (Claude Code) и `codex` лежат в образе (`/usr/local/bin`, версии — `ARG` в `Containerfile`). У каждой подписки своя папка `DRIFT_AGENTS_DIR/<id>` (в compose — `./data/agents`, в `.gitignore`) — это `HOME`, `CLAUDE_CONFIG_DIR=<home>/.claude` и `CODEX_HOME=<home>/.codex` для процессов агента, поэтому аккаунтов одного агента может быть несколько. Учётные данные лежат только там и удаляются вместе с подпиской.

| Агент | Вход | Лимиты |
| --- | --- | --- |
| Claude Code | `claude auth login --claudeai`: Drift показывает ссылку, страница Anthropic выдаёт код, его вставляют в Drift (уходит в stdin CLI). Запасной путь — токен из `claude setup-token` (`CLAUDE_CODE_OAUTH_TOKEN`, проверяется через `GET /v1/models`) | `GET api.anthropic.com/api/oauth/usage` с токеном из `.credentials.json` — недокументированный эндпоинт панели `/usage`, может сломаться. С токеном из `setup-token` недоступны (нет scope `user:profile`) |
| Codex | `codex app-server`, `account/login/start` с `chatgptDeviceCode`: ссылка и код, вход завершается сам (`account/login/completed`). В ChatGPT должен быть разрешён вход по коду устройства | `account/rateLimits/read` того же app-server |
| Antigravity | пока нет: `agy` без графического окружения требует keyring/D-Bus | — |

Текущая подписка — одна на агента (`is_current`, уникальный индекс), первая подключённая становится текущей сама, при удалении текущей текущей становится последняя из оставшихся. Агент доступен в новом чате, если у него есть активная подписка. Незавершённые входы живут в памяти процесса (до 15 минут, не больше 4 одновременно); после перезапуска они помечаются как неудачные, и вход начинают заново. Лимиты кешируются на минуту. Контейнеру `drift` в `compose.yml` дано 1.5 ГБ памяти — CLI агентов тяжёлые.

### Запуск агентов

Сообщение пользователя запускает CLI агента в папке сессии с окружением текущей подписки; пока агент работает, новые сообщения копятся и уходят следующим промптом вместе с картинками. Одновременно — не больше двух запусков.

| Агент | Команда | Режимы разрешений |
| --- | --- | --- |
| Claude Code | `claude -p --output-format stream-json [--resume id]`, промпт в stdin; с картинками — `--input-format stream-json` и блоки `image` (base64) | `--permission-mode` как есть, `default` — без флага (инструменты, которым нужно подтверждение, отклоняются) |
| Codex | `codex exec [resume <thread id>] --json --skip-git-repo-check -c approval_policy="never" [--image=…] -- -`, промпт в stdin, id треда из `thread.started` | `default` и `plan` — `sandbox_mode="read-only"`, `acceptEdits` и `auto` — `"workspace-write"`, `bypassPermissions` — `--dangerously-bypass-approvals-and-sandbox` |
| Antigravity | `agy --output-format stream-json [--conversation id] --prompt=…`; картинки — пути к временным файлам в конце промпта | `--mode accept-edits\|auto\|plan`, `bypassPermissions` — `--dangerously-skip-permissions` |

**Картинки** (`POST /api/attachments`, тело — байты файла, `?name=`): PNG, JPEG, GIF или WebP до 5 МБ (тип определяется по содержимому), до 10 на сообщение. Хранятся в Postgres (`attachments.data`), отдаются по `GET /api/attachments/{id}/content`. Загрузка ни к чему не привязана, пока её id не придёт в `attachmentIds` нового сообщения или сессии; непривязанные удаляются через сутки. Для Codex и Antigravity картинки на время запуска пишутся во временную папку.

### API и realtime

JSON API под `/api` требует сессию Kratos (идентичность кешируется на 30 с). Имена по Google API design guide: `page_size`/`page_token` → `nextPageToken` (AIP-158), `createTime`/`updateTime` (AIP-142).

| Метод | Путь |
| --- | --- |
| `GET`, `POST` | `/api/projects` |
| `PATCH`, `DELETE` | `/api/projects/{id}` |
| `GET`, `POST` | `/api/projects/{id}/sessions` (POST создаёт сессию и первое сообщение) |
| `GET`, `POST` | `/api/projects/{id}/tree?path=` (POST создаёт пустой файл или папку, AIP-133) |
| `GET`, `PATCH` | `/api/projects/{id}/file?path=` (PATCH сохраняет текст, AIP-134) |
| `GET`, `POST` | `/api/subscriptions` (POST начинает вход или принимает токен) |
| `DELETE` | `/api/subscriptions/{id}` |
| `POST` | `/api/subscriptions/{id}:activate`, `:startLogin`, `:completeLogin` (AIP-136) |
| `GET` | `/api/subscriptions/{id}/usage` (AIP-156) |
| `GET`, `PATCH`, `DELETE` | `/api/sessions/{id}` |
| `GET`, `POST` | `/api/sessions/{id}/messages` (POST принимает `text` и/или `attachmentIds`) |
| `POST` | `/api/attachments?name=` (тело — картинка); `GET` `/api/attachments/{id}/content` |
| `GET` | `/api/sessions/{id}/git` (статус, AIP-156), `/git/diff?path=&staged=`, `/git/branches`, `/git/commits` (AIP-158) |
| `POST` | `/api/sessions/{id}/git:init`, `:stage`, `:unstage`, `:discard`, `:commit`, `:checkout`, `:pull`, `:push` (AIP-136, отвечают свежим статусом) |
| `GET` | `/api/drafts`; `PUT`, `DELETE` `/api/drafts/{key}` |
| `GET` | `/api/delivery` (AIP-156), `/api/delivery/releases`; `POST` `/api/delivery:check`, `:install` |
| `GET` | `/api/events` — Server-Sent Events |

`/api/events` — одно соединение на вкладку, без брокера: события пользователя рассылаются из памяти процесса, последние 256 хранятся для переподключения по `Last-Event-ID`. Если пропущено больше или процесс перезапускался, приходит `resync`, и клиент перезапрашивает данные. Каждое изменение несёт `origin` — id устройства из заголовка `X-Drift-Device`, чтобы вкладка не применяла собственное эхо.

Контракт описан в `app/api/openapi.yaml`. Фронтенд не пишет запросы руками: Kubb (`web/kubb.config.ts`) генерирует из спеки типы, функции запросов и Vue Query хуки (query, mutation, infinite по `page_token`) в `web/src/gen`, транспорт — `web/src/shared/http/client.ts` на `URL`/`Request`/`Response`. После правки спеки: `pnpm generate` в `web/`.

Миграции из `app/db/migrations` применяются при старте бинарника (таблица `schema_migrations` совместима с `migrate` из taskfile). SQL-запросы — `app/db/queries`, код генерирует `task app:sqlc`.

## Обновления (delivery)

Drift обновляет себя сам из [GitHub Releases](https://github.com/e1berd/drift/releases).

- **Версия** зашивается при сборке: `-ldflags "-X github.com/e1berd/drift/app/internal/delivery.version=v1.2.0"`. `task build` берёт её из `git describe --tags`, CI — из тега, без флага бинарник считается dev-сборкой. `drift version` печатает её.
- **Релиз** выпускается командой `task release` (нужны `gh auth login` и право push): из чистого master, совпадающего с `origin/master` или опережающего его, ставится тег-дата `vГГГГ.ММДД.N` по UTC (`v2026.1118.0`, второй релиз за день — `v2026.1118.1`; это валидный semver, и версии растут), master и тег пушатся атомарно, команда ждёт workflow и печатает ссылку на релиз. `task release:next` показывает следующую версию. Workflow `.github/workflows/release.yml` собирает `drift_linux_{amd64,arm64}.gz` и `checksums.txt` (sha256) и публикует релиз с автоматическими release notes; их можно отредактировать на GitHub — Drift показывает текст релиза как есть. Тег с суффиксом (`v1.3.0-rc.1`) становится пре-релизом: его можно поставить вручную, но уведомления о нём нет.
- **Проверка**: сервер раз в час (и через 10 с после старта) читает релизы и число звёзд через публичный API GitHub (`DRIFT_DELIVERY_REPO`, по умолчанию `e1berd/drift`; `DRIFT_DELIVERY_API_URL`). Браузеры к GitHub не ходят, новый релиз приходит всем вкладкам событием `delivery.updated`.
- **Уведомление**: при новой версии у всех пользователей открывается модальное окно с текущей и новой версией и release notes. «Не обновлять» сначала спрашивает подтверждение; подтверждённый отказ откладывает напоминание на сутки (в `localStorage` устройства). Внизу боковой панели — текущая версия (открывает выбор любой версии из последних 30 релизов, в том числе откат), кнопка новой версии и ссылка на GitHub со звёздами.
- **Установка**: сервер скачивает архив своей платформы, сверяет sha256 с `checksums.txt`, распаковывает рядом с бинарником, проверяет, что `<новый бинарник> version` отвечает нужным тегом, атомарно подменяет файл (`rename`), корректно останавливается и делает `exec` себя же — pid не меняется, systemd этого не замечает. Вкладки опрашивают `/api/delivery` и перезагружаются, когда видят новую версию. Работающие агенты прерываются.
- **Ограничения**: процессу нужна запись в папку бинарника (на сервере это `~/drift/bin`), иначе обновление недоступно (`installBlocker: read_only`) — так в контейнере, где образ обновляется пересборкой. Миграции БД при откате не откатываются. Обновлять может любой вошедший пользователь.

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

На сервере `drift` работает бинарником на хосте под пользователем (`systemd --user`), а не в контейнере: агенты и терминал видят toolchain проектов (node, pnpm, go, python, podman), `~/.gitconfig` и SSH-ключи, а лишнего контейнерного слоя нет. В podman — только Postgres и Kratos, в сети хоста и на `127.0.0.1`, без pod, `pasta` и `rootlessport`. HTTPS и сжатие — во внешнем reverse proxy (Caddy).

Сборка бинарника на машине разработчика (`GOAMD64=v3` — для CPU с AVX2):

```sh
task web:build && (cd app && GOAMD64=v3 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ../bin/drift ./cmd/server)
```

Каталог на сервере (`~/drift`): `compose.yml`, `deploy/kratos/`, `bin/drift`, `.env`, `compose.override.yml`, данные в `data/`. `compose.override.yml` в `.gitignore`, podman-compose подхватывает его сам. Пример для продакшена:

```yaml
x-podman:
  in_pod: false

x-kratos-dsn: &kratos-dsn postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@127.0.0.1:5432/kratos?sslmode=disable&max_conns=5&max_idle_conns=1

services:
  drift:
    profiles: [container]

  postgres:
    network_mode: host
    ports: !reset []
    mem_limit: 256m
    restart: always
    shm_size: 64m
    healthcheck:
      interval: 30s
    command:
      - postgres
      - -c
      - listen_addresses=127.0.0.1
      - -c
      - max_connections=25
      - -c
      - shared_buffers=48MB
      - -c
      - effective_cache_size=192MB
      - -c
      - work_mem=2MB
      - -c
      - maintenance_work_mem=32MB
      - -c
      - wal_buffers=2MB
      - -c
      - min_wal_size=32MB
      - -c
      - max_wal_size=256MB
      - -c
      - jit=off
      - -c
      - max_worker_processes=2
      - -c
      - max_parallel_workers=0
      - -c
      - max_parallel_workers_per_gather=0
      - -c
      - autovacuum_max_workers=1

  kratos-db:
    network_mode: host
    environment:
      PGHOST: 127.0.0.1

  kratos-migrate:
    network_mode: host
    environment:
      DSN: *kratos-dsn

  kratos:
    network_mode: host
    ports: !reset []
    mem_limit: 128m
    restart: always
    command: serve -c /etc/config/kratos/kratos.yml
    healthcheck:
      interval: 30s
    environment:
      DSN: *kratos-dsn
      LOG_LEVEL: warning
      GOMEMLIMIT: 48MiB
      SERVE_PUBLIC_HOST: 127.0.0.1
      SERVE_ADMIN_HOST: 127.0.0.1
      SERVE_PUBLIC_BASE_URL: https://${DRIFT_DOMAIN}/.ory/kratos/public/
      SERVE_ADMIN_BASE_URL: http://127.0.0.1:4434/
      SELFSERVICE_DEFAULT_BROWSER_RETURN_URL: https://${DRIFT_DOMAIN}/
      SELFSERVICE_ALLOWED_RETURN_URLS: https://${DRIFT_DOMAIN}
      SELFSERVICE_METHODS_PASSKEY_CONFIG_RP_ID: ${DRIFT_DOMAIN}
      SELFSERVICE_METHODS_PASSKEY_CONFIG_RP_ORIGINS: https://${DRIFT_DOMAIN}
      SELFSERVICE_FLOWS_ERROR_UI_URL: https://${DRIFT_DOMAIN}/error
      SELFSERVICE_FLOWS_LOGIN_UI_URL: https://${DRIFT_DOMAIN}/login
      SELFSERVICE_FLOWS_REGISTRATION_UI_URL: https://${DRIFT_DOMAIN}/setup
      SELFSERVICE_FLOWS_SETTINGS_UI_URL: https://${DRIFT_DOMAIN}/settings
      SELFSERVICE_FLOWS_LOGOUT_AFTER_DEFAULT_BROWSER_RETURN_URL: https://${DRIFT_DOMAIN}/login
```

- `drift` из compose выключен профилем: его заменяет `systemd`-сервис ниже (`podman-compose --profile container up` вернёт контейнер).
- `in_pod: false` — podman-compose по умолчанию собирает сервисы в pod, а pod несовместим с `userns_mode`.
- `restart: always` — после перезагрузки контейнеры поднимает `podman-restart.service` (`systemctl --user enable podman-restart.service`, нужен `loginctl enable-linger`); `unless-stopped` он не трогает.
- `healthcheck.interval: 30s` — каждая проверка podman запускает процесс через таймер systemd; `start_interval` podman не поддерживает, поэтому первая проверка ждёт полный интервал.
- Kratos без `--dev` (secure cookie, нужен HTTPS); URL из `kratos.yml` перекрываются переменными окружения.

`.env` (`chmod 600`; его читают и podman-compose, и `systemd`):

```dotenv
DRIFT_DOMAIN=drift.example.com

POSTGRES_USER=drift
POSTGRES_PASSWORD=<openssl rand -hex 24>
POSTGRES_DB=drift

KRATOS_SECRETS_COOKIE=<openssl rand -hex 32>
KRATOS_SECRETS_CIPHER=<openssl rand -hex 16>
KRATOS_PASSKEY_ENABLED=true

DRIFT_ADDR=127.0.0.1:8080
DRIFT_PUBLIC_URL=https://drift.example.com
DATABASE_URL=postgres://drift:<openssl rand -hex 24>@127.0.0.1:5432/drift?sslmode=disable
KRATOS_PUBLIC_URL=http://127.0.0.1:4433
KRATOS_ADMIN_URL=http://127.0.0.1:4434
DRIFT_WORKSPACE_ROOT=/home/user/work
DRIFT_CHATS_DIR=/home/user/drift/data/chats
DRIFT_AGENTS_DIR=/home/user/drift/data/agents
DRIFT_SHELL=fish
GOMEMLIMIT=192MiB

GIT_CONFIG_GLOBAL=/home/user/.gitconfig
XDG_RUNTIME_DIR=/run/user/1000
DOCKER_HOST=unix:///run/user/1000/podman/podman.sock
TESTCONTAINERS_RYUK_DISABLED=true
HTTP_PROXY=http://127.0.0.1:2080
HTTPS_PROXY=http://127.0.0.1:2080
http_proxy=http://127.0.0.1:2080
https_proxy=http://127.0.0.1:2080
NO_PROXY=localhost,127.0.0.1,::1
no_proxy=localhost,127.0.0.1,::1
```

`~/.config/systemd/user/drift.service`:

```ini
[Unit]
Description=Drift (web UI for coding agents)
After=network-online.target evx-gw.service podman-restart.service
Wants=podman-restart.service

[Service]
WorkingDirectory=%h/drift
EnvironmentFile=%h/drift/.env
Environment=PATH=%h/.local/bin:%h/.npm-global/bin:/usr/local/bin:/usr/bin:/bin
ExecStart=%h/drift/bin/drift
Restart=always
RestartSec=3
KillMode=mixed
TimeoutStopSec=20

[Install]
WantedBy=default.target
```

Запуск и обновление:

```sh
cd ~/drift && podman-compose up -d
systemctl --user enable --now drift.service
scp bin/drift server:drift/bin/drift.new && ssh server 'mv ~/drift/bin/drift.new ~/drift/bin/drift && systemctl --user restart drift'
```

CLI агентов ставятся на хост в `PATH` сервиса (`~/.local/bin`, `~/.npm-global/bin`), версии — из `ARG` в `Containerfile`: `claude`, `codex` и `agy` (архив `linux-x64-musl/cli_linux_x64_musl.tar.gz` из `antigravity-public`, сверить `AGY_SHA512_AMD64`, бинарник `antigravity` установить как `agy`).

Caddy: `reverse_proxy 127.0.0.1:8080` (SSE и WebSocket проходят без настроек).
