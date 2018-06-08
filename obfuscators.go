// Scramble functions.
// Input `s []byte` is required to be not nil.
package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"net"
        "encoding/json"
        "fmt"
)

var Salt []byte

const emailDomain = "@example.com"

func GenScrambleBytes(maxLength uint) func([]byte) []byte {
	return func(s []byte) []byte {
		// TODO: pad or extend s to maxLength
		return ScrambleBytes(s)[:maxLength]
	}
}

var bytesOutputAlphabetLength = byte(len(bytesOutputAlphabet))
var bytesKeep = []byte("',\\{}")
var bytesOutputAlphabet = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+-_")

// Modifies `s` in-place.
func ScrambleBytes(s []byte) []byte {
        fmt.Printf("Scramble bytes %v\n", string(s))
        isArray := len(s) >= 2 && s[0] == '{' && s[len(s)-1] == '}'

	hash := sha256.New()
	// Hard-coding this constant wins less than 3% in BenchmarkScrambleBytes
	const sumLength = 32 // SHA256/8
	hash.Write(Salt)
	hash.Write(s)
	sumBytes := hash.Sum(nil)

	reader := bytes.NewReader(s)
        var r rune
	var err error
	for i := 0; ; i++ {
		r, _, err = reader.ReadRune()
		if err != nil {
			s = s[:i]
			break
		}
		if !isArray || bytes.IndexRune(bytesKeep, r) == -1 {
			// Do not insert, so should not obstruct reader.
			s[i] = bytesOutputAlphabet[(sumBytes[i%sumLength]+byte(r))%bytesOutputAlphabetLength]
		} else {
			// Possibly shift bytes to beginning of s.
			s[i] = byte(r)
		}
	}
	return s
}

func ScrambleDigits(s []byte) []byte {
        fmt.Printf("Scramble Digits %v\n", s)
	hash := sha256.New()
	const sumLength = 32 // SHA256/8
	hash.Write(Salt)
	hash.Write(s)
	sumBytes := hash.Sum(nil)

	for i, b := range s {
		if b >= '0' && b <= '9' {
			s[i] = '0' + (sumBytes[i%sumLength]+b)%10
		}
	}
	return s
}

func scrambleOneEmail(s []byte) []byte {
	atIndex := bytes.IndexRune(s, '@')
	mailbox := Salt
	if atIndex != -1 {
		mailbox = s[:atIndex]
	}
	s = make([]byte, len(mailbox)+len(emailDomain))
	copy(s, mailbox)
	// ScrambleBytes is in-place, but may return string shorter than input.
	mailbox = ScrambleBytes(s[:len(mailbox)])
	copy(s[len(mailbox):], emailDomain)
	// So final len(mailbox) may be shorter than whole allocated string.
	return s[:len(mailbox)+len(emailDomain)]
}

// Supports array of emails in format {email1,email2}
func ScrambleEmail(s []byte) []byte {
        fmt.Printf("Scramble email %v\n", string(s))
	if len(s) < 2 {
		// panic("ScrambleEmail: input is too small: '" + string(s) + "'")
		return s
	}
	if s[0] != '{' && s[len(s)-1] != '}' {
		return scrambleOneEmail(s)
	}
	parts := bytes.Split(s[1:len(s)-1], []byte{','})
	partsNew := make([][]byte, len(parts))
	outLength := 2 + len(parts) - 1
	for i, bs := range parts {
		partsNew[i] = scrambleOneEmail(bs)
		outLength += len(partsNew[i])
	}
	s = make([]byte, outLength)
	s[0] = '{'
	s[len(s)-1] = '}'
	copy(s[1:len(s)-1], bytes.Join(partsNew, []byte{','}))
	return s
}

func ScrambleInet(s []byte) []byte {
	hash := sha256.New()
	const sumLength = 32 // SHA256/8
	hash.Write(Salt)
	hash.Write(s)
	sumBytes := hash.Sum(nil)

	var ip net.IP
	// Decide to output IPv4 or IPv6 based on first bit of hash sum.
	// Gives about 50% chance to each option.
	if sumBytes[0]&0x80 != 0 {
		ip = net.IP(sumBytes[:16])
	} else {
		ip = net.IP(sumBytes[:4])
	}
	return []byte(ip.String())
}

func ScrambleJson(s []byte) []byte {
        var m map[string]interface{}
        err := json.Unmarshal(s, &m)
        if err != nil {
                panic(err)
        }

        r := scrambleJsonObject(m)
        payload, _ := json.Marshal(r)
        return payload
}

func scrambleJsonObject(m map[string]interface{}) map[string]interface{} {
    for k, v := range m {
        fmt.Printf("PRE : %v - Value %v of %T\n", k , v, v)
        if mv, ok := v.(map[string]interface{}); ok {
            m[k] = scrambleJsonObject(mv)
            fmt.Printf("POST - Key : %v - Value %v\n", string(k))
        } else {
            if ov, ok := v.(string); ok {
                m[k] = ScrambleBytes([]byte(ov))
                fmt.Printf("POST - Key : %v - Value %v\n", string(k))
            }
        }
    }
    fmt.Printf("Map : %v\n", m)
    return m
}

func init() {
	Salt = make([]byte, 16)
	_, err := rand.Read(Salt)
	if err != nil {
		panic(err)
	}
}
