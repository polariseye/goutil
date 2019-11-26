package googleAuthUtil

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

/*
	func:基于当前时间生成一个密钥
	returns:
		secret:密钥
		err:错误信息
*/
func GetSecret() (secret string, err error) {
	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, nowUnix())
	if err != nil {
		return
	}

	secret = strings.ToUpper(base32encode(hmacSha1(buf.Bytes(), nil)))
	return
}

/*
	func:获取当前时间的授权码
	parameter:
		secret: 密钥
	返回值:
		string: 当前授权码
		error: 错误信息
*/
func GetNowAuthCode(secret string) (string, error) {
	secretUpper := strings.ToUpper(secret)
	secretKey, err := base32decode(secretUpper)
	if err != nil {
		return "", err
	}
	number := oneTimePassword(secretKey, toBytes(time.Now().Unix()/30))
	return fmt.Sprintf("%06d", number), nil
}

/*
	func:获取密钥对应的二维码地址
	parameter:
		user:用户Id
		secret:密钥
	returns:
		string:密钥对应的二维码地址
*/
func GetQrcode(user, secret string) string {
	return fmt.Sprintf("otpauth://totp/%s?secret=%s", user, secret)
}

/*
	func:获取谷歌的二维码URL
	parameter:
		user:用户Id
		secret:密钥
	returns:
		string:谷歌的二维码URL
*/
func GetQrcodeUrl(user, secret string) string {
	qrcode := GetQrcode(user, secret)
	return fmt.Sprintf("http://www.google.com/chart?chs=200x200&chld=M%%7C0&cht=qr&chl=%s", qrcode)
}

/*
	func:验证极权码是否正确
	parameter:
		secret:密钥
		code:极权码
	returns:
		bool：是否授权成功
		error:错误信息
*/
func VerifyCode(secret, code string) (bool, error) {
	_code, err := GetNowAuthCode(secret)
	fmt.Println(_code, code, err)
	if err != nil {
		return false, err
	}
	return _code == code, nil
}

func nowUnix() int64 {
	return time.Now().Unix() / 30
}

func hmacSha1(key, data []byte) []byte {
	h := hmac.New(sha1.New, key)
	if total := len(data); total > 0 {
		h.Write(data)
	}
	return h.Sum(nil)
}

func base32encode(src []byte) string {
	return base32.StdEncoding.EncodeToString(src)
}

func base32decode(s string) ([]byte, error) {
	return base32.StdEncoding.DecodeString(s)
}

func toBytes(value int64) []byte {
	var result []byte
	mask := int64(0xFF)
	shifts := [8]uint16{56, 48, 40, 32, 24, 16, 8, 0}
	for _, shift := range shifts {
		result = append(result, byte((value>>shift)&mask))
	}
	return result
}

func toUint32(bts []byte) uint32 {
	return (uint32(bts[0]) << 24) + (uint32(bts[1]) << 16) +
		(uint32(bts[2]) << 8) + uint32(bts[3])
}

func oneTimePassword(key []byte, data []byte) uint32 {
	hash := hmacSha1(key, data)
	offset := hash[len(hash)-1] & 0x0F
	hashParts := hash[offset : offset+4]
	hashParts[0] = hashParts[0] & 0x7F
	number := toUint32(hashParts)
	return number % 1000000
}
