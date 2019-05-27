package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"hash"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// 生成code的最大长度
	MAX_PASSCODE_LENGTH int = 9

	// 默认时间间隔 /s
	INTERVAL int64 = 30

	// 默认code长度
	PASS_CODE_LENGTH int = 6

	// 检查前后间隔数量
	ADJACENT_INTERVALS int = 1
)

var (
	DIGITS_POWER = [MAX_PASSCODE_LENGTH + 1]int32{1, 10, 100, 1000, 10000, 100000, 1000000, 10000000, 100000000, 1000000000}
)

type Signer interface {
	Sign([]byte) []byte
}

type HmacSigner struct {
	h hash.Hash
}

func (hs *HmacSigner) Sign(data []byte) []byte {
	hs.h.Write(data)
	result := hs.h.Sum(nil)
	hs.h.Reset()
	return result
}

func NewHmacSigner(key []byte) Signer {
	mac := hmac.New(sha1.New, key)
	return &HmacSigner{h: mac}
}

type PasscodeGenerator struct {
	sign       Signer
	codeLength int
}

func NewPasscodeGenerator(sign Signer, passCodeLength int) (*PasscodeGenerator, error) {
	if passCodeLength < 0 || passCodeLength > MAX_PASSCODE_LENGTH {
		return nil, errors.New(fmt.Sprintf("PassCodeLength must be between 1 and %d digits.", MAX_PASSCODE_LENGTH))
	}

	return &PasscodeGenerator{
		sign:       sign,
		codeLength: passCodeLength,
	}, nil
}

func (pg *PasscodeGenerator) GenerateResponseCode(challenge int64) string {
	h := pg.sign.Sign(Int64ToByte(challenge))

	offset := int(h[len(h)-1] & 0xF)
	// 去掉符号位
	truncatedHash := HashToInt32(h, offset) & 0x7FFFFFFF
	result := truncatedHash % DIGITS_POWER[pg.codeLength]
	return pg.padOutput(result)
}

func (pg *PasscodeGenerator) VerifyTimeoutCode(timeoutCode string, currentInterval int64, pastIntervals, futureIntervals int) bool {
	if pastIntervals < 0 {
		pastIntervals = 0
	}
	if futureIntervals < 0 {
		futureIntervals = 0
	}

	for i := -pastIntervals; i <= futureIntervals; i++ {
		candidate := pg.GenerateResponseCode(currentInterval - int64(i))
		if candidate == timeoutCode {
			return true
		}
	}

	return false
}

func (pg *PasscodeGenerator) padOutput(value int32) string {

	return fmt.Sprintf("%0"+strconv.Itoa(pg.codeLength)+"d", value)
}

func HashToInt32(h []byte, offset int) int32 {

	return int32(binary.BigEndian.Uint32(h[offset : offset+4]))
}

func Int64ToByte(value int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(value))
	return buf
}

var accounts = make(map[string]*PasscodeGenerator, 10)

var file = flag.String("f", "./key.properties", "file path")

func initAccounts() {
	read, err := os.Open(*file)
	if err != nil {
		panic(err)
	}
	defer read.Close()
	scanner := bufio.NewScanner(read)
	i := 0
	for scanner.Scan() {
		i++
		pair := strings.TrimSpace(scanner.Text())
		if pair != "" {
			keyvalue := strings.Split(pair, "=")

			if len(keyvalue) < 2 {
				panic(fmt.Sprintf("error L:%d", i))
			}

			accountstr := strings.TrimSpace(keyvalue[0])
			keystr := strings.TrimSpace(keyvalue[1])

			if accountstr == "" || keystr == "" {
				panic(fmt.Sprintf("error L:%d", i))
			}

			keydata, e := base32.StdEncoding.DecodeString(keystr)
			if e != nil {
				panic(fmt.Sprintf("error base32 L:%d", i))
			}
			signer := NewHmacSigner(keydata)
			generator, e := NewPasscodeGenerator(signer, 6)
			if e != nil {
				panic(e)
			}

			accounts[accountstr] = generator
		}

	}
}

func main() {
	flag.Parse()

	initAccounts()

	if len(accounts) == 0 {
		return
	}

	t := time.Tick(time.Second)
	select {
	case n := <-t:
		fmt.Println(n)
	}

	flag := int64(0)

	for {
		select {
		case n := <-t:
			interval := n.Unix() / INTERVAL
			if interval != flag {
				fmt.Println("------------------------------------------")
				for k, v := range accounts {
					code := v.GenerateResponseCode(n.Unix() / INTERVAL)
					fmt.Println(fmt.Sprintf("%s\t%s", k, code))
				}
				flag = interval
				fmt.Println()
			}

		}
	}

}
