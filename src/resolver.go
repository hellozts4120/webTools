package main

import (
	"os"
	"bytes"
	"encoding/binary"
	"net"
	"strings"
	"strconv"
)

type DNSHeader struct {
	ID            uint16
	Flag          uint16
	QuestionCount uint16
	AnswerRRs     uint16
	AuthorityRRs  uint16
	AdditionalRRs uint16
}

type DNSQuery struct {
	QuestionType     uint16
	QuestionClass    uint16
}

func (header *DNSHeader) SetFlag(QR uint16, OperationCode uint16, AuthoritativeAnswer uint16, Truncation uint16, RecursionDesired uint16, RecursionAvailable uint16, ResponseCode uint16) {
	header.Flag = QR<<15 + OperationCode<<11 + AuthoritativeAnswer<<10 + Truncation<<9 + RecursionDesired<<8 + RecursionAvailable<<7 + ResponseCode
}

func ParseDomainName(domain string) []byte {
	// parsing domain to right format
	var (
		buffer   bytes.Buffer
		segments []string = strings.Split(domain, ".")
	)
	for _, seg := range segments {
		binary.Write(&buffer, binary.BigEndian, byte(len(seg)))
		binary.Write(&buffer, binary.BigEndian, []byte(seg))
	}
	binary.Write(&buffer, binary.BigEndian, byte(0x00))

	return buffer.Bytes()
}

func SendDNSQuery(host string, server string) (string) {
	var (
		dns_header   DNSHeader
		dns_query    DNSQuery
	)

	dns_header.ID = 0xFFFF
	dns_header.SetFlag(0, 0, 0, 0, 1, 0, 0)
	dns_header.QuestionCount = 1
	dns_header.AnswerRRs = 0
	dns_header.AuthorityRRs = 0
	dns_header.AdditionalRRs = 0

	dns_query.QuestionType = 1  //IPv4  
	dns_query.QuestionClass = 1

	var (  
		conn net.Conn
		err  error

	    buffer bytes.Buffer
	    answer = make([]byte, 2048)
	)

	if conn, err = net.DialTimeout("udp", server, 5000000000); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(0)
	}  
	defer conn.Close()  

	binary.Write(&buffer, binary.BigEndian, dns_header)
	binary.Write(&buffer, binary.BigEndian, ParseDomainName(host))
	binary.Write(&buffer, binary.BigEndian, dns_query)

	if _, err := conn.Write(buffer.Bytes()); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(0)
	}
	i, err := conn.Read(answer);
	if err != nil { 
		os.Stderr.WriteString(err.Error())
		os.Exit(0)
	}
	return decodeDNSResponse(answer, i, len(buffer.Bytes()))
}

func decodeDNSResponse(answer []byte, num int, startPos int) (string) {
	/*
	 * 0 -> 11: Response Header
	 * 12 -> (startPos - 1): Question
	 * startPos -> END: Answer
	 */

	if answer[3] != 128 {
		// RCODE not equal to 0, must exist error
		os.Exit(0)
	}

	ans := 0
	for i:= startPos; i < num; i++{
		if answer[i] == uint8(0) && answer[i+1] == uint8(1) && answer[i+2] == uint8(0) && answer[i+3] == uint8(1) && answer[i+8] == uint8(0) && answer[i+9] == uint8(4) {
			ans = i+10
			break
		}
	}
	if ans == 0 {
		os.Exit(0)
	}

	res := []string{strconv.Itoa(int(answer[ans])), strconv.Itoa(int(answer[ans+1])), strconv.Itoa(int(answer[ans+2])), strconv.Itoa(int(answer[ans+3]))};
	return strings.Join(res, ".")


	// fmt.Println(ans+1);
	// fmt.Println(answer[:num])
}

