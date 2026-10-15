# 📚 Содержание WOL Pro

Путеводитель по документации и файлам проекта WOL Pro (Go + Wails версия).

## 🚀 Для новичков

Если это первый раз вы видите WOL Pro, начните отсюда:

1. **[QUICKSTART.md](QUICKSTART.md)** - Быстрый старт за 5-10 минут
   - Установка готового приложения ИЛИ сборка из исходного кода
   - Первоначальная настройка WOL
   - Базовое использование приложения

2. **[README.md](README.md)** - Полное описание проекта
   - Основные возможности приложения
   - Требования к системе
   - Общее руководство по использованию
   - Как найти MAC адреса устройств

## 🔨 Для разработчиков

Если вы хотите собрать приложение из исходного кода или внести вклад:

1. **[BUILDING.md](BUILDING.md)** - Подробная инструкция по сборке и разработке
   - Установка Go и Wails
   - Запуск в режиме разработке (`wails dev` с горячей перезагрузкой)
   - Сборка для Windows, macOS и Linux
   - Оптимизация и решение проблем при сборке

2. **[DEVELOPERS.md](DEVELOPERS.md)** - Руководство разработчика
   - Архитектура проекта (Go + Wails + JavaScript)
   - Описание компонентов (WOL, Device Manager, API)
   - Добавление новых функций
   - Соглашения кодирования (Go и JavaScript)
   - Идеи для развития проекта

3. **[CHANGELOG.md](CHANGELOG.md)** - История изменений
   - Что было добавлено в версии 2.0.0 (миграция на Go)
   - Что было в версии 1.0.0 (архив, Python версия)
   - Планы развития проекта

## 📁 Структура файлов проекта

### Код приложения (Go)
```
main.go                         - Точка входа приложения
internal/
├── app/app.go                 - Основная логика (API методы)
├── device/manager.go          - Управление устройствами (JSON)
└── wol/magic_packet.go        - WOL функциональность
```

### Интерфейс (HTML/CSS/JavaScript)
```
frontend/dist/
├── index.html                 - HTML структура интерфейса
└── app.js                     - JavaScript логика
```

### Конфигурация и сборка
```
go.mod                          - Go модули и зависимости
go.sum                          - Checksum зависимостей
build.sh                        - Скрипт сборки для текущей платформы
```

### Документация
```
README.md                       - Основная документация
QUICKSTART.md                   - Быстрый старт
BUILDING.md                     - Инструкция по сборке и разработке
DEVELOPERS.md                   - Руководство разработчика
CHANGELOG.md                    - История изменений
INDEX.md                        - Этот файл
```

## 🎯 Быстрые ссылки по задачам

### "Хочу просто использовать приложение"
→ Скачайте готовое приложение с GitHub Releases
→ ИЛИ прочитайте [QUICKSTART.md](QUICKSTART.md)

### "Хочу собрать приложение из исходного кода"
→ [BUILDING.md](BUILDING.md) - полная инструкция по сборке

### "Хочу разрабатывать и расширять функциональность"
→ [BUILDING.md](BUILDING.md) - для `wails dev`
→ [DEVELOPERS.md](DEVELOPERS.md) - архитектура и добавление функций

### "У меня проблема при использовании"
→ [QUICKSTART.md](QUICKSTART.md#решение-проблем) - решение проблем
→ [README.md](README.md#предварительные-условия-для-wol) - требования WOL

### "Хочу узнать, что изменилось в версии"
→ [CHANGELOG.md](CHANGELOG.md) - история всех версий

## 📋 Требования для разных сценариев

### Для использования готового приложения
- Windows 10+, macOS 11+ или Linux с GTK3+
- Интернет для скачивания (первый раз)
- Знание MAC адреса целевого устройства
- WOL включен в BIOS целевого устройства

### Для разработки из исходного кода
- Go 1.22+
- Wails 2.12+
- Git (для клонирования репозитория)
- Для macOS: Xcode Command Line Tools (`xcode-select --install`)
- Для Linux: GTK 3+ и WebKit2GTK4 (`sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev`)

## 🔗 Важные ссылки

### Про Wake-on-LAN (WOL)
- [Wake-on-LAN на Wikipedia](https://en.wikipedia.org/wiki/Wake-on-LAN)
- [Как включить WOL в BIOS](https://support.microsoft.com/en-us/help/12356)
- [Как найти MAC адрес](https://www.wikihow.com/Find-the-MAC-Address-of-Your-Computer)

### Про технологии проекта
- [Go Official Documentation](https://golang.org/doc)
- [Wails Documentation](https://wails.io/docs)
- [JavaScript на MDN](https://developer.mozilla.org/en-US/docs/Web/JavaScript)

## 📊 Версии и поддержка

| Версия | Статус | Технологии | Примечание |
|--------|--------|-----------|-----------|
| **2.0.1** | ✅ Текущая | Go + Wails | Валидация IP, стабильная сортировка, обновленные сборки |
| 2.0.0 | 📦 Архив | Go + Wails | Полная миграция на Go |
| 1.0.0 | 📦 Архив | Python + tkinter | Устаревшая, использовать не рекомендуется |

## 📞 Поддержка и связь

- 🐛 **Найден баг?** Создайте [Issue](https://github.com/bikuta13/WOLPRo/issues) на GitHub
- 💡 **Есть идея?** Обсудите в [Discussions](https://github.com/bikuta13/WOLPRo/discussions)
- ❓ **Вопрос?** Смотрите [QUICKSTART.md FAQ](QUICKSTART.md#часто-задаваемые-вопросы)

## 📄 Лицензия

MIT License - смотрите LICENSE файл для деталей

---

**Версия**: 2.0.1 (Go)
**Последнее обновление**: 2026-06-12
**Статус**: ✅ Готово к использованию
