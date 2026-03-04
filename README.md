# Практическое занятие №2
## gRPC: создание простого микросервиса, вызовы методов

**ФИО:** Бакланова Е.С.
**Группа:** ЭФМО-01-25

## Цели работы

- Научиться работать с gRPC: описывать контракт в .proto файле.
- Реализовать gRPC-сервер в сервисе Auth и gRPC-клиента в сервисе Tasks.
- Научить сервисы общаться синхронно через gRPC вместо HTTP.
- Реализовать корректную обработку ошибок и использовать deadline/timeout при вызовах.

## Теория

### GRPC И PROTOCOL BUFFERS

**gRPC** —  это высокопроизводительный фреймворк для удаленного вызова процедур (RPC), разработанный Google. Он использует HTTP/2 для транспорта и Protocol Buffers (protobuf) в качестве языка описания интерфейсов и сериализации данных.

Преимущества gRPC:

- Высокая производительность за счет бинарной сериализации (protobuf) и мультиплексирования HTTP/2.
- Четкий контракт. Файл .proto является единственным источником о структурах данных и методах сервиса.
- Поддержка множества языков. Генерация кода клиента и сервера под любую платформу.
- Встроенная поддержка аутентификации, потоковой передачи данных, дедлайнов и отмены запросов.

### PROTO

Файл .proto выступает в роли контракта между клиентом и сервером. Он описывает:

1. Сервисы (Services) - какие RPC-методы доступны.
2. Сообщения (Messages) - структуры запросов и ответов.
Любое изменение в API должно начинаться с изменения этого файла.

### Содержание проекта

**Auth service** (порт 8081 +  gRPC порт 50051)
- Аутентификация пользователей
- Выдача токенов
- Проверка токенов через gRPC

**Tasks service** (порт 8082)
- CRUD для задач (TODO-список)
- Перед каждой операцией проверяет токен, отправляя gRPC-запрос в Auth service

### Эндпоинты

Auth

 | Метод | Эндпоинт | Описание | Тело запроса | Ответ |
 |-------|----------|----------|--------------|--------|
 | POST | /v1/auth/login | Получение токена | {"username":"student","password":"student"} | 200: {"access_token":"demo-token","token_type":"Bearer"} 400: Неверный JSON 401: Неверные данные |
 | GET | /v1/auth/verify | Проверка токена | Заголовок: Authorization: Bearer <token> | 200: {"valid":true,"subject":"student"} 401: {"valid":false,"error":"unauthorized"} |

Tasks

 | Метод | Эндпоинт | Описание | Тело запроса | Ответ |
 |-------|----------|----------|--------------|--------|
 | POST | /v1/tasks | Создать задачу | {"title":"...","description":"...","due_date":"..."} | 201: Задача создана 400: Неверные данные 401: Неавторизован |
 | GET | /v1/tasks | Все задачи | - | 200: Список задач 401: Неавторизован |
 | GET | /v1/tasks/{id} | Задача по ID | - | 200: Задача 404: Не найдена 401: Неавторизован |
 | PATCH | /v1/tasks/{id} | Обновить задачу | {"title":"...","done":true} | 200: Обновлено 404: Не найдена 401: Неавторизован |
 | DELETE | /v1/tasks/{id} | Удалить задачу | - | 204: Удалено 404: Не найдена 401: Неавторизован |

+  **504 Gateway Timeout, 503 Service Unavailable, если Auth недоступен**

### Структура

<img width="359" height="728" alt="image" src="https://github.com/user-attachments/assets/64e566f7-dc4f-4c82-9d9b-8a2c13819737" />

### Инструкция по запуску

1. Установка плагинов protoc
   
   - go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   - go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
     
3. Генерация кода из корня проекта
   
   - gen/go/auth/auth.pb.go (структуры данных)
   - gen/go/auth/auth_grpc.pb.go (интерфейсы клиента и сервера)
     
4.  Запуск Auth

  - cd services/auth
  - $env:AUTH_PORT="8081"
  - go run ./cmd/auth
  
2. Запуск Tasks

  - cd services/tasks
  - $env:TASKS_PORT="8082"
  - $env:AUTH_GRPC_ADDR="localhost:50051"
  - go run ./cmd/tasks

### Тестирование

1. Получить токен

```bash
curl -s -X POST http://localhost:8081/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"student","password":"student"}'
```

<img width="900" height="472" alt="image" src="https://github.com/user-attachments/assets/48fdbf9d-e237-4ee7-a2b9-796dd637fb1d" />

2. Создать задачу через Tasks (с проверкой gRPC)

```bash
  curl -i -X POST http://localhost:8082/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer demo-token" \
  -d '{"title":"Do PZ18","description":"split services","due_date":"2026-01-10"}'
```

<img width="900" height="443" alt="image" src="https://github.com/user-attachments/assets/7d7a57e4-2525-402e-b93b-de33936f3132" />

Лог

<img width="900" height="260" alt="image" src="https://github.com/user-attachments/assets/3f09ec99-0d26-4f4f-aad0-0f305e1d3bf7" />

3. Получить список задач

```bash
  curl -X GET http://localhost:8082/v1/tasks \
  -H "Authorization: Bearer demo-token"
```

<img width="900" height="369" alt="image" src="https://github.com/user-attachments/assets/c9aa114f-d038-47ec-b724-307b9a59e60b" />

<img width="900" height="185" alt="image" src="https://github.com/user-attachments/assets/a90d9880-1efe-40ad-b6ca-c47152e71c2a" />


4. Попробовать без токена

```bash
  curl -i http://localhost:8082/v1/tasks 
```

<img width="900" height="484" alt="image" src="https://github.com/user-attachments/assets/c59e7873-a2a1-45b1-967e-3fe2d4b3077e" />


<img width="974" height="481" alt="image" src="https://github.com/user-attachments/assets/8b223143-ca77-4434-b673-4d423f7543c6" />

5. После остановке сервиса 

<img width="900" height="415" alt="image" src="https://github.com/user-attachments/assets/ccfb65a6-fb26-4a39-afd0-0f692a3c3065" />

<img width="900" height="97" alt="image" src="https://github.com/user-attachments/assets/6e32300d-4efc-4d12-a360-05fe61131c25" />


6. Проверка request-id

<img width="900" height="555" alt="image" src="https://github.com/user-attachments/assets/d664dac8-7a35-4946-8b43-f5cd63fbe1ec" />

<img width="900" height="172" alt="image" src="https://github.com/user-attachments/assets/766139d0-baa4-4404-ab3f-03d0310349e3" />


### Контрольные вопросы

1.	Что такое .proto и почему он считается контрактом?

файл, в котором описываются сервисы и структуры данных для gRPC. Он считается контрактом, потому что и клиент, и сервер строго следуют этому описанию

2. Что такое deadline в gRPC и чем он полезен?

максимальное время ожидания ответа от сервера. Полезен тем, что, если сервер упал или тормозит — по истечении времени приходит ошибка

3. Почему "exactly-once" не даётся просто так даже в RPC?

Из-за ненадёжности сети подтверждение о выполнении может потеряться, и клиент повторит запрос

4. Как обеспечивать совместимость при расширении .proto?

- Не менять числовые теги у существующих полей
- Добавлять новые поля с новыми тегами
- Не удалять старые поля без reserved
- Старые клиенты не видят новые поля 
