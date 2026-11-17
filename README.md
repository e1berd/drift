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

## Запуск

```sh
cp .env.example .env        # смените пароли и секреты Kratos
task up                     # Postgres + Kratos (БД kratos создаётся и мигрируется автоматически)
task web:install
task dev                    # Vite :5173 + Go (air, тег dev) :8080 → открывать http://localhost:8080
```

В dev-сборке (`-tags dev`) Go проксирует фронтенд на Vite вместе с HMR. Release:

```sh
task build                  # vite build → go build → bin/drift
```

Бинарнику нужны только переменные окружения из `.env.example` и запущенный Kratos.

## Продакшен

`deploy/kratos/kratos.yml` рассчитан на `http://localhost:8080` и запускается с `--dev`. Для боевого домена поменяйте `base_url`, все `ui_url`, `allowed_return_urls`, `passkey.config.rp.id/origins`, уберите `--dev` (нужен HTTPS) и задайте `DRIFT_PUBLIC_URL`.
