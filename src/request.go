package main

import (
	"os"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"time"
	"regexp"
	"math"
	"crypto/tls"
	"compress/gzip"
	"bytes"
)

func getRequestHeader(u *URL, cookie string) (string) {
	get := "GET " + u.Path + " HTTP/1.1\r\n"
	os.Stderr.WriteString(get)  // record startline of request

	get += "Host: " + u.Host + "\r\n"
	get += "Connection: keep-alive\r\n"
	get += "Pragma: no-cache\r\n"
	get += "Cache-Control: no-cache\r\n"
	get += "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8\r\n"
	get += "Accept-Encoding: gzip"
	get += "Accept-Language: zh-CN,zh;q=0.8,en;q=0.6\r\n"
	get += "Set-Cookie: " + cookie + "\r\n"
	get += "\r\n"
	get += u.Query + "\r\n"

	return get
}

func hexStringTrans(target string) (int) {
	ans := 0
	lens := len(target)
	for i := 0; i < lens; i++ {
		cur := int(target[i])
		if cur >= 48 && cur <= 57 {
			ans += int(math.Pow(16, float64(lens - i - 1))) * (cur - 48)
		} else if cur >= 97 && cur <= 102 {
			ans += int(math.Pow(16, float64(lens - i - 1))) * (cur - 87)
		} else if cur == 13 || cur == 10 {
			continue;
		} else {
			return 0
		}
	}

	return ans
}

func sendRequest(u *URL, hostIP string, cookie string) {
	var (
		conn net.Conn
		err error
	)

	if u.Port != "80" && u.Port != "443" {
		u.Host += ":" + u.Port
	}

	if u.Scheme == "http" {  // http scheme
		conn, err = net.DialTimeout("tcp", hostIP + ":" + u.Port, 5000000000)
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(0)
		}
		defer conn.Close()
	} else if u.Scheme == "https" {  // https scheme
		dialer := &net.Dialer{Timeout: 5000000000}
		conn, err = tls.DialWithDialer(dialer, "tcp", u.Host + ":" + u.Port, nil)
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(0)
		}
	} else {
		os.Stderr.WriteString("Error: Unknown Scheme!\n")
		os.Exit(0)
	}

	get := getRequestHeader(u, cookie)

	if _, err := conn.Write([]byte(get)); err != nil {
		os.Stderr.WriteString(err.Error())  
		os.Exit(0)
	}
	var answer = make([]byte, 4096)
	isDelayed := 0
	i, err := conn.Read(answer)
	if err != nil { 
		os.Stderr.WriteString(err.Error())
		os.Exit(0)
	}
	regChunked := regexp.MustCompile("Transfer-Encoding: chunked")
	regGziped := regexp.MustCompile("Content-Encoding: gzip")
	redirect1 := regexp.MustCompile("HTTP/1.1 301")
	redirect2 := regexp.MustCompile("HTTP/1.1 302")
	chunked := len(regChunked.FindStringSubmatch(string(answer))) > 0
	gziped := len(regGziped.FindStringSubmatch(string(answer))) > 0
	redirected := len(redirect1.FindStringSubmatch(string(answer))) > 0 || len(redirect2.FindStringSubmatch(string(answer))) > 0

	if redirected {  // handle redirect
		newUrl := strings.Fields(strings.Split(string(answer), "Location: ")[1])[0]

		// handle cookie
		cookie := ""
		cookieReg := regexp.MustCompile("Cookie: ")
		if len(cookieReg.FindStringSubmatch(string(answer))) > 0 {
			cookie = strings.Fields(strings.Split(string(answer), "Cookie: ")[1])[0]
		}
		
		// get parsed-url
		u := Parse(newUrl)

		// make the dns-query
		hostIP := SendDNSQuery(u.Host, "202.120.224.26:53")
		os.Stderr.WriteString(hostIP)

		// send request
		sendRequest(u, hostIP, cookie)
		os.Exit(0)
	}

	if chunked == true {  // chunked transfer encoding

		recordLength := 0
		count := len(strings.Split(string(answer[:i]), "\r\n\r\n"))
		os.Stderr.WriteString(strings.Split(string(answer[:i]), "\n")[0] + "\n")
		if count <= 1 {
			isDelayed = 1
		}
		if count >= 2 {
			count1 := len(strings.Split(string(answer[:i]), "\r\n\r\n")[0] + "\r\n\r\n")
			if gziped {
				buffer := bytes.NewBuffer(answer)
				r, err := gzip.NewReader(buffer)
				if err != nil {
					os.Stderr.WriteString(err.Error())
					os.Exit(0)
				}
				defer r.Close()
				answer, _ = ioutil.ReadAll(r)
			}
			for ; count1 < i; {
				if recordLength > 0 {
					if i - count1 < recordLength {
						recordLength = recordLength - i + count1
						os.Stdout.WriteString(string(answer[count1:i]))
						count1 = i
					} else {
						os.Stdout.WriteString(string(answer[count1:(count1+recordLength)]))
						count1 += (recordLength + 2)
						recordLength = 0
					}
				}
				if count1 >= i - 2 {
					break
				}
				recordLength = hexStringTrans(strings.Split(string(answer[count1:i]), "\r\n")[0])
				if recordLength == 0 {
					os.Exit(0)
				}
				count1 += len(strings.Split(string(answer[count1:i]), "\r\n")[0] + "\r\n")
			}
		}

		// parsing the response data
		count = 0
		var answer1 = make([]byte, 100000)
		conn.SetReadDeadline((time.Now().Add(time.Second * 5)))
		i1, err := conn.Read(answer1);
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(0)
		}
		if gziped {
			buffer := bytes.NewBuffer(answer1)
			r, err := gzip.NewReader(buffer)
			if err != nil {
				os.Stderr.WriteString(err.Error())
				os.Exit(0)
			}
			defer r.Close()
			answer1, _ = ioutil.ReadAll(r)
		}
		if isDelayed > 0 && len(strings.Split(string(answer1[:i1]), "\r\n\r\n")) > 1 {
			count = len(strings.Split(string(answer1[:i1]), "\r\n\r\n")[0] + "\r\n\r\n")
		}
		for ; count < i1; {
			if recordLength > 0 {
				if i1 - count < recordLength {
					recordLength = recordLength - i1 + count
					os.Stdout.WriteString(string(answer1[count:i1]))
					count = i1
				} else {
					os.Stdout.WriteString(string(answer1[count:(count+recordLength)]))
					count += (recordLength + 2)
					recordLength = 0
				}
			}
			if count >= i1 - 2 {
				break
			}
			recordLength = hexStringTrans(strings.Split(string(answer1[count:i1]), "\r\n")[0])
			if recordLength == 0 {
				os.Exit(0)
			}
			count += len(strings.Split(string(answer1[count:i1]), "\r\n")[0] + "\r\n")
		}
		
		for {
			conn.SetReadDeadline((time.Now().Add(time.Second * 5)))
			i1, err := conn.Read(answer1);
			if err != nil {
				if err != io.EOF {
					os.Stderr.WriteString(err.Error())
					os.Exit(0)
				}
				break
			}
			if gziped {
				buffer := bytes.NewBuffer(answer1)
				r, err := gzip.NewReader(buffer)
				if err != nil {
					os.Stderr.WriteString(err.Error())
					os.Exit(0)
				}
				defer r.Close()
				answer1, _ = ioutil.ReadAll(r)
			}
			count = 0
			for ; count < i1; {
				if recordLength > 0 {
					if i1 - count < recordLength {
						recordLength = recordLength - i1 + count
						os.Stdout.WriteString(string(answer1[count:i1]))
						count = i1
					} else {
						os.Stdout.WriteString(string(answer1[count:(count+recordLength)]))
						count += (recordLength + 2)
						recordLength = 0
					}
				}
				if count >= i1 - 2 {
					break
				}

				recordLength = hexStringTrans(strings.Split(string(answer1[count:i1]), "\r\n")[0])
				if recordLength == 0 {
					os.Exit(0)
				}
				count += len(strings.Split(string(answer1[count:i1]), "\r\n")[0] + "\r\n")
			}
		}

	} else {  // normal transfer

		r1 := regexp.MustCompile(`Content-Length: \S+\r\n`)
		r2 := regexp.MustCompile("(Content-Length: [0-9]+\r\n)")

		if gziped {
			buffer := bytes.NewBuffer(answer)
			r, err := gzip.NewReader(buffer)
			if err != nil {
				os.Stderr.WriteString(err.Error())
				os.Exit(0)
			}
			defer r.Close()
			answer, _ = ioutil.ReadAll(r)
		}
		
		// record the response startline
		count := len(strings.Split(string(answer[:i]), "\r\n\r\n"))
		if count <= 1 {
			isDelayed = 1
		}
		if count >= 2 {
			if len(r1.FindStringSubmatch(string(answer[:i]))) > 0 && len(r2.FindStringSubmatch(string(answer[:i]))) == 0 {
				os.Exit(0)
			}
			os.Stdout.WriteString(strings.Join(strings.Split(string(answer[:i]), "\r\n\r\n")[1:], "\r\n\r\n"))
		}
		os.Stderr.WriteString(strings.Split(string(answer[:i]), "\n")[0] + "\n")

		// parsing the response data
		var answer1 = make([]byte, 100000)
		conn.SetReadDeadline((time.Now().Add(time.Second * 5)))
		i1, err := conn.Read(answer1);
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(0)
		}
		if gziped {
			buffer := bytes.NewBuffer(answer1)
			r, err := gzip.NewReader(buffer)
			if err != nil {
				os.Stderr.WriteString(err.Error())
				os.Exit(0)
			}
			defer r.Close()
			answer1, _ = ioutil.ReadAll(r)
		}
		if isDelayed > 0 && len(strings.Split(string(answer1[:i1]), "\r\n\r\n")) > 1 {
			if len(r1.FindStringSubmatch(string(answer1[:i1]))) > 0 && len(r2.FindStringSubmatch(string(answer1[:i1]))) == 0 {
				os.Exit(0)
			}
			os.Stdout.WriteString(strings.Join(strings.Split(string(answer1[:i1]), "\r\n\r\n")[1:], "\r\n\r\n"))
		} else {
			os.Stdout.WriteString(string(answer1[:i1]))
		}
		for {
			conn.SetReadDeadline((time.Now().Add(time.Second * 5)))
			i1, err := conn.Read(answer1);
			if err != nil {
				if err != io.EOF {
					os.Stderr.WriteString(err.Error())
					os.Exit(0)
				}
				break
			}
			if gziped {
				buffer := bytes.NewBuffer(answer1)
				r, err := gzip.NewReader(buffer)
				if err != nil {
					os.Stderr.WriteString(err.Error())
					os.Exit(0)
				}
				defer r.Close()
				answer1, _ = ioutil.ReadAll(r)
			}
			os.Stdout.WriteString(string(answer1[:i1]))
		}
	}

}