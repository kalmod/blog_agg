package main

import (
	"encoding/xml"
	"net/http"
)

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	Channel struct {
		Text  string `xml:",chardata"`
		Title string `xml:"title"`
		Link  struct {
			Text string `xml:",chardata"`
			Href string `xml:"href,attr"`
			Rel  string `xml:"rel,attr"`
			Type string `xml:"type,attr"`
		} `xml:"link"`
		Description   string `xml:"description"`
		Generator     string `xml:"generator"`
		Language      string `xml:"language"`
		LastBuildDate string `xml:"lastBuildDate"`
		Item          []struct {
			Text        string `xml:",chardata"`
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			PubDate     string `xml:"pubDate"`
			Guid        string `xml:"guid"`
			Description string `xml:"description"`
		} `xml:"item"`
	} `xml:"channel"`
}

func GetFeedDataFromUrl(url string) (Rss, error) {
	r, err := http.Get(url)
	if err != nil {
		return Rss{}, err
	}

	var rss Rss
	dec := xml.NewDecoder(r.Body)
	err = dec.Decode(&rss)
	if err != nil {
		return Rss{}, err
	}

	return rss, nil
}
