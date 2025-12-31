# Test gRPC Server

Тестовый gRPC сервер для проверки рефлекта.

## Запуск

Требуется установленный `protoc` и плагины:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Затем:

```bash
make run
```

Или вручную:

```bash
make build
./testserver
```

Сервер запустится на порту `:50051` с включенной рефлексией.
