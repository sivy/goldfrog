package main

import (
	"flag"
	"fmt"

	"github.com/sivy/goldfrog/pkg/webmention"
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

func main() {

	var target string
	var source string
	var send bool

	flag.StringVar(&target, "target", "", "")
	flag.StringVar(&source, "source", "", "")
	flag.BoolVar(&send, "send", false, "")

	flag.Parse()

	fmt.Println(fmt.Sprintf("Checking %s", target))

	endpoint := findEndpoint(target)
	fmt.Println(endpoint)

	if send {
		sendMention(source, target)
	}
}
