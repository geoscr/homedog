package main

import (
	"fmt"
	"net/url"
	"strconv"
)

func (sub *Subscriber) ParkingUrlForCraigslist() *string {
	// apa / prk
	base := "https://montreal.craigslist.org/search/prk?"

	v := url.Values{}
	v.Set("availabilityMode", "0")
	v.Set("bundleDuplicates", "1")
	v.Set("cc", "us")
	v.Set("format", "rss")
	v.Set("max_price", strconv.Itoa(sub.Properties.Max_price))
	v.Set("min_price", strconv.Itoa(sub.Properties.Min_price))
	v.Set("postal", sub.Properties.Postal)
	v.Set("search_distance", strconv.Itoa(sub.Properties.Search_distance))

	r := fmt.Sprintf("%s%s", base, v.Encode())
	return &r
}

func (sub *Subscriber) ParkingUrlForKijiji() *string {
	rssUrl := "https://www.kijiji.ca/rss-srp-storage-parking/ville-de-montreal/"
	locationCode := "c39l1700281"
	rssUrl = fmt.Sprintf("%s%s", rssUrl, locationCode)

	v := url.Values{}
	v.Set("ad", "offering")
	v.Set("radius", "1.0")
	v.Set("address", sub.Properties.Postal)
	v.Set("more-info", "parking")
	// Don't set ll, to preserve comma

	r := fmt.Sprintf("%s?%s&ll=%s", rssUrl, v.Encode(), sub.Properties.Coordinates)
	return &r
}
