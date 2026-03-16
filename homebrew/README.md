# Homebrew Tap for Noisy

This repository contains the Homebrew formula for installing Noisy on macOS and Linux.

## Usage

```bash
# Add the tap
brew tap jtprogru/noisy https://github.com/jtprogru/noisy

# Install noisy
brew install noisy
```

## For Maintainers

### Creating a New Release

1. **Create a new tag in the main noisy repository:**
   ```bash
   cd /path/to/noisy
   git tag -a v0.2.0 -m "Release v0.2.0"
   git push origin v0.2.0
   ```

2. **Calculate the SHA256 for the new tarball:**
   ```bash
   curl -sL https://github.com/jtprogru/noisy/archive/refs/tags/v0.2.0.tar.gz | shasum -a 256
   ```

3. **Update the formula:**
   - Update the `version` field
   - Update the `sha256` field with the calculated hash

4. **Commit and push the changes:**
   ```bash
   git add homebrew/noisy.rb
   git commit -m "chore: update formula for v0.2.0"
   git push origin main
   ```

### Formula Structure

The formula builds Noisy from source using Go. Requirements:
- Go 1.21+ (installed as build dependency)
- Git (for fetching the source)

### Testing

After installation, verify:
```bash
noisy --help
brew test noisy
```

## License

GPL-3.0 (same as the main Noisy project)
