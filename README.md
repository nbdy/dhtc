# dhtc

![Linux](https://github.com/nbdy/dhtc/actions/workflows/linux.yml/badge.svg)
![Windows](https://github.com/nbdy/dhtc/actions/workflows/windows.yml/badge.svg)
![OSX](https://github.com/nbdy/dhtc/actions/workflows/osx.yml/badge.svg)

dht crawler with a web ui<br>
build your own torrent search engine!

## requirements
[golang 1.21](https://go.dev/dl/) or
[docker](https://docs.docker.com/get-docker/)

## features

- [X] Multiplatform compatibility
  - [X] Linux
  - [X] Windows
  - [X] OSX
- [X] Counter of found torrents
- [X] Search by
  - [X] Name
  - [X] Info hash
  - [X] File name
  - [X] Date (only by day)
- [X] Search types (does not work for Date category)
  - [X] contains
  - [X] equals
  - [X] starts with
  - [X] ends with
- [X] Sortable tables
- [X] Interface to add filters for notifications
- [X] Notify on title found
  - [X] Telegram
  - [ ] Mail
- [ ] Expandable list items with extra info
  - [ ] List of files
  - [ ] Movie/Book/Music metadata lookup
- [X] Regex based blacklist
- [ ] Statistics
  - [ ] Line charts for day, week, month, year
  - [ ] Pie / Bubble charts for categories
  - [ ] Pie / Bubble charts for file types


## how to..
### ..run locally
```shell
go run main.go
```
### ..run containerized
```shell
./run.sh
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
