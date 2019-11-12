package main

func (sub *Subscriber) UrlForSource(source string) *string {
	switch sub.Type {

	case "housing":
		switch source {

		case "craigslist":
			return sub.HousingUrlForCraigslist()

		case "kijiji":
			return sub.HousingRssUrlForKijiji()
		}

	case "parking":
		switch source {

		case "craigslist":
			return sub.ParkingUrlForCraigslist()

		case "kijiji":
			return sub.ParkingUrlForKijiji()
		}
	}
	return nil
}
