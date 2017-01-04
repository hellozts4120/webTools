package main

import (
	"os"
)

func main() {
	readURL()
}

func readURL() {
	if len(os.Args) < 2 {
		os.Stderr.WriteString("Error: You must input URL!\n")
		os.Exit(0)
	}

	if len(os.Args) > 2 {
		os.Stderr.WriteString("Error: You can only input one URL!\n")
		os.Exit(0)
	}

	rawURL := os.Args[1]

	// get parsed-url
	u := Parse(rawURL)

	// make the dns-query
	hostIP := SendDNSQuery(u.Host, "202.120.224.26:53")
	os.Stderr.WriteString(hostIP)

	// send request
	sendRequest(u, hostIP, "")
}