package main

import (
	"fmt"
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

	fmt.Println(rawURL)
	u := Parse(rawURL)
	fmt.Println(u.Scheme)
    fmt.Println(u.UserInfo)

    fmt.Println(u.Host)
    fmt.Println(u.Port)
    fmt.Println(u.Fragment)
    fmt.Println(u.Authority)
    fmt.Println(u.Path)

	if u.Scheme == "http" {

	} else if u.Scheme == "https" {

	} else {
		os.Stderr.WriteString("Error: Unknown Scheme!\n")
		os.Exit(-1)
	}
}