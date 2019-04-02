package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	// log "github.com/Sirupsen/logrus"
	"github.com/go-pg/pg"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"
	"unicode"
	"unicode/utf8"

	"homedog/ORM"
	Database "homedog/Platform/Database"
)

// Example is we're trying to find a 2 bedroom apartment for July 1, not basement or condo,
// 5km radius from coordinates except for Hochelaga and Outremont, specific areas of town we want to avoid

// That's easy to do by starting with Craigslist and Kijiji's built-in filters from their website
// (RSS links configured below), supplemented with a few 'removal' keywords that flag a post as
// definitely not interesting.
//
var REMOVALS = []string{"may 1*",
	"june 1*",
	"hochelaga",
	"outremont",
	"condo",
	"basement",
	"sous[- ]sol",
	// ( "2e", "triplex" ),
}

type Query struct {
	Channel Channel `xml:"channel"`
}
type Channel struct {
	Items []Item `xml:"item"`
}
type Item struct {
	Id    int
	Title string `xml:"title"`
	Link  string `xml:"link"`
	Body  string `xml:"description"`
}

var (
	db         *pg.DB
	flag_email *bool
	flag_init  *bool
	SENDER     = os.Getenv("HOMEDOG_SENDER")
)

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile | log.Lmicroseconds)

	flag_email = flag.Bool("email", true, "send emails")
	flag.Parse()

	// log.SetFlags(log.LstdFlags | log.Lshortfile)
	db_init()
}

func main() {
	log.Println("Homedog starting")

	config := getSubscribers()

	for {
		for _, subscriber := range config.Subscribers {
			check("craigslist", subscriber)
			check("kijiji", subscriber)
		}

		duration := 2 * time.Second
		log.Println("sleeping for", duration, "...")
		// time.Sleep(time.Duration(60*6) * time.Second)
		time.Sleep(duration)
	}
}

// --------------------------------------------------------------------------------

func db_init() {
	log.Println("Connecting to DB")

	db = Database.Connect()

	log.Println(db)
}

// --------------------------------------------------------------------------------

func check(source string, subscriber *Subscriber) { //url string, email string) {
	log.Println("Checking", source)

	items := fetch(source, subscriber)

	post_items(source, items, subscriber.Email)
}

// --------------------------------------------------------------------------------

func fetch(source string, subscriber *Subscriber) []Item {
	var (
		err       error
		xml_bytes []byte
		url       = subscriber.UrlForSource(source)
	)
	log.Println("Fetch:", source, url)

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	client.Get(url)

	if !strings.HasPrefix(url, "file://") {
		res, err := client.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		xml_bytes, err = ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		reader, err := os.Open(strings.TrimLeft(url, "file://"))
		if err != nil {
			log.Fatal(err)
		}
		xml_bytes, err = ioutil.ReadAll(reader)
		if err != nil {
			log.Fatal(err)
		}
	}

	enc, _ := charset.Lookup("utf-8")
	filter := transform.Chain(enc.NewDecoder(), transform.RemoveFunc(func(r rune) bool {
		return r == utf8.RuneError
	}))
	t := transform.NewReader(strings.NewReader(string(xml_bytes)), filter)
	xml_bytes, err = ioutil.ReadAll(t)
	if err != nil {
		log.Printf("ReadAll returned %s", err)
	}
	// log.Printf("%v\n",xml_bytes)

	var items []Item
	if source == "kijiji" {
		items, err = unmarshal_kijiji(xml_bytes)
	}
	if source == "craigslist" {
		items, err = unmarshal_craigslist(xml_bytes)
	}
	if err != nil {
		log.Println(err)
		return nil
	}

	preprocess(items)

	return items
}

func unmarshal_craigslist(bytes []byte) ([]Item, error) {
	var q Channel
	err := xml.Unmarshal(bytes, &q)
	return q.Items, err
}

func unmarshal_kijiji(bytes []byte) ([]Item, error) {
	var q Query
	err := xml.Unmarshal(bytes, &q)
	return q.Channel.Items, err
}

// --------------------------------------------------------------------------------

func preprocess(items []Item) {
	for ix := range items {
		item := &items[ix]
		item.Link = html.UnescapeString(item.Link)
		item.Title = html.UnescapeString(item.Title)
		item.Body = html.UnescapeString(item.Body)
	}
}

// items from RSS
func post_items(source string, items []Item, email string) {
	var posts []ORM.Post
	if err := db.Model(&posts).Where("recip = ?", email).Select(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	for _, rssItem := range items {
		match := false
		for _, post := range posts {
			score := rate(rssItem, &post)

			if score >= 1 {
				match = true
				break
			}
		}
		if !match {
			send(source, rssItem, email)
		}
	}
}

func rate(rssItem Item, dbItem *ORM.Post) int {
	score := 0

	rssTitle := normalize(rssItem.Title)
	rssBody := normalize(rssItem.Body)

	dbTitle := normalize(dbItem.Title)
	dbBody := normalize(dbItem.Body)

	if rssTitle == dbTitle {
		score += 1
	}
	if rssBody == dbBody {
		score += 1
	}
	if rssItem.Link == dbItem.Url {
		score += 1
	}

	if score > 1 && rssItem.Link != dbItem.Url {
		// increment(dbItem)
	}

	agg := fmt.Sprintf("%s %s", rssTitle, rssBody)

	words := strings.Split(agg, " ")

	for _, kwd := range REMOVALS {
		if contains(words, kwd) {
			score += 1
		}
	}

	return score
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

func normalize(s string) string {
	// Convert &amp; to &, etc
	s = html.UnescapeString(s)

	// Remove accents (Mn: nonspacing marks), and non-alphabetic characters except spaces
	isNonAlphabetic := func(r rune) bool {
		return unicode.Is(unicode.Mn, r) || (!unicode.IsLetter(r) && !unicode.IsSpace(r))
	}
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isNonAlphabetic), norm.NFC)
	reader := transform.NewReader(strings.NewReader(s), t)
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Sprintf("<err:%s %s>", err, s)
	}
	n := string(bytes[:])

	// Convert any whitespace to ' '
	re1 := regexp.MustCompile("(\\s)+")
	re2 := regexp.MustCompile(" (\\s)+")

	n = re1.ReplaceAllString(n, " ")
	n = re2.ReplaceAllString(n, "")

	return n
}

/*
func increment(dbItem Item) {
	log.WithFields(log.Fields{
		"title": dbItem.Title,
		"id":    dbItem.Id,
	}).Info("Hide")

	stmt, err := dbi.Prepare("UPDATE posts SET counter=counter+1 WHERE id=?")
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(dbItem.Id)
	if err != nil {
		log.Fatal(err)
	}
}
*/

func send(source string, rssItem Item, recip string) {
	log.Printf("Sending to %s: %s\n", recip, rssItem.Title)

	post := ORM.Post{
		Source:  source,
		Recip:   recip,
		Title:   rssItem.Title,
		Body:    rssItem.Body,
		Url:     rssItem.Link,
		Counter: 1,
	}
	if err := db.Insert(&post); err != nil {
		log.Fatal(err)
	}

	subject := fmt.Sprintf("Homedog #%d - %s", post.Id, rssItem.Title)

	email(post.Id, recip, source, subject, rssItem.Title, rssItem.Link, rssItem.Body)
}

func email(id int64, recip string, source string, subject string, title string, link string, body string) {
	if !*flag_email {
		return
	}

	auth := smtp.PlainAuth("", os.Getenv("HOMEDOG_AWS_KEY"),
		os.Getenv("HOMEDOG_AWS_SECRET"),
		os.Getenv("HOMEDOG_SMTP_HOST"))

	to := []string{SENDER, recip}

	type Post struct {
		ID      int64
		Source  string
		Subject string
		Title   string
		Body    string
		Link    string
	}
	post := Post{id,
		source,
		html.UnescapeString(subject),
		html.UnescapeString(title),
		html.UnescapeString(body),
		link}

	tmpl, err := template.New("test").Parse(
		"To: " + recip + "\r\n" +
			"Subject: {{.Subject}}\r\n" +
			"MIME-Version: 1.0\r\nContent-Type: text/html\r\n\r\n<!DOCTYPE html>\r\n<html>\r\n<body>\r\n" +
			"<p><a href=\"{{.Link}}\">{{.Title}}</a><p>{{.Body}}<p>Source: {{.Source}}</body></html>\r\n")

	if err != nil {
		panic(err)
	}

	var doc bytes.Buffer

	err = tmpl.Execute(&doc, post)
	if err != nil {
		panic(err)
	}

	var (
		smtpHost   = os.Getenv("HOMEDOG_SMTP_HOST")
		smtpPort   = os.Getenv("HOMEDOG_SMTP_PORT")
		smtpServer = fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	)

	// log.Println("Connecting to", smtpServer)

	msg := []byte(doc.Bytes())
	err = smtp.SendMail(smtpServer, auth, SENDER, to, msg)
	if err != nil {
		log.Fatal(err)
	}
}
