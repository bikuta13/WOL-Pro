# WOL Pro готовые сборки

Эта папка содержит готовые архивы приложения для скачивания с GitHub.

## Файлы

- `WOL-Pro-macOS-arm64.zip` - macOS Apple Silicon (M1/M2/M3/M4)
- `WOL-Pro-macOS-amd64.zip` - macOS Intel
- `WOL-Pro-Windows-amd64.zip` - Windows 64-bit

## Как запустить

### macOS

1. Скачайте нужный zip-архив.
2. Распакуйте архив.
3. Откройте `WOL Pro.app`.

Приложение подписано локальной self-signed подписью Wails. Если macOS покажет предупреждение безопасности, откройте приложение через контекстное меню: правый клик по `WOL Pro.app` -> `Open`.

### Windows

1. Скачайте `WOL-Pro-Windows-amd64.zip`.
2. Распакуйте архив.
3. Запустите `WOL-Pro.exe`.

Для работы Wails-приложения на Windows нужен Microsoft Edge WebView2 Runtime. На современных Windows он обычно уже установлен.

## Проверка SHA256

Хеши архивов лежат в `SHA256SUMS.txt`.
