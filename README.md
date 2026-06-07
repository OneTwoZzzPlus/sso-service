# Сервис авторизации

Учебный pet-проект на go и typescript next.js

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![JWT](https://img.shields.io/badge/JWT-black?style=for-the-badge&logo=JSON%20web%20tokens&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)
![gRPC](https://img.shields.io/badge/gRPC-244c5a?style=for-the-badge&logo=grpc&logoColor=white)
![Next.js](https://img.shields.io/badge/Next.js-black?style=for-the-badge&logo=nextdotjs&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white)

## Запуск
- Сборка и запуск `docker-compose up --build`
- Запуск `docker compose up`

## Микросервисы
|     Имя     | Технология | Источник  | Описание |
| ----------- | ---------- | --------- | -------- |
| db          | PostgreSQL | DockerHub | Хранилище - реляционная база данных
| sso-service | Go         | self      | Сервис для регистрации, авторизации и управления правами
| front       | TS Next.js | self      | Интерфейс пользователей

## sso-service
Сервис регистрации и авторизации.

- auth *service*
- gRPC server *app*
- REST gateway *app*
- email provider (smtp) *app*
- storage *app*

Запуск sso-service (необходима запущенная БД).
```powershell
cd ./sso-service
$env:CONFIG_PATH="./config/local.yml
go run .
```

## front
Простые страницы с формами ввода и кнопками, отправляющими запросы GRPC-web клиента.

Запуск front веб-сервиса.ы
```powershell
cd ./front
yarn dev
```

## protos
Автогенерация PROTOBUF файлов.
```powershell
cd ./protos
./generate.ps1
```

## TODO
- Create register page (front)
- Gorutine email
- Add confirmation of mail
- ...
