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
	parser          *gofeed.Parser
	currentFeed     *feedMutex
	lastElementTime time.Time
	updatesChannel  chan *gofeed.Item
}

type feedMutex struct {
	gofeed.Feed
	sync.RWMutex
}

// NewRssSubscriber returns create a new subscriber
// it will request a link with a given interval
func NewRssSubscriber(link string, interval time.Duration) (rss *RssSubscriber, err error) {

	rss = &RssSubscriber{
		link:            link,
		interval:        interval,
		parser:          gofeed.NewParser(),
		currentFeed:     &feedMutex{},
		lastElementTime: time.Time{},
		updatesChannel:  make(chan *gofeed.Item),
	}

	feed, err := rss.parser.ParseURL(link)
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

	go rss.subscribe()

	return
}

// GetUpdatesChannel returns channel
// for get updates from rss
func (rss *RssSubscriber) GetUpdatesChannel() (c chan *gofeed.Item) {
	return rss.updatesChannel
}

// GetCurrentFeed returns actual rss feed
func (rss *RssSubscriber) GetCurrentFeed() (feed gofeed.Feed) {
	rss.currentFeed.RLock()
	defer rss.currentFeed.RUnlock()
	return rss.currentFeed.Feed
}

func (rss *RssSubscriber) subscribe() {
	for {
		time.Sleep(rss.interval)

		feed, err := rss.parser.ParseURL(rss.link)
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
				rss.updatesChannel <- item
			}
		}
		rss.lastElementTime = lastElementTmp
	}
}
