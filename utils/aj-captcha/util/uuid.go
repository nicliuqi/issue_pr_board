package util

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"issue_pr_board/config"
	rand2 "math/rand"
	"time"
)

func GetUuid() string {
	b := make([]byte, 16)
	io.ReadFull(rand.Reader, b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func RandString(codeLen int) string {
	rawStr := config.AppConfig.RandRawString
	buf := make([]byte, 0, codeLen)
	b := bytes.NewBuffer(buf)
	rand2.Seed(time.Now().UnixNano())
	for rawStrLen := len(rawStr); codeLen > 0; codeLen-- {
		randNum := rand2.Intn(rawStrLen)
		b.WriteByte(rawStr[randNum])
	}
	return b.String()
}
