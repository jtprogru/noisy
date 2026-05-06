# Noisy

A simple Go utility that generates random HTTP/DNS traffic noise in the background while you go about your regular web browsing, to make your web traffic data less valuable for selling and for extra obscurity.

This is a Go port of the original [Python project](https://github.com/1tayH/noisy) by Itay Hury.

## Quick Start

### Installation via Homebrew (macOS/Linux)

```shell
brew tap jtprogru/tap https://github.com/jtprogru/homebrew-tap
brew install noisy
```

### Installation from Binary Releases

Pre-built binaries (`tar.gz`) for `darwin/linux × amd64/arm64` are published on the [Releases page](https://github.com/jtprogru/noisy/releases) by GoReleaser, with GPG-signed `checksums.txt`.

```shell
# Example — Linux x86_64
VERSION=0.2.0
curl -sL "https://github.com/jtprogru/noisy/releases/download/v${VERSION}/noisy_Linux_x86_64.tar.gz" \
  | tar -xz noisy
sudo mv noisy /usr/local/bin/
```

### Installation via `go install`

```shell
go install github.com/jtprogru/noisy@latest
```

### Build from Source

If you prefer to build from source:

```shell
git clone https://github.com/jtprogru/noisy.git
cd noisy
go build -o noisy .
```

Or using make:

```shell
make build
```

### Configuration

Before running, create a `config.json` file with your settings. See [config.example.json](config.example.json) for a template:

```json
{
  "max_depth": 25,
  "min_sleep": 1,
  "max_sleep": 5,
  "timeout": false,
  "root_urls": [
    "https://www.example.com"
  ],
  "blacklisted_urls": [],
  "user_agents": [
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
  ]
}
```

## Usage

### First Run

```shell
./noisy --config config.json
```

> **⚠️ Note on Startup Delay:** When you first start `noisy`, it may appear to hang for several seconds (typically 3-10 seconds). This is **normal behavior** - the crawler is:
> 1. Loading and parsing the configuration file
> 2. Connecting to the first root URL from your list
> 3. Downloading and parsing the HTML to extract links
>
> **Tip:** Use `--log debug` or `--verbose` on first run to see detailed progress:
> ```shell
> ./noisy --config config.json --verbose
> ```

### Command Line Options

```shell
$ ./noisy --help
Usage of ./noisy:
  -config string
        config file path (required)
  -log string
        logging level: debug, info, warning, error (default "info")
  -timeout int
        stop after N seconds (0 = no timeout)
  -verbose
        enable verbose/debug logging (shorthand for -log debug)
```

### Logging Levels

| Level | Description |
|-------|-------------|
| `debug` | Maximum detail: link extraction, URL visits, errors, dead ends |
| `info` | Basic information: visited URLs, timeout notifications |
| `warning` | Only warnings: connection errors, response read errors |
| `error` | Only critical errors |

### Examples

```shell
# Run with debug logging (see all activity)
./noisy --config config.json --log debug

# Use verbose flag (same as --log debug)
./noisy --config config.json --verbose

# Run with a 5-minute timeout
./noisy --config config.json --timeout 300

# Run with minimal logging (only warnings and errors)
./noisy --config config.json --log warning
```

### Expected Output

With `--log debug`, you'll see continuous activity:

```shell
$ ./noisy --config config.json --verbose
2026/03/15 14:53:09 noisy.go:388: [DEBUG] found 1 links
2026/03/15 14:53:09 noisy.go:319: [INFO] Visiting https://iana.org/domains/example
2026/03/15 14:53:10 noisy.go:307: [DEBUG] Hit a dead end, moving to the next root URL
2026/03/15 14:53:11 noisy.go:388: [DEBUG] found 3 links
2026/03/15 14:53:11 noisy.go:319: [INFO] Visiting https://www.iana.org/
2026/03/15 14:53:12 noisy.go:319: [INFO] Visiting https://www.iana.org/time-zones
...
```

With `--log info`, only visited URLs are shown:

```shell
$ ./noisy --config config.json --log info
2026/03/15 14:53:09 noisy.go:319: [INFO] Visiting https://iana.org/domains/example
2026/03/15 14:53:11 noisy.go:319: [INFO] Visiting https://www.iana.org/
2026/03/15 14:53:12 noisy.go:319: [INFO] Visiting https://www.iana.org/time-zones
...
```

## Docker

### Build the Image

```shell
docker build -t noisy .
```

### Run the Container

```shell
docker run -it noisy --config config.json
```

### Run Multiple Containers

Use `docker-compose` to run multiple instances simultaneously for more noise:

```shell
cd docker-compose
docker-compose build
docker-compose up --scale noisy=3
```

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build Go binary |
| `make run` | Run the crawler |
| `make fmt` | Format Go code (`go fmt`) |
| `make vet` | Run static analysis (`go vet`) |
| `make lint` | Run all linters (fmt + vet) |
| `make clean` | Remove build artifacts |
| `make docker.build` | Build Docker image via docker-compose |
| `make docker.run` | Build and run Docker container |
| `make docker.logs` | Show last 100 lines of container logs |
| `make docker.logf` | Follow container logs in real-time |

## systemd (Auto-start on Boot)

You can configure `noisy` to start automatically on system boot using systemd. The provided service file includes:

- Automatic restart on failure
- Proper logging via journald
- Graceful shutdown on SIGINT/SIGTERM

### Setup

```shell
# Copy the service file
sudo cp systemd/noisy.service /etc/systemd/system

# Reload systemd and enable the service
sudo systemctl daemon-reload
sudo systemctl enable noisy && sudo systemctl start noisy
```

### Viewing Logs

```shell
# View recent logs (last 100 lines)
journalctl -u noisy -n 100

# Follow logs in real-time
journalctl -f -u noisy

# View logs with full timestamps and no pager
journalctl -u noisy --no-pager -o short-precise
```

### Changing Log Level

Edit `/etc/systemd/system/noisy.service` and modify the `--log` flag:

```ini
ExecStart=/opt/noisy/noisy --config /opt/noisy/config.json --log debug
```

Then reload:

```shell
sudo systemctl daemon-reload
sudo systemctl restart noisy
```

## Prerequisites

- **Go** 1.21 or higher (for building from source)
- **Docker** (optional, for containerized deployment)
- **docker-compose** (optional, for multi-container deployment)

## Troubleshooting

### "Noisy appears to hang on startup"

This is normal. The initial connection and link extraction can take 3-10 seconds. Use `--verbose` to see progress.

### "Error connecting to root url"

Some URLs in your `root_urls` list may be temporarily unavailable. The crawler will automatically retry with a different URL. Consider:
- Checking your internet connection
- Verifying URLs are accessible in your browser
- Adding more diverse root URLs to your config

### "Timeout has exceeded"

The crawler stopped after the specified timeout period. This is expected behavior. To run continuously:
- Remove the `--timeout` flag
- Set `"timeout": false` in your config

## Authors

- **Itay Hury** - *Initial work (Python)* - [1tayH](https://github.com/1tayH)
- **Michael Savin** - *Go port* - [jtprogru](https://github.com/jtprogru)

See also the list of [contributors](https://github.com/1tayH/Noisy/contributors) who participated in the original Python project.

## License

This project is licensed under the GNU GPLv3 License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

This project has been inspired by:

- [RandomNoise](http://www.randomnoise.us)
- [web-traffic-generator](https://github.com/ecapuano/web-traffic-generator)
