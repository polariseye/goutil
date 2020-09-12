package imageUtil

import (
	"io"
	"os"
	"testing"

	"github.com/tuotoo/qrcode"
)

func TestCreateQrPngFile(t *testing.T) {
	val := "今天是个好天气"
	err := CreateQrPngFile(val, 256, 256, "qrCode.png")
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	flObj, err := os.Open("qrCode.png")
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	content, err := GetQrImageContent(flObj)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if val != content {
		t.Fatal("匹配失败 val:", val, "  content:", content)
		return
	}
}

func GetQrImageContent(readerObj io.Reader) (content string, err error) {
	var matrixObj *qrcode.Matrix
	matrixObj, err = qrcode.Decode(readerObj)
	if err != nil {
		return
	}

	content = matrixObj.Content
	return
}
