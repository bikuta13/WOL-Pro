#!/bin/bash
# WOL Pro - Build Script for Go + Wails
# Собирает приложение для разных платформ

set -e

VERSION="2.0.1"
APP_NAME="WOL-Pro"

echo "🏗️  WOL Pro v${VERSION} Builder"
echo "=================================="

# Проверка зависимостей
if ! command -v wails &> /dev/null; then
    echo "❌ Wails не установлен"
    echo "Установите: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo "❌ Go не установлен"
    exit 1
fi

# Очистка старых сборок
echo "🧹 Очистка старых сборок..."
rm -rf build/bin
mkdir -p build/bin

# Получение информации о системе
OS=$(uname -s)
ARCH=$(uname -m)

case "$OS" in
    Darwin)
        case "$ARCH" in
            arm64)
                PLATFORM="darwin/arm64"
                OUTPUT="$APP_NAME"
                ;;
            x86_64)
                PLATFORM="darwin/amd64"
                OUTPUT="$APP_NAME"
                ;;
            *)
                echo "❌ Неподдерживаемая архитектура macOS: $ARCH"
                exit 1
                ;;
        esac
        ;;
    Linux)
        PLATFORM="linux/amd64"
        OUTPUT="wol-pro"
        ;;
    *)
        echo "❌ Неподдерживаемая ОС: $OS"
        exit 1
        ;;
esac

echo "📦 Сборка для: $PLATFORM"
echo "🎯 Выходной файл: $OUTPUT"

# Загрузка зависимостей
echo "📥 Загрузка зависимостей..."
go mod tidy

# Сборка с Wails
echo "⚙️  Сборка приложения..."
wails build -platform "$PLATFORM" -o "$OUTPUT"

echo ""
echo "✅ Сборка завершена!"
echo "📍 Приложение находится в: ./build/bin/$OUTPUT"

if [ "$OS" = "Darwin" ]; then
    echo "📱 На macOS приложение упаковано в .app бандл"
    echo "🚀 Для запуска: open ./build/bin/$OUTPUT.app"
fi
