package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"

	"github.com/andyleap/microformats"
	"github.com/sivy/goldfrog/pkg/webmention"
	"gopkg.in/yaml.v2"
)

var version string // set in linker with ldflags -X main.version=

func findEndpoint(mentionTarget string) string {
	client := webmention.NewWebMentionClient()

	endpoint, err := client.EndpointDiscovery(mentionTarget)
	if err != nil {
		fmt.Printf("Could not fetch endpoint: %s", err)
	}
	return endpoint
}

func sendMention(source string, target string) {
	client := webmention.NewWebMentionClient()
	endpoint, err := client.EndpointDiscovery(target)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	client.SendMention(endpoint, source, target)
}

func findMF(target string, format string) {
	parser := microformats.New()
	resp, err := http.Get(target)
	if err != nil {
		fmt.Println(err)
		return
	}
	urlparsed, _ := url.Parse(target)
	data := parser.Parse(resp.Body, urlparsed)

	var marshalled []byte

	if format == "yaml" {
		marshalled, _ = yaml.Marshal(data)
	} else if format == "json" {
		marshalled, _ = json.MarshalIndent(data, "", "  ")
	}

	fmt.Println(string(marshalled))
}

func main() {

	var target string
	var source string
	var send bool
	var parseTarget string
	var parseOutputFormat string

	flag.StringVar(&target, "target", "", "")
	flag.StringVar(&source, "source", "", "")

	flag.StringVar(&parseTarget, "parse-wm", "", "")
	flag.StringVar(&parseOutputFormat, "fmt", "", "")

	flag.BoolVar(&send, "send", false, "")

	flag.Parse()

	fmt.Println(fmt.Sprintf("Checking %s", target))

	endpoint := findEndpoint(target)
	fmt.Println(endpoint)

	if send {
		sendMention(source, target)
	}

	if parseTarget != "" {
		findMF(parseTarget, parseOutputFormat)
	}
}
