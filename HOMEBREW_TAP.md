# Инструкция по созданию Homebrew Tap

## Проблема
Homebrew требует, чтобы формулы находились в отдельном репозитории (tap), а не в основном репозитории проекта.

## Решение: Создать отдельный репозиторий homebrew-tap

### Шаг 1: Создать репозиторий на GitHub

Создайте новый публичный репозиторий:
```
https://github.com/jtprogru/homebrew-tap
```

Или с префиксом formulae:
```
https://github.com/jtprogru/homebrew-formulae
```

### Шаг 2: Клонировать новый репозиторий

```bash
git clone https://github.com/jtprogru/homebrew-tap.git
cd homebrew-tap
```

### Шаг 3: Создать структуру формулы

```bash
mkdir -p Formula
cp ../noisy/homebrew/noisy.rb Formula/noisy.rb
```

### Шаг 4: Создать README

```bash
cat > README.md << EOF
# Homebrew Tap for Noisy

Random HTTP/DNS traffic noise generator

## Usage

\`\`\`bash
brew tap jtprogru/noisy https://github.com/jtprogru/homebrew-tap
brew install noisy
\`\`\`

## License

GPL-3.0
EOF
```

### Шаг 5: Закоммитить и отправить

```bash
git add .
git commit -m "Add noisy formula"
git push origin main
```

### Шаг 6: Проверить установку

```bash
# Удалить старую версию (если была установлена локально)
brew uninstall noisy 2>/dev/null || true

# Добавить tap
brew tap jtprogru/noisy https://github.com/jtprogru/homebrew-tap

# Установить noisy
brew install noisy

# Проверить
noisy --help
```

## Альтернатива: Использовать go install

Пока tap не создан, пользователи могут установить noisy через:

```bash
go install github.com/jtprogru/noisy@latest
```

Для этого нужно:
1. Убедиться, что в noisy.go есть правильный `package main`
2. Пользователи должны иметь Go 1.21+

## Альтернатива: GitHub Releases с бинарниками

Создайте GitHub Action для автоматической сборки бинарников:

1. Создать `.github/workflows/release.yml`
2. Настроить сборку для macOS (arm64, amd64) и Linux (amd64)
3. При создании тега автоматически загружать бинарники в Releases

Тогда пользователи смогут:
```bash
# macOS
brew install --cask noisy  # если будет cask

# Или скачать бинарник напрямую
curl -sL https://github.com/jtprogru/noisy/releases/latest/download/noisy-darwin-arm64 -o noisy
```
