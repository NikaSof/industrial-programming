# Проверка API

Команды предназначены для PowerShell. Ниже приведены восемь тестовых запросов и ожидаемые результаты; фактические результаты фиксируются на скриншотах.

## Подготовка

В первом терминале запустите сервер из папки `practice-03`:

```powershell
cd E:\study\MAGA\Ind_prog\practice-03
go run ./cmd/server
```

Выполняйте запросы во втором терминале по порядку. Начните с заново запущенного сервера и не перезапускайте его между запросами: задачи хранятся в памяти. При таком выполнении первые две задачи получат ID 1 и 2.

Используется `curl.exe`, чтобы не вызвать одноимённый псевдоним PowerShell. Параметр `-i` показывает статус и заголовки. В POST-запросах JSON передаётся через стандартный ввод: `--data-binary '@-'` позволяет избежать различий обработки кавычек в версиях PowerShell. Примеры содержат ASCII, чтобы не зависеть от кодировки стандартного ввода.

## 1. Состояние сервера

```powershell
curl.exe -i http://localhost:8080/health
```

Ожидается `200 OK`, `Content-Type: application/json` и тело:

```json
{"status":"ok"}
```

## 2. Создание первой задачи

```powershell
'{"title":"Learn Go"}' | curl.exe -i -X POST http://localhost:8080/tasks -H "Content-Type: application/json" --data-binary '@-'
```

Ожидается `201 Created` и тело:

```json
{"id":1,"title":"Learn Go","done":false}
```

## 3. Создание второй задачи

```powershell
'{"title":"Read a book"}' | curl.exe -i -X POST http://localhost:8080/tasks -H "Content-Type: application/json" --data-binary '@-'
```

Ожидается `201 Created` и тело:

```json
{"id":2,"title":"Read a book","done":false}
```

## 4. Получение списка

```powershell
curl.exe -i http://localhost:8080/tasks
```

Ожидается `200 OK` и массив с обеими задачами. Порядок элементов может отличаться:

```json
[{"id":1,"title":"Learn Go","done":false},{"id":2,"title":"Read a book","done":false}]
```

## 5. Фильтрация по названию

```powershell
curl.exe -i "http://localhost:8080/tasks?q=go"
```

Ожидается `200 OK` и только задача `Learn Go`. Поиск не учитывает регистр:

```json
[{"id":1,"title":"Learn Go","done":false}]
```

## 6. Получение задачи по ID

```powershell
curl.exe -i http://localhost:8080/tasks/1
```

Ожидается `200 OK` и один объект, а не массив:

```json
{"id":1,"title":"Learn Go","done":false}
```

Если задачи создавались ранее без перезапуска сервера, подставьте фактический ID из ответа POST.

## 7. Некорректный ID

```powershell
curl.exe -i http://localhost:8080/tasks/abc
```

Ожидается `400 Bad Request`:

```json
{"error":"invalid id"}
```

## 8. Несуществующая задача

```powershell
curl.exe -i http://localhost:8080/tasks/9999
```

Ожидается `404 Not Found`, если задачи с таким ID нет:

```json
{"error":"task not found"}
```

Все перечисленные ответы должны иметь `Content-Type: application/json`. В терминале сервера после каждого запроса должны появляться метод, путь, статус и длительность обработки. Query-параметры в текущем формате логирования не выводятся.

## Скриншоты

Сохраняйте изображения в `screenshots/`. На снимках запросов должны быть видны команда, HTTP-статус, Content-Type и тело ответа.

| Файл | Содержание |
|---|---|
| `health-response.png` | Запрос 1: состояние сервера |
| `create-tasks.png` | Запросы 2 и 3: создание двух задач, статусы 201 и выданные ID |
| `list-filter.png` | Запросы 4 и 5: полный список и результат поиска |
| `get-task.png` | Запрос 6: получение одной задачи |
| `errors.png` | Запросы 7 и 8: ответы 400 и 404 |
| `server-logs.png` | Терминал сервера после запросов: метод, путь, статус, длительность |
| `project-structure.png` | Дерево файлов проекта в VS Code |

Если два ответа не помещаются на одном снимке, сохраните их отдельными файлами без уменьшения текста до нечитаемого размера.