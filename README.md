# dhtc

[![](http://github-actions.40ants.com/nbdy/dhtc/matrix.svg?only=build)](https://github.com/nbdy/dhtc)

dht crawler with a web ui<br>
build your own torrent search engine!

#### this project is in maintenance mode

#### i'll not add new features due to time being scarce


## requirements
[golang 1.24](https://go.dev/dl/) or
[docker](https://docs.docker.com/get-docker/)

## features

- [X] Multiplatform compatibility
- [X] Counter of found torrents
- [X] Search
- [X] Sortable tables
- [X] Interface to add filters for notifications
- [X] Regex based blacklist
- [X] Notify on title found
  - [X] Telegram
  - [X] Discord
  - [X] Slack
  - [X] Gotify
- [X] Expandable list items with extra info
  - [X] List of files
- [X] Statistics
  - [X] Dashboard with charts
- [X] Download integration
  - [X] Transmission
  - [X] Aria2
  - [X] Deluge
  - [X] qBittorrent

## how to..

### ..run locally

[latest release](https://github.com/nbdy/dhtc/releases/latest)

or

```shell
go run cmd/dhtc/main.go
```

### ..run containerized

```shell
docker compose up
```

either way an instance should be running on [localhost:4200](http://127.0.0.1:4200).

## screenshots

### dashboard

![dashboard](https://i.ibb.co/0rJfG1g/image.png)

### search

![search](https://i.ibb.co/PwWbyK6/image.png)

### watches

![watches](https://i.ibb.co/MfRxvPH/image.png)

### blacklist

![blacklist](https://i.ibb.co/CbwXP5Z/image.png)

## development

### environment setup

to contribute to this project, you'll need:

1. **go**: ensure you have [go 1.24+](https://go.dev/dl/) installed.
2. **pre-commit**: install the [pre-commit](https://pre-commit.com/) framework (see below for installation options).
3. **go modules**: download the project dependencies:
   ```shell
   go mod download
   ```
4. **ui (optional)**: if you plan to modify the styling, you'll need [nodejs](https://nodejs.org/):
   ```shell
   npm install
   ```

### pre-commit hooks

this project uses [pre-commit](https://pre-commit.com/) to ensure code quality through linting, formatting, and automated testing.

to set up pre-commit:

1. install pre-commit:

   it is recommended to use your system package manager or `pipx` to avoid issues with externally managed python environments (PEP 668):

   - **pipx**: `pipx install pre-commit`
   - **arch linux**: `sudo pacman -S pre-commit`
   - **debian/ubuntu**: `sudo apt install pre-commit`
   - **macos**: `brew install pre-commit`

2. install the git hook scripts:
   ```shell
   pre-commit install
   ```

now, `pre-commit` will run automatically on every `git commit`.

you can also run it manually on all files:
```shell
pre-commit run --all-files
```

the hooks include:
- `trailing-whitespace`: trims trailing whitespace.
- `end-of-file-fixer`: ensures files end with a newline.
- `check-yaml`: validates yaml files.
- `golangci-lint`: runs a suite of go linters.
- `go-fmt`: formats go code.
- `go-mod-tidy`: ensures `go.mod` and `go.sum` are up to date.
- `go-test`: runs all project tests.
