# Инструкция по сборке WOL Pro (Go + Wails)

Это руководство содержит подробные инструкции по сборке и разработке приложения WOL Pro на Go + Wails.

## Предварительные требования

### Для разработки
- **Go 1.22 или выше**
- **Wails 2.12 или выше**
- **Git** - для клонирования репозитория

### Для сборки под macOS
- **macOS 11+**
- **Xcode Command Line Tools** (`xcode-select --install`)

### Для сборки под Linux
- **GTK 3.0+** или **libwebkit2gtk-4.0** (для Debian/Ubuntu)
- На Ubuntu/Debian: `sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev`

### Для сборки под Windows
- **Windows 10+**
- **MinGW-w64** или **Visual Studio Build Tools** (опционально)

## Установка зависимостей

### 1. Установка Go

#### macOS
```bash
brew install go
```

#### Windows
Скачайте с https://golang.org/dl и установите MSI файл

#### Linux (Ubuntu/Debian)
```bash
sudo apt-get update
sudo apt-get install golang-go
```

Убедитесь, что Go установлен:
```bash
go version
```

### 2. Установка Wails

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
```

Убедитесь, что Wails в PATH:
```bash
wails version
```

### 3. Клонирование проекта

```bash
git clone https://github.com/bikuta13/WOLPRo.git
cd WOLPRo
```

### 4. Загрузка зависимостей Go

```bash
go mod tidy
```

## Разработка

### Запуск в режиме разработки (с горячей перезагрузкой)

```bash
wails dev
```

Это откроет приложение с поддержкой горячей перезагрузки при изменении кода. Фронтенд будет доступен на `http://localhost:34115`.

### Структура проекта для разработки

```
wol-pro/
├── main.go                      # Точка входа приложения
├── internal/
│   ├── app/
│   │   └── app.go              # Основная логика (API)
│   ├── device/
│   │   └── manager.go          # Менеджер устройств
│   └── wol/
│       └── magic_packet.go     # WOL функциональность
├── frontend/
│   └── dist/
│       ├── index.html          # HTML интерфейс
│       └── app.js              # JavaScript логика
├── go.mod                       # Go зависимости
└── go.sum
```

### Изменение интерфейса

Редактируйте файлы в `frontend/dist/`:
- `index.html` - HTML структура
- `app.js` - JavaScript логика и взаимодействие с бэкендом

При запуске `wails dev` изменения автоматически перезагружаются.

### Изменение логики

Редактируйте файлы в `internal/`:
- `app/app.go` - методы API для фронтенда
- `device/manager.go` - управление устройствами
- `wol/magic_packet.go` - WOL функциональность

При сохранении файлов приложение автоматически перекомпилируется.

## Сборка приложения

### macOS (Apple Silicon / M4)

```bash
wails build -platform darwin/arm64 -o "WOL Pro"
```

Результат: `./build/bin/WOL Pro.app`

**Запуск**:
```bash
open "./build/bin/WOL Pro.app"
```

### macOS (Intel x86_64)

```bash
wails build -platform darwin/amd64 -o "WOL Pro"
```

**Создание DMG образа**:
```bash
hdiutil create -volname "WOL Pro" -srcfolder "./build/bin" -ov -format UDZO wolpro.dmg
```

### Windows (AMD64)

```bash
wails build -platform windows/amd64 -o "WOL-Pro.exe"
```

Результат: `.\build\bin\WOL-Pro.exe`

### Linux (x86_64)

```bash
wails build -platform linux/amd64 -o wol-pro
```

Результат: `./build/bin/wol-pro`

### Использование скрипта сборки

Для быстрой сборки на текущей платформе:

```bash
chmod +x build.sh
./build.sh
```

## Отладка

### Логирование в разработке

При запуске `wails dev` логи приложения выводятся в консоль.

Go логи:
```go
log.Println("Debug message")
```

JavaScript логи:
```javascript
console.log("Debug message");
```

### DevTools в режиме разработки

При `wails dev` можно открыть DevTools:
- **macOS**: Cmd+Alt+I
- **Windows/Linux**: Ctrl+Alt+I

### Отладка Go кода

Используйте Delve debugger:

```bash
dlv debug ./cmd/wails
```

## Сборка для распространения

### Финальная сборка со всеми платформами

```bash
# macOS M4
wails build -platform darwin/arm64 -o "WOL Pro"

# macOS Intel
wails build -platform darwin/amd64 -o "WOL Pro-Intel"

# Windows
wails build -platform windows/amd64 -o "WOL-Pro.exe"

# Linux
wails build -platform linux/amd64 -o wol-pro
```

### Создание архивов

**macOS**:
```bash
ditto -c -k --sequesterRsrc "./build/bin/WOL Pro.app" "WOL-Pro-macOS-M4.zip"
```

**Windows**:
```bash
7z a WOL-Pro-Windows.zip .\build\bin\WOL-Pro.exe
```

**Linux**:
```bash
tar -czf wol-pro-linux.tar.gz ./build/bin/wol-pro
```

## Проверка сборки

### Проверка зависимостей

```bash
go mod verify
```

### Статический анализ кода

```bash
go vet ./...
```

### Форматирование кода

```bash
gofmt -w .
```

## Решение проблем

### Ошибка: "wails: command not found"

Убедитесь, что `$GOPATH/bin` в PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Ошибка: "Go version too old"

Обновите Go:

```bash
go get golang.org/dl/go1.21
go1.21 download
```

### Ошибка при сборке на Linux: "package development files for GTK not found"

```bash
# Ubuntu/Debian
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev

# Fedora
sudo dnf install gtk3-devel webkit2gtk3-devel
```

### Ошибка при сборке на macOS: "SDK not found"

```bash
xcode-select --install
```

### Приложение не запускается после сборки

Проверьте логи:
```bash
# macOS
log stream --predicate 'eventMessage contains "WOL"'

# Linux
journalctl -f
```

## Производительность

### Оптимизация сборки

Убедитесь что сборка использует оптимизацию:

```bash
wails build -platform darwin/arm64 -o "WOL Pro" -w
```

Флаг `-w` убирает отладочную информацию для меньшего размера.

### Размер приложения

Типичные размеры:

- macOS .app: ~60 MB
- Windows .exe: ~55 MB
- Linux binary: ~50 MB

## Дополнительно

### Кроссплатформенная сборка с Docker

Для сборки Linux версии на macOS/Windows:

```dockerfile
FROM golang:1.22-alpine

RUN apk add --no-cache \
    git \
    build-base \
    libwebkit2gtk-4.0-dev

WORKDIR /build
COPY . .

RUN go mod tidy
RUN wails build -platform linux/amd64 -o wol-pro
```

## Поддержка

Если у вас есть проблемы при сборке:

1. Проверьте версии: `go version` и `wails version`
2. Убедитесь, что все зависимости установлены
3. Создайте issue с описанием проблемы и выводом ошибки

## Ссылки

- [Документация Wails](https://wails.io/docs)
- [Go Official](https://golang.org)
- [GitHub Issues](https://github.com/bikuta13/WOLPRo/issues)
