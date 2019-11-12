package main

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// type HousingCategory struct {
// }

func (sub *Subscriber) HousingUrlForCraigslist() *string {
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

	r := fmt.Sprintf("%s%s", base, v.Encode())
	return &r
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

func (sub *Subscriber) HousingRssUrlForKijiji() *string {
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

	r := fmt.Sprintf("%s?%s&ll=%s", rssUrl, v.Encode(), sub.Properties.Coordinates)
	return &r
	// return fmt.Sprintf("%s%s", rssUrl, v.Encode())
}

// Unused (for generating a link to web version of RSS feed)
func (sub *Subscriber) HousingWebUrlForKijiji() string {
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

func (sub *Subscriber) WouldRemove(title string, body string) bool {
	agg := fmt.Sprintf("%s %s", title, body)

	words := strings.Split(agg, " ")

	for _, kwd := range sub.Properties.Exclusions {
		if contains(words, kwd) {
			return true
		}
	}

	return false
}
