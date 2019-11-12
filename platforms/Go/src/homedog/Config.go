package main

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
	Exclusions      []string
}

type Subscriber struct {
	Type       string
	Email      string
	Properties SubscriberProperties
}

type Config struct {
	Subscribers []*Subscriber
}
