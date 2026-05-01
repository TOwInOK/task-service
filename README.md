# Task-service

Сервис - Task tracker

## Project structure
```
task-service/
├── cmd/
│   └── server/
│       └── main.go             # Точка входа. Только запуск сервера.
├── internal/                   # Весь приватный код приложения (не импортируется извне)
│   ├── config/
│   │   └── config.go           # atomic.Value для конфигурации
│   ├── storage/
│   │   └── actor.go            # Актор (горутина + каналы) для работы с задачами
│   └── http/
│        ├── handler.go      # Ручки (handlers), работают с HTTP
│        └── router.go       # Сборка роутера chi
├── configs/
│   └── local.yaml              # Файл с настройками
├── go.mod
└── go.sum
```

## Technlogy
- web: htmlx
- backend: go
