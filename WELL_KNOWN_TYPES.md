# Well-Known Types Support

grpc-gui поддерживает автоматическое распознавание и правильную генерацию значений для стандартных типов Protocol Buffers.

## Поддерживаемые типы

### google.protobuf.Timestamp

**Формат JSON:** RFC3339 строка

**Примеры:**
```json
"2026-02-05T14:05:47Z"
"2024-12-31T23:59:59.999Z"
"2025-01-01T00:00:00+03:00"
```

**UI виджет:** Date/Time picker с кнопками "Now" и "Apply"

### google.protobuf.Duration

**Формат JSON:** Строка с числом и единицей измерения

**Примеры:**
```json
"1.5s"     // 1.5 секунды
"300s"     // 5 минут (300 секунд)
"0.1s"     // 100 миллисекунд
"3600s"    // 1 час
```

**UI виджет:** Ввод с выбором единицы измерения (s/m/h) и набором preset значений

### google.protobuf.Empty

**Формат JSON:** Пустой объект

**Пример:**
```json
{}
```

### google.protobuf.Struct

**Формат JSON:** Обычный JSON объект

**Пример:**
```json
{
  "key": "value",
  "nested": {
    "field": 123
  }
}
```

### google.protobuf.Value

**Формат JSON:** Любое JSON значение (null, число, строка, bool, объект, массив)

**Примеры:**
```json
null
"string value"
123
true
{"key": "value"}
[1, 2, 3]
```

### google.protobuf.ListValue

**Формат JSON:** JSON массив

**Пример:**
```json
[1, "two", true, null]
```

### google.protobuf.Any

**Формат JSON:** Объект с полем `@type`

**Пример:**
```json
{
  "@type": "type.googleapis.com/package.MessageName",
  "field": "value"
}
```

## Как это работает

1. **Определение типа:** При reflection grpc-gui автоматически определяет, является ли поле well-known типом
2. **Генерация примера:** Для каждого типа генерируется корректное значение по умолчанию
3. **UI виджеты:** В редакторе JSON рядом с полями well-known типов появляются специальные иконки:
   - Часы для Timestamp
   - Песочные часы для Duration
   - Круг с вопросом для Enum

4. **Валидация:** gRPC сервер автоматически проверяет корректность формата при получении запроса

## Техническая документация

### RFC3339 формат для Timestamp

RFC3339 - это стандарт ISO 8601 с обязательным указанием timezone:

```
YYYY-MM-DDTHH:MM:SS.sssZ
YYYY-MM-DDTHH:MM:SS.sss+HH:MM
```

Где:
- `T` - разделитель между датой и временем
- `Z` - UTC timezone (или можно указать offset типа `+03:00`)
- `.sss` - опциональные миллисекунды/наносекунды

### Duration формат

Формат: число с плавающей точкой + единица измерения

Поддерживаемые единицы:
- `s` - секунды (по умолчанию)
- `m` - минуты (конвертируется в секунды)
- `h` - часы (конвертируется в секунды)

Внутри protobuf Duration хранится как:
```protobuf
message Duration {
  int64 seconds = 1;
  int32 nanos = 2;
}
```

Но в JSON всегда используется строковый формат.

## Ссылки

- [Protocol Buffers JSON Mapping](https://protobuf.dev/programming-guides/proto3/#json)
- [google.protobuf.Timestamp](https://protobuf.dev/reference/protobuf/google.protobuf/#timestamp)
- [google.protobuf.Duration](https://protobuf.dev/reference/protobuf/google.protobuf/#duration)
- [RFC3339 Specification](https://www.rfc-editor.org/rfc/rfc3339)
