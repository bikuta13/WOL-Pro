# Руководство разработчика

Это руководство предназначено для разработчиков, желающих внести вклад в проект WOL Pro (Go + Wails версия) или расширить функциональность.

## Архитектура проекта

```
wol-pro/
├── main.go                          # Точка входа приложения
├── internal/
│   ├── app/
│   │   └── app.go                  # Основная логика приложения (API методы)
│   ├── device/
│   │   └── manager.go              # Управление устройствами (JSON хранилище)
│   └── wol/
│       └── magic_packet.go         # WOL функциональность (отправка пакетов)
├── frontend/
│   └── dist/
│       ├── index.html              # HTML интерфейс пользователя
│       └── app.js                  # JavaScript логика и взаимодействие с бэкендом
├── go.mod                          # Зависимости Go
├── go.sum                          # Checksum зависимостей
├── build.sh                        # Скрипт сборки
└── README.md, BUILDING.md и т.д.   # Документация
```

## Структура кода

### `internal/wol/magic_packet.go`
Содержит низкоуровневую функциональность WOL:
- `SendMagicPacket()` - основная функция отправки пакета на целевое устройство
- `createMagicPacket()` - создает магический пакет (6x0xFF + 16 повторений MAC)
- `ValidateMAC()` - проверяет корректность MAC адреса
- `NormalizeMAC()` - нормализует MAC адрес к формату XX:XX:XX:XX:XX:XX

### `internal/device/manager.go`
Управление хранилищем устройств:
- `Device` - структура для представления устройства
- `Manager` - класс для работы с JSON хранилищем
- Методы: `Add()`, `Get()`, `GetAll()`, `Update()`, `Delete()`, `Load()`, `Save()`
- Использует threads-safe доступ с sync.RWMutex

### `internal/app/app.go`
Основная логика приложения - API методы для фронтенда:
- `AppInstance` - основная структура приложения
- `GetDevices()` - получить все устройства
- `AddDevice()` - добавить новое устройство
- `UpdateDevice()` - обновить существующее устройство
- `DeleteDevice()` - удалить устройство
- `SendWOL()` - отправить WOL пакет

### `frontend/dist/index.html`
HTML структура интерфейса:
- Таблица устройств
- Кнопки управления (Wake, Add, Edit, Delete)
- Модальное окно для добавления/редактирования
- Статус-бар

### `frontend/dist/app.js`
JavaScript логика:
- `WOLApp` класс - управляет интерфейсом
- `window.go.main.AppInstance` - вызов Go методов из JS
- Обработчики событий кнопок
- Управление таблицей и модальным окном

## Установка для разработки

### Требования
- Go 1.21+
- Wails 2.6+
- Git

### Шаги установки

```bash
# 1. Клонирование репозитория
git clone https://github.com/yourusername/wol-pro.git
cd wol-pro

# 2. Загрузка зависимостей Go
go mod tidy

# 3. Запуск в режиме разработки (с горячей перезагрузкой)
wails dev

# Приложение откроется с автоматической перезагрузкой при изменении кода
```

## Разработка

### Режим разработки

```bash
wails dev
```

Это команда:
- Запускает Go бэкенд с поддержкой горячей перезагрузки
- Открывает фронтенд в окне браузера на `http://localhost:34115`
- Автоматически перезагружает при изменении файлов

### Структура команд

При разработке можно менять:
- Go файлы в `internal/` - приложение автоматически перекомпилируется
- HTML/CSS/JS файлы в `frontend/dist/` - страница автоматически перезагружается

### Отладка

**Go код**:
```bash
# Логирование в консоль
log.Println("Debug message")

# Вывидит в консоль разработчика при wails dev
log.Printf("Value: %v\n", someValue)
```

**JavaScript код**:
```javascript
// Логирование в консоль браузера
console.log("Debug message");
console.log("Value:", someValue);

// Открыть DevTools: Cmd+Alt+I (macOS) или Ctrl+Alt+I (Windows/Linux)
```

## Расширение функциональности

### Добавление новой WOL функции

1. **Добавьте функцию в `internal/wol/magic_packet.go`**:
```go
// NewWOLFeature реализует новую функцию WOL
func NewWOLFeature(macAddr string) (bool, string) {
    // Реализация
    return true, "Выполнено"
}
```

2. **Добавьте метод в `internal/app/app.go`** для доступа из фронтенда:
```go
// CallNewWOLFeature вызывает новую функцию
func (a *AppInstance) CallNewWOLFeature(macAddr string) (bool, string) {
    return wol.NewWOLFeature(macAddr)
}
```

3. **Используйте в `frontend/dist/app.js`**:
```javascript
// В JS классе WOLApp
async onNewFeature() {
    const [success, msg] = await window.go.main.AppInstance.CallNewWOLFeature(mac);
    if (success) {
        this.showSuccess(msg);
    } else {
        this.showError(msg);
    }
}
```

### Добавление новой операции с устройивами

1. **Добавьте метод в `internal/device/manager.go`**:
```go
// NewOperation выполняет новую операцию с устройством
func (m *Manager) NewOperation(id string) error {
    m.mu.RLock()
    device, exists := m.devices[id]
    m.mu.RUnlock()

    if !exists {
        return fmt.Errorf("устройство не найдено")
    }

    // Реализация операции
    return nil
}
```

2. **Добавьте API метод в `internal/app/app.go`**:
```go
func (a *AppInstance) DoOperation(id string) string {
    if err := a.manager.NewOperation(id); err != nil {
        return fmt.Sprintf("Ошибка: %v", err)
    }
    return "✅ Операция выполнена"
}
```

3. **Добавьте кнопку и обработчик в `frontend/dist/index.html` и `app.js`**:
```html
<button class="btn-custom" id="btnNewOp">🎯 Новая операция</button>
```

```javascript
document.getElementById('btnNewOp').addEventListener('click', () => this.onNewOp());

async onNewOp() {
    if (!this.selectedDeviceId) {
        this.showError('Выберите устройство');
        return;
    }
    const msg = await window.go.main.AppInstance.DoOperation(this.selectedDeviceId);
    this.showSuccess(msg);
}
```

## Соглашения кодирования

### Go код
- Следуйте `go fmt` (используйте `gofmt`)
- Используйте `golint` для проверки стиля: `golint ./...`
- UPPER_SNAKE_CASE для констант: `const MAX_RETRIES = 3`
- camelCase для переменных: `maxRetries := 3`
- PascalCase для экспортированных функций: `func GetDevices()`
- Комментарии на английском (Go соглашение): `// GetDevices returns all devices`

### JavaScript код
- Используйте Prettier/ESLint если требуется
- camelCase для переменных и методов: `deviceId`, `onWakeClick()`
- PascalCase для классов: `class WOLApp`
- Комментарии на русском: `// Получить все устройства`

### HTML/CSS
- Используйте семантический HTML5
- BEM или подобную методологию для CSS классов
- Проверяйте валидность HTML: `html5validator`

## Версионирование

Используйте семантическое версионирование (MAJOR.MINOR.PATCH):
- **MAJOR**: Несовместимые изменения API или функциональности
- **MINOR**: Новые функции (обратно совместимо)
- **PATCH**: Исправления ошибок и малые улучшения

Примеры:
- 2.0.0 - полная миграция на Go + Wails
- 2.1.0 - добавлена группировка устройств
- 2.1.1 - исправлена ошибка отправки пакета

## Создание Pull Request

1. **Форкните репозиторий** на GitHub
2. **Создайте ветку**: `git checkout -b feature/название-функции`
3. **Внесите изменения**:
   - Правьте код согласно соглашениям
   - Добавляйте тесты для новых функций
4. **Тестирование**:
   ```bash
   go vet ./...
   gofmt -w .
   wails dev  # Проверьте все работает
   ```
5. **Сделайте коммиты**: `git commit -m "Описание изменений"`
   - Используйте понятные сообщения коммитов
   - На английском или русском по смыслу кода
6. **Отправьте**: `git push origin feature/название-функции`
7. **Создайте Pull Request** на главную ветку

## Идеи для развития

### Краткосрочные (v2.1.0)
- [ ] Группировка устройств по категориям
- [ ] Экспорт/импорт списка устройств (CSV/JSON)
- [ ] История последних включений
- [ ] Отправка нескольких пакетов подряд с задержкой

### Среднесрочные (v2.2.0)
- [ ] Веб-интерфейс доступный через localhost:3000
- [ ] Проверка доступности устройства (ping)
- [ ] Расписание для автоматического включения (cron)
- [ ] REST API сервер для интеграций
- [ ] Темный/светлый режим переключатель

### Долгосрочные (v3.0.0)
- [ ] Облачная синхронизация (облако)
- [ ] Мобильное приложение (Flutter/React Native)
- [ ] Аутентификация и управление правами доступа
- [ ] Поддержка WoL через прокси/VPN

## Тестирование

### Ручное тестирование

1. **Юнит тесты Go**:
```bash
go test ./internal/...
go test -v ./internal/device/...
```

2. **Проверка с помощью wails dev**:
   - Добавьте несколько устройств
   - Нажмите Wake на каждом
   - Отредактируйте и удалите устройство
   - Проверьте, что данные сохраняются в JSON файл

3. **Проверка API методов** в DevTools консоли:
```javascript
// В консоли браузера (Cmd+Alt+I)
window.go.main.AppInstance.GetDevices().then(devices => console.log(devices));
```

## Сборка под различные платформы

### На macOS
```bash
# Для M4/Apple Silicon
wails build -platform darwin/arm64 -o "WOL Pro"

# Для Intel x86_64
wails build -platform darwin/amd64 -o "WOL Pro-Intel"
```

### На Linux
```bash
# Требует GTK3+ и WebKit2
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev
wails build -platform linux/amd64 -o wol-pro
```

### На Windows
```bash
wails build -platform windows/amd64 -o "WOL-Pro.exe"
```

### Кроссплатформенная сборка с Docker
```bash
# Linux версия на macOS/Windows
docker run --rm -v $(pwd):/work golang:1.21-alpine \
    sh -c "apk add --no-cache libwebkit2gtk-4.0-dev && cd /work && wails build -platform linux/amd64"
```

## Лицензия

Этот проект распространяется под MIT лицензией.

## Поддержка и связь

- 📧 Issues: GitHub Issues для сообщений об ошибках
- 💬 Discussions: GitHub Discussions для обсуждений и идей
- 📖 Документация: README.md и BUILDING.md

## Спасибо за вклад!

Мы ценим каждый вклад в развитие проекта WOL Pro! 🙏

Если у вас есть вопросы - создавайте Issues или Discussions в GitHub репозитории.
