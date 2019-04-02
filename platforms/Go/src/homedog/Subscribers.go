package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
)

type SubscriberProperties struct {
	HasPic          int
	Max_bedrooms    int
	Max_price       int
	Min_bedrooms    int
	Min_price       int
	Postal          string
	Coordinates     string
	Search_distance int
	Furnished       int
}

type Subscriber struct {
	Email      string
	Properties SubscriberProperties
}

type Config struct {
	Subscribers []*Subscriber
}

func getSubscribers() *Config {
	var (
		err         error
		config_file []byte
	)

	if config_file, err = ioutil.ReadFile("config/config.json"); err != nil {
		log.Printf("File error: %v\n", err)
		os.Exit(1)
	}

	config_str := string(config_file)

	var config Config
	if err = json.Unmarshal([]byte(config_str), &config); err != nil {
		log.Panicf("File error: %v\n", err)
		os.Exit(1)
	}

	return &config
}

func (sub *Subscriber) UrlForSource(source string) string {
	if source == "craigslist" {
		return sub.UrlForCraigslist()
	}
	if source == "kijiji" {
		return sub.RssUrlForKijiji()
	}
	return ""
}

func (sub *Subscriber) UrlForCraigslist() string {
	base := "https://montreal.craigslist.org/search/apa?"

	v := url.Values{}
	v.Set("availabilityMode", "0")
	v.Set("bundleDuplicates", "1")
	v.Set("format", "rss")
	v.Set("hasPic", strconv.Itoa(sub.Properties.HasPic))
	v.Set("max_bedrooms", strconv.Itoa(sub.Properties.Max_bedrooms))
	v.Set("max_price", strconv.Itoa(sub.Properties.Max_price))
	v.Set("min_bedrooms", strconv.Itoa(sub.Properties.Min_bedrooms))
	v.Set("min_price", strconv.Itoa(sub.Properties.Min_price))
	v.Set("postal", sub.Properties.Postal)
	v.Set("search_distance", strconv.Itoa(sub.Properties.Search_distance))
	v.Set("is_furnished", strconv.Itoa(sub.Properties.Furnished))

	return fmt.Sprintf("%s%s", base, v.Encode())
}

var (
	KIJIJI_WEB_MONTREAL_ROOM_COUNTS = []string{
		"2+1+2__3+1+2",
		"4+1+2",
		"5+1+2",
		"6+1+2+plus",
	}

	KIJIJI_RSS_MONTREAL_ROOM_COUNTS = []string{
		"1+bedroom__1+bedroom+den",
		"2+bedrooms",
		"3+bedrooms",
		"4+plus+bedrooms",
	}
)

func (sub *Subscriber) RssUrlForKijiji() string {
	rssUrl := "https://www.kijiji.ca/rss-srp-apartments-condos/ville-de-montreal/"

	minBedrooms := sub.Properties.Min_bedrooms
	maxBedrooms := sub.Properties.Max_bedrooms
	bedroomFragment := ""

	for r := minBedrooms; r <= maxBedrooms; r++ {
		joiner := ""
		if r < maxBedrooms {
			joiner = "__"
		}
		bedroomFragment = fmt.Sprintf("%s%s%s", bedroomFragment, KIJIJI_RSS_MONTREAL_ROOM_COUNTS[r-1], joiner)
	}
	rssUrl = fmt.Sprintf("%s%s", rssUrl, bedroomFragment)

	locationCode := fmt.Sprintf("c37l1700281a27949001r%d.0", sub.Properties.Search_distance)
	priceCode := fmt.Sprintf("%d__%d", sub.Properties.Min_price, sub.Properties.Max_price)

	rssUrl = fmt.Sprintf("%s/%s", rssUrl, locationCode)

	v := url.Values{}
	v.Set("ad", "offering")
	v.Set("price", priceCode)
	v.Set("address", sub.Properties.Postal)
	v.Set("furnished", strconv.Itoa(sub.Properties.Furnished))
	// v.Set("ll", sub.Properties.Coordinates)

	return fmt.Sprintf("%s?%s&ll=%s", rssUrl, v.Encode(), sub.Properties.Coordinates)
	// return fmt.Sprintf("%s%s", rssUrl, v.Encode())
}

func (sub *Subscriber) WebUrlForKijiji() string {
	webUrl := "https://www.kijiji.ca/b-a-louer/ville-de-montreal/apartment/"

	minBedrooms := sub.Properties.Min_bedrooms
	maxBedrooms := sub.Properties.Max_bedrooms
	bedroomFragment := ""

	for r := minBedrooms; r <= maxBedrooms; r++ {
		joiner := ""
		if r < maxBedrooms {
			joiner = "__"
		}
		bedroomFragment = fmt.Sprintf("%s%s%s", bedroomFragment, KIJIJI_WEB_MONTREAL_ROOM_COUNTS[r-1], joiner)
	}
	webUrl = fmt.Sprintf("%s%s", webUrl, bedroomFragment)

	locationCode := fmt.Sprintf("c37l1700281a27949001r%d.0", sub.Properties.Search_distance)
	priceCode := fmt.Sprintf("%d__%d", sub.Properties.Min_price, sub.Properties.Max_price)

	webUrl = fmt.Sprintf("%s/%s", webUrl, locationCode)

	v := url.Values{}
	v.Set("ad", "offering")
	v.Set("price", priceCode)
	v.Set("address", url.QueryEscape(sub.Properties.Postal))
	v.Set("meuble", strconv.Itoa(sub.Properties.Furnished))

	return fmt.Sprintf("%s?%s&ll=%s", webUrl, v.Encode(), sub.Properties.Coordinates)
}
