# ai-memory

[English version](README_EN.md)

Go-пакет для хранения структурированных воспоминаний (нейронов) с лимитом активной памяти, связями между записями и персистентностью через JSON.

## Идея

Память делится на два слоя:

- **Active** — полные записи, которые живут в оперативной памяти. Количество ограничено `maxNeurons`.
- **Crumbs** — «крошки»: вытесненные записи без поля `Body`. Метаданные сохраняются, тяжёлый контент — нет.

Когда активный слой заполняется, старейший нейрон автоматически вытесняется в крошки.

## Установка

```bash
go get https://github.com/junhekdevsru/ai-memory/
```

## Быстрый старт

```go
import "github.com/junhekdevsru/ai-memory"

// Создать регион с лимитом 100 активных нейронов
r, err := memory.NewRegion(100)

// Добавить нейрон
err = r.Add(memory.Neuron{
    ID:          "task-42",
    Title:       "Рефакторинг auth",
    Theme:       "backend",
    TaskName:    "AUTH-42",
    Description: "Вынесли middleware в отдельный пакет",
    Body:        "Полный контекст задачи...",
})

// Найти по названию и теме
neuron, loc := r.Lookup("Рефакторинг auth", "backend")
// loc == memory.Active или memory.Crumbs или memory.NotFound

// Связать два нейрона
err = r.Link("task-42", "task-99")

// Вытеснить вручную
err = r.Evict("task-42")

// Сохранить на диск / загрузить с диска
err = r.Save("memory.json")
r, err = memory.LoadRegion("memory.json")
```

## API

### Типы

```go
type Neuron struct {
    ID          string
    Title       string
    Theme       string
    TaskName    string
    Description string
    Body        string
    CreatedAt   time.Time
    LastSeen    time.Time
}

type Edge struct {
    A, B string // ID нейронов
}

type Location int // NotFound | Active | Crumbs
```

### Region

| Метод | Описание |
|---|---|
| `NewRegion(maxNeurons int) (*Region, error)` | Создать регион. `maxNeurons` ≥ 1 |
| `Add(n Neuron) error` | Добавить нейрон. При заполнении — вытесняет первый |
| `Lookup(title, theme string) (*Neuron, Location)` | Найти нейрон по заголовку и теме |
| `Link(a, b string) error` | Создать связь между двумя нейронами |
| `Evict(id string) error` | Вытеснить нейрон в крошки вручную |
| `Active() []Neuron` | Снимок активного слоя |
| `Crumbs() []Neuron` | Снимок крошек |
| `Save(path string) error` | Сохранить в JSON-файл |
| `LoadRegion(path string) (*Region, error)` | Загрузить из JSON-файла |

### Ошибки

| Ошибка | Когда |
|---|---|
| `ErrInvalidMaxNeurons` | `maxNeurons` < 1 |
| `ErrNeuronNotFound` | Нейрон с таким ID не найден |
| `ErrDuplicateID` | Нейрон с таким ID уже существует |
| `ErrSelfLink` | Попытка связать нейрон с самим собой |

## Структура пакета

| Файл | Содержимое |
|---|---|
| `neuron.go` | Типы данных: `Neuron`, `Edge`, `Location` |
| `region.go` | Сущность `Region` и вся логика |
| `region_io.go` | `Save` / `LoadRegion` — персистентность |
| `errors.go` | Сентинел-ошибки |

## Потокобезопасность

Все методы `Region` защищены `sync.RWMutex`. Безопасно использовать из нескольких горутин.
