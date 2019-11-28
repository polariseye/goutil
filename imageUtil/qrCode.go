package imageUtil

import (
	"bytes"
	"image/png"
	"os"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

/*
	func:创建一个二维码图片
	parameters:
		data:二维码图片的数据
		width:图片宽度
		height:图片高度
	returns:
		pngBytes:图片字节
		err:错误信息
*/
func GetQrCodePng(data string, width int, height int) (pngBytes []byte, err error) {
	var qrCodeImg barcode.Barcode
	qrCodeImg, err = qr.Encode(data, qr.M, qr.Auto)
	if err != nil {
		return
	}

	qrCodeImg, err = barcode.Scale(qrCodeImg, width, height)
	if err != nil {
		return
	}

	pngBuffer := &bytes.Buffer{}
	err = png.Encode(pngBuffer, qrCodeImg)
	if err != nil {
		return
	}

	return
}

/*
	func:创建一个二维码图片文件
	parameters:
		data:二维码图片的数据
		width:图片宽度
		height:图片高度
		filePath:图片存储位置
	returns:
		pngBytes:图片字节
		err:错误信息
*/
func CreateQrPngFile(data string, width int, height int, filePath string) (err error) {
	var qrCodeImg barcode.Barcode
	qrCodeImg, err = qr.Encode(data, qr.M, qr.Auto)
	if err != nil {
		return
	}

	qrCodeImg, err = barcode.Scale(qrCodeImg, width, height)
	if err != nil {
		return
	}

	fileObj, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fileObj.Close()

	err = png.Encode(fileObj, qrCodeImg)
	if err != nil {
		return
	}

	return
}
