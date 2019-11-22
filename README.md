
## Setup
```
go get github.com/boltegg/go-rss-subscriber
```

## Usage
```
rss, err := NewRssSubscriber("https://www.reddit.com/.rss", time.Second * 7)
if err != nil {
    panic(err)
}

for update := range rss.GetUpdatesChannel() {
    fmt.Println(update.Title)
    fmt.Println(update.Description)
    fmt.Println(update.Link)
}
```