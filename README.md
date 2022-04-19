# dhtc

dht crawler with a web ui<br>

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
- [X] Interface to add filters for notifications
- [X] Notify on title found
  - [X] Telegram
  - [ ] Mail
- [ ] Expandable list items with extra info
  - [ ] List of files
  - [ ] Movie/Book/Music metadata lookup
- [ ] Safe mode
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
