# 🔐 GophKeeper

**GophKeeper** — клиент-серверная система для безопасного хранения конфиденциальных данных. Это защищённый менеджер паролей и чувствительной информации.

## 📋 Описание проекта
- **Язык:** Go 1.25+
- **База данных:** PostgreSQL 16+
- **Архитектура:** Client-Server с REST API

Система предназначена для централизованного хранения:
- Логинов и паролей
- Банковских карт
- Текстовых файлов
- Бинарных данных (файлы, ключи, сертификаты)

## ✨ Возможности

### Сервер
- REST API для управления элементами хранилища
- JWT-based аутентификация с настраиваемым временем жизни токенов
- AES-256-GCM шифрование данных на уровне сервера
- PostgreSQL для надёжного хранения данных
- Автоматическая миграция базы данных
- Поддержка TLS/HTTPS
- Структурированное логирование (zap)
- Health checks и graceful shutdown

### Клиент
- CLI интерфейс для всех операций
- Регистрация и аутентификация пользователей
- CRUD операции для всех типов данных
- Загрузка секретных данных как plain text (`--data`) или из файла (`--file`)
- Локальное кэширование для offline работы
- Поддержка небезопасных TLS соединений (для разработки)

## 🔒 Безопасность

- **Хеширование паролей:** bcrypt с cost factor 10
- **Шифрование данных:** AES-256-GCM с уникальными nonce
- **Аутентификация:** JWT токены с подписью HMAC-SHA256
- **TLS/HTTPS:** Поддержка защищённых соединений
- **Защита от SQL injection:** Подготовленные запросы (pgx)

## 🏗️ Архитектура

Проект построен на принципах:

- **Слоистая архитектура**: handlers, services, repositories
- **Dependency Injection**: слабая связанность компонентов
- **Interface Segregation**: использование интерфейсов для абстракции
- **RESTful API**: стандартизированные эндпоинты с HTTP методами
- **Stateless Authentication**: JWT токены без состояния на сервере
- **Configuration Management**: конфигурация через переменные окружения и флаги

### Структура проекта
```
gophkeeper/
├── cmd/                         # Точки входа
│   ├── client/                  # CLI клиент
│   └── server/                  # HTTP сервер
├── internal/                    # Внутренняя логика
│   ├── client/                  # Клиентская часть
│   │   ├── app/                 # CLI приложение (cobra commands)
│   │   ├── config/              # Конфигурация клиента
│   │   ├── repositories/        # Локальный кэш
│   │   └── services/            # API клиент
│   └── server/                  # Серверная часть
│       ├── app/                 # Инициализация приложения
│       ├── config/              # Конфигурация сервера
│       ├── handlers/            # HTTP хендлеры
│       ├── middleware/          # HTTP middleware
│       ├── repositories/        # Работа с БД
│       └── services/            # Бизнес-логика
├── models/                      # Общие модели данных
├── pkg/                         # Переиспользуемые пакеты
│   ├── crypto/                  # AES-256 шифрование
│   ├── jwt/                     # JWT утилиты
│   └── logger/                  # Структурированное логирование (zap)
└── internal/server/app/migrations/  # SQL миграции БД
```

## 🚀 Быстрый старт

### Генерация ключей

```bash
# Сгенерировать MASTER_KEY для шифрования
openssl rand -base64 32

# Сгенерировать JWT_SECRET
openssl rand -base64 64
```

### Сервер: Docker Compose (рекомендуется)

```bash
# Скопировать и настроить переменные окружения
cp .env.example .env

# Автоматически сгенерировать и установить секретные ключи
sed -i.bak "s|JWT_SECRET=.*|JWT_SECRET=$(openssl rand -base64 64 | tr -d '\n')|" .env
sed -i.bak "s|MASTER_KEY=.*|MASTER_KEY=$(openssl rand -base64 32 | tr -d '\n')|" .env
sed -i.bak "s|POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD=$(openssl rand -base64 16 | tr -d '\n')|" .env
rm .env.bak

# Запустить весь стек (PostgreSQL + Server)
docker-compose up -d

# Проверить логи
docker-compose logs -f server
```

Сервер будет доступен на `http://localhost:8080`

### Клиент: Локальная сборка (рекомендуется)

Для клиента рекомендуется локальная сборка, так как это интерактивная CLI-утилита:

```bash
# Установить зависимости
go mod download

# Собрать клиент
go build -o bin/client ./cmd/client

# Или с версией
go build -ldflags "-X main.buildVersion=1.0.0 -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o bin/client ./cmd/client

# Использовать
./bin/client register --username alice --password secret123
./bin/client login --username alice --password secret123
./bin/client list
```

### Локальная сборка (полная)

```bash
# Установить зависимости
go mod download

# Собрать сервер
go build -o bin/server ./cmd/server

# Собрать клиент
go build -o bin/client ./cmd/client

# Собрать оба с версией
go build -ldflags "-X main.buildVersion=1.0.0 -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o bin/server ./cmd/server

go build -ldflags "-X main.buildVersion=1.0.0 -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o bin/client ./cmd/client
```

### Docker сборка сервера

```bash
# Собрать образ сервера
docker build -t gophkeeper-server .

# Запустить сервер
docker run -p 8080:8080 --env-file .env gophkeeper-server
```

## ⚙️ Конфигурация

### Сервер

#### Переменные окружения

| Переменная | Флаг | Описание | По умолчанию | Обязательна |
|-----------|------|----------|--------------|-------------|
| `LOG_LEVEL` | `-l` | Уровень логирования (debug, info, warn, error) | `info` | Нет |
| `SERVER_ADDR` | `-a` | Адрес для прослушивания | `localhost:8080` | Нет |
| `DATABASE_DSN` | `-d` | PostgreSQL connection string | - | **Да** |
| `JWT_SECRET` | `--jwt-secret` | Секретный ключ для JWT | - | **Да** |
| `JWT_EXPIRATION` | `--jwt-exp` | Время жизни JWT токена | `24h` | Нет |
| `MASTER_KEY` | `--master-key` | Base64 ключ для AES-256 шифрования | - | **Да** |
| `TLS_CERT_FILE` | `--tls-cert` | Путь к TLS сертификату | - | Нет |
| `TLS_KEY_FILE` | `--tls-key` | Путь к TLS ключу | - | Нет |

#### Примеры запуска сервера

**Минимальная конфигурация:**
```bash
./server \
  -d "postgres://user:pass@localhost:5432/gophkeeper?sslmode=disable" \
  --jwt-secret "my-secret-key-at-least-32-characters-long" \
  --master-key "$(openssl rand -base64 32)"
```

**С кастомным адресом и портом:**
```bash
./server \
  -a "0.0.0.0:9090" \
  -d "postgres://user:pass@localhost:5432/gophkeeper?sslmode=disable" \
  --jwt-secret "my-secret-key" \
  --master-key "$(openssl rand -base64 32)"
```

**С debug логами и коротким временем токена:**
```bash
./server \
  -l debug \
  -a "localhost:8080" \
  -d "postgres://user:pass@localhost:5432/gophkeeper?sslmode=disable" \
  --jwt-secret "my-secret-key" \
  --jwt-exp 1h \
  --master-key "$(openssl rand -base64 32)"
```

**Полная конфигурация с TLS:**
```bash
./server \
  -l info \
  -a "0.0.0.0:8443" \
  -d "postgres://gophkeeper:secure_pwd@db.example.com:5432/gophkeeper?sslmode=require" \
  --jwt-secret "super-secure-jwt-secret-key-with-random-chars-$(openssl rand -hex 16)" \
  --jwt-exp 24h \
  --master-key "$(openssl rand -base64 32)" \
  --tls-cert "/etc/ssl/certs/server.crt" \
  --tls-key "/etc/ssl/private/server.key"
```

**Через переменные окружения:**
```bash
export LOG_LEVEL=info
export SERVER_ADDR=0.0.0.0:8080
export DATABASE_DSN="postgres://user:pass@localhost:5432/gophkeeper?sslmode=disable"
export JWT_SECRET="my-secret-key"
export JWT_EXPIRATION=24h
export MASTER_KEY="$(openssl rand -base64 32)"

./server
```

### Клиент

#### Переменные окружения

| Переменная | Флаг | Описание | По умолчанию | Обязательна |
|-----------|------|----------|--------------|-------------|
| `SERVER_ADDR` | `-a` | Адрес GophKeeper сервера | `http://localhost:8080` | Нет |
| `LOG_LEVEL` | `-l` | Уровень логирования | `info` | Нет |
| `TLS_INSECURE` | `-v` | Отключить проверку TLS сертификата | `false` | Нет |
| `CACHE_PATH` | `-c` | Путь к файлу кэша | `./cache.json` | Нет |
| `TOKEN_PATH` | `-t` | Путь к файлу с JWT токеном | `./token` | Нет |

#### Команды клиента

**Аутентификация:**
```bash
# Регистрация нового пользователя
gophkeeper register --username <имя> --password <пароль>

# Вход существующего пользователя
gophkeeper login --username <имя> --password <пароль>
```

**Создание элементов:**
```bash
# Синтаксис: create --type <тип> --title <название> [--file <путь> | --data <текст> | --meta <метаданные>]
# Типы: credentials | text | card | binary

# Учетные данные из файла
gophkeeper create --type credentials --title "GitHub" --file data.json

# Учетные данные из текста (JSON)
gophkeeper create --type credentials --title "AWS" --data '{"access_key":"AKIA...","secret_key":"..."}'

# Текстовая заметка с метаданными
gophkeeper create --type text --title "Notes" --meta "My secret notes"

# Текстовая заметка с данными
gophkeeper create --type text --title "Secret" --data "My important secret text"

# Банковская карта из файла
gophkeeper create --type card --title "Visa" --file card-data.json

# Банковская карта из текста (JSON)
gophkeeper create --type card --title "MasterCard" --data '{"number":"5555...","holder":"John","cvv":"123","expiry":"12/25"}'

# Бинарные данные из файла
gophkeeper create --type binary --title "SSH Key" --file ~/.ssh/id_rsa
```

**Управление элементами:**
```bash
# Список всех элементов
gophkeeper list

# Получить конкретный элемент по ID (вывод в stdout)
gophkeeper get --id <uuid>

# Получить элемент и сохранить в файл
gophkeeper get --id <uuid> > secret.json

# Обновить элемент (опционально: --type, --title, --meta, --file, --data)
gophkeeper update --id <uuid> --title "New Title" --file new-data.json
gophkeeper update --id <uuid> --data '{"username":"new","password":"secret"}'

# Удалить элемент
gophkeeper delete --id <uuid>
```

**Примеры получения данных:**
```bash
# Вывод данных в stdout (по умолчанию)
$ gophkeeper get --id 123e4567-e89b-12d3-a456-426614174000
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "type": "credentials",
  "title": "GitHub",
  "data": "{\"username\":\"alice\",\"password\":\"secret123\"}",
  "metadata": "",
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}

# Сохранить данные в файл
$ gophkeeper get --id 123e4567-e89b-12d3-a456-426614174000 > github-creds.json
$ cat github-creds.json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "type": "credentials",
  "title": "GitHub",
  "data": "{\"username\":\"alice\",\"password\":\"secret123\"}",
  ...
}
```

**Прочее:**
```bash
# Показать версию клиента
gophkeeper version
```

#### Примеры запуска клиента

**Минимальная конфигурация (локальный сервер):**
```bash
# Регистрация
./client register --username alice --password mypassword

# Вход
./client login --username alice --password mypassword

# Создание элемента через --data (без файла)
./client create --type credentials --title "Email" --data '{"username":"alice@example.com","password":"secret"}'

# Или текст с метаданными
./client create --type text --title "Secret Note" --meta "My important note"

# Или из файла
echo '{"username":"alice@example.com","password":"secret"}' > creds.json
./client create --type credentials --title "Email" --file creds.json
```

**С кастомным сервером:**
```bash
./client -a "https://gophkeeper.example.com" \
  login --username alice --password mypassword
```

**С отключенной проверкой TLS (для самоподписанных сертификатов):**
```bash
./client -a "https://localhost:8443" -v \
  login --username alice --password mypassword
```

**С debug логами и кастомными путями:**
```bash
./client \
  -l debug \
  -a "http://localhost:8080" \
  -c "/tmp/my-cache.json" \
  -t "/tmp/my-token" \
  list
```

**Полная конфигурация:**
```bash
# Использовать --data для передачи JSON напрямую
./client \
  -l info \
  -a "https://api.gophkeeper.com:8443" \
  -c "$HOME/.config/gophkeeper/cache.json" \
  -t "$HOME/.config/gophkeeper/token" \
  create --type credentials --title "AWS" --data '{"access_key":"AKIA...","secret_key":"wJal..."}'

# Или из файла
echo '{"access_key":"AKIA...","secret_key":"wJal..."}' > aws-creds.json
./client \
  -l info \
  -a "https://api.gophkeeper.com:8443" \
  -c "$HOME/.config/gophkeeper/cache.json" \
  -t "$HOME/.config/gophkeeper/token" \
  create --type credentials --title "AWS" --file aws-creds.json
```

**Через переменные окружения:**
```bash
export SERVER_ADDR=https://api.example.com
export LOG_LEVEL=debug
export TLS_INSECURE=true
export CACHE_PATH=$HOME/.gophkeeper/cache.json
export TOKEN_PATH=$HOME/.gophkeeper/token

./client login --username alice --password secret
./client list
```

## 🧪 Тестирование

```bash
# Запустить все тесты
go test ./...

# С покрытием
go test -cover ./...

# Подробный вывод
go test -v ./...

# Тесты конкретного пакета
go test ./internal/server/handlers/...
```

## 📝 Лицензия

Проект создан в образовательных целях.

## 👤 Автор

Pro100x3mal
