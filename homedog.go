package main

import (
    "bytes"
    "database/sql"
    "encoding/xml"
    "flag"
    "fmt"
    "golang.org/x/net/html"
    "golang.org/x/net/html/charset"
    "golang.org/x/text/transform"
    "golang.org/x/text/unicode/norm"
    "io/ioutil"
    log "github.com/Sirupsen/logrus"
    "net/http"
    "net/smtp"
    "os"
    "regexp"
    "strings"
    "text/template"
    "time"
    "unicode"
    "unicode/utf8"
    _ "github.com/go-sql-driver/mysql"
)

var REMOVALS = []string{ "may 1*",
                         "june 1*",
                         "boringtown",
                         "ghetto",
                         "condo",
                         "basement",
                         "sous[- ]sol",
                         // ( "2e", "triplex" ),
                     }

type Query struct {
    Channel    Channel    `xml:"channel"`
}
type Channel struct {
    Items    []Item    `xml:"item"`
}
type Item struct {
    Id       int
    Title    string    `xml:"title"`
    Link     string    `xml:"link"`
    Body     string    `xml:"description"`
}

var (
    dbi *sql.DB
    flag_email  *bool
    flag_init   *bool
)

func init() {
    flag_email    = flag.Bool("email", true, "send emails")
    flag_init    = flag.Bool("init",  false, "init db")
    flag.Parse()

    // log.SetFlags(log.LstdFlags | log.Lshortfile)
        db_connect()
    // }

}

func main() {
    if *flag_init {
        db_init()
        return
    }

    // set up filters using Craigslist and Kijiji websites, then click RSS and copy the URL in here
    cl := "https://...craigslist.ca/search/apa?lang=en&cc=us&availabilityMode=0&format=rss&max_price=...&min_bedrooms=...&postal=......&search_distance=5"
    kj := "http://www.kijiji.ca/rss-srp-2-bedroom-apartments-condos/.../...r5.0?price=....&address=......&ll=45...,-73...&furnished=0"

    for {
        check("craigslist", cl, my_email)
        check("kijiji", kj, my_email)
        log.Println("sleeping...")
        time.Sleep(time.Duration(60*6)*time.Second)
    }
}

// --------------------------------------------------------------------------------

func db_connect() {
    var err error
    dbi, err = sql.Open("mysql", "root:password@/homedog")
    if err != nil {
        log.Println("DB down : ", err)
    }
}

func db_init() {
    log.Println("Resetting DB and exiting")

    _, err := dbi.Exec(`DROP TABLE IF EXISTS posts`)
    if err != nil {
        log.Fatal(err)
    }

    _, err = dbi.Exec(`CREATE TABLE posts (
                           id        int(11) unsigned  NOT NULL AUTO_INCREMENT,
                           counter   int(11) unsigned  NOT NULL,
                           source    text              NOT NULL,
                           title     text              NOT NULL,
                           body      text              NOT NULL,
                           url       text              NOT NULL,
                           # state     enum             ('NEW','FLAG','HIDE') DEFAULT 'NEW',
                           timestamp datetime         DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                           PRIMARY KEY (id)
                         ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;`)
    if err != nil {
        log.Fatal(err)
    }
}

// --------------------------------------------------------------------------------

func check(source string, url string, email string) {
    log.WithFields(log.Fields{
        source: source,
    }).Info("Check")

    items := fetch(source, url)
    
    post_items(source, items, email)
}

// --------------------------------------------------------------------------------


func fetch(source string, url string) []Item {
    var (
        err error
        xml_bytes []byte
    )

    if !strings.HasPrefix(url, "file://") {        
        res, err := http.Get(url)
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
    filter := transform.Chain(enc.NewDecoder(), transform.RemoveFunc(func (r rune) bool {
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

func preprocess(items []Item){
    for ix := range items {
        item := &items[ix]
        item.Link  = html.UnescapeString(item.Link)
        item.Title = html.UnescapeString(item.Title)
        item.Body  = html.UnescapeString(item.Body)
    }
}

// items from RSS
func post_items(source string, items []Item, email string) {
    for _, rssItem := range items {
        rows, err := dbi.Query("select id, title, body, url from posts where recip=?", email)
        if err != nil {
            log.Println(err)
            return
        }
        defer rows.Close()

        match := false
        for rows.Next() {
            var (
                id int
                title string
                body string
                url string
            )
            err = rows.Scan(&id, &title, &body, &url)
            if err != nil {
                log.Println(err)
                return
            }
            dbItem := Item{ id, title, url, body }

            score := rate(rssItem, dbItem)

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

func rate(rssItem Item, dbItem Item) int {
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
    if rssItem.Link == dbItem.Link {
        score += 1
    }

    if score > 1 && rssItem.Link != dbItem.Link {
        // log.WithFields(log.Fields{
        //     "rssTitle": rssTitle,
        //     "rssBody": rssBody,
        //     "dbTitle": dbTitle,
        //     "dbBody": dbBody,
        // })
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

func increment(dbItem Item) {
    log.WithFields(log.Fields{
        "title": dbItem.Title,
        "id": dbItem.Id,
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

func send(source string, rssItem Item, recip string) {
        log.WithFields(log.Fields{
            "title": rssItem.Title,
        }).Info("Send")

        stmt, err := dbi.Prepare("INSERT INTO posts(source, recip, title, body, url) VALUES(?,?,?,?,?)")        
        if err != nil {
            log.Fatal(err)
        }

        res, err := stmt.Exec(source, recip, rssItem.Title, rssItem.Body, rssItem.Link)
        if err != nil {
            log.Fatal(err)
        }
        id,_ := res.LastInsertId()

        subject := fmt.Sprintf("Homedog #%d - %s", id, rssItem.Title)

        email(id, recip, source, subject, rssItem.Title, rssItem.Link, rssItem.Body)
}

func email(id int64, recip string, source string, subject string, title string, link string, body string) {
    if !*flag_email {
        return
    }

    auth := smtp.PlainAuth("", "key",
                               "secret",
                               "email-smtp.us-east-1.amazonaws.com")

    to := []string{"email@example.com", recip}

    type Post struct {
        ID int64
        Source string
        Subject string
        Title string
        Body string
        Link  string
    }
    post := Post{ id,
                  source,
                  html.UnescapeString(subject),
                  html.UnescapeString(title),
                  html.UnescapeString(body),
                  link }

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

    msg := []byte(doc.Bytes())
    err = smtp.SendMail("email-smtp.us-east-1.amazonaws.com:587", auth, "email@example.com", to, msg)
    if err != nil {
        log.Fatal(err)
    }
}
