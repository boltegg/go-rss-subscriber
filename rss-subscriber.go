package grs

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"sync"
	"time"
)

// RssSubscribe is a main struct of this package
type RssSubscriber struct {
	link            string
	interval        time.Duration
	Parser          *gofeed.Parser
	currentFeed     *feedMutex
	lastElementTime time.Time
}

type feedMutex struct {
	gofeed.Feed
	sync.RWMutex
}

// NewRssSubscriber create a new subscriber
func NewRssSubscriber(link string, interval time.Duration) (rss *RssSubscriber) {

	rss = &RssSubscriber{
		link:            link,
		interval:        interval,
		Parser:          gofeed.NewParser(),
		currentFeed:     &feedMutex{},
		lastElementTime: time.Time{},
	}
	return
}

// Subscribe to the feed
func (rss *RssSubscriber) Subscribe() (ch chan *gofeed.Item, err error) {

	ch = make(chan *gofeed.Item)

	feed, err := rss.Parser.ParseURL(rss.link)
	if err != nil {
		return
	}

	rss.currentFeed.Feed = *feed

	for _, item := range feed.Items {
		if item.PublishedParsed == nil {
			err = fmt.Errorf("can't work with this feed, PublishedParsed is nil")
			return
		}
		if item.PublishedParsed.Sub(rss.lastElementTime) > 0 {
			rss.lastElementTime = *item.PublishedParsed
		}
	}

	go rss.subscribe(ch)
	return
}

// GetCurrentFeed returns actual rss feed
func (rss *RssSubscriber) GetCurrentFeed() (feed gofeed.Feed) {
	rss.currentFeed.RLock()
	defer rss.currentFeed.RUnlock()
	return rss.currentFeed.Feed
}

func (rss *RssSubscriber) subscribe(ch chan *gofeed.Item) {
	for {
		time.Sleep(rss.interval)

		feed, err := rss.Parser.ParseURL(rss.link)
		if err != nil {
			continue
		}

		rss.currentFeed.Lock()
		rss.currentFeed.Feed = *feed
		rss.currentFeed.Unlock()

		var lastElementTmp = time.Time{}
		for _, item := range feed.Items {
			if item.PublishedParsed == nil {
				continue
			}
			if item.PublishedParsed.Sub(lastElementTmp) > 0 {
				lastElementTmp = *item.PublishedParsed
			}

			if item.PublishedParsed.Sub(rss.lastElementTime) > 0 {
				ch <- item
			}
		}
		rss.lastElementTime = lastElementTmp
	}
}
