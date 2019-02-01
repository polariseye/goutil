package bytesUtil

import (
	"bytes"
	"encoding/binary"
	"math"
)

// 字节数组转换成整形
// b：字节数组
// order：大、小端的枚举
// 返回值：对应的int值
func BytesToInt(b []byte, order binary.ByteOrder) int {
	bytesBuffer := bytes.NewBuffer(b)

	var result int
	binary.Read(bytesBuffer, order, &result)

	return result
}

// 字节数组转换成整形
// b：字节数组
// order：大、小端的枚举
// 返回值：对应的int16值
func BytesToInt16(b []byte, order binary.ByteOrder) int16 {
	bytesBuffer := bytes.NewBuffer(b)

	var result int16
	binary.Read(bytesBuffer, order, &result)

	return result
}

// 字节数组转换成整形
// b：字节数组
// order：大、小端的枚举
// 返回值：对应的int32值
func BytesToInt32(b []byte, order binary.ByteOrder) int32 {
	bytesBuffer := bytes.NewBuffer(b)

	var result int32
	binary.Read(bytesBuffer, order, &result)

	return result
}

// 字节数组转换成整形
// b：字节数组
// order：大、小端的枚举
// 返回值：对应的int64值
func BytesToInt64(b []byte, order binary.ByteOrder) int64 {
	bytesBuffer := bytes.NewBuffer(b)

	var result int64
	binary.Read(bytesBuffer, order, &result)

	return result
}

// 字节数组转换成整形
// b：字节数组
// order：大、小端的枚举
// 返回值：对应的float32值
func BytesToFloat32(b []byte, order binary.ByteOrder) float32 {
	bits := order.Uint32(b)
	return math.Float32frombits(bits)
}

// 字节数组转换成整形
// b：字节数组
// order：大、小端的枚举
// 返回值：对应的float64值
func BytesToFloat64(b []byte, order binary.ByteOrder) float64 {
	bits := order.Uint64(b)
	return math.Float64frombits(bits)
}
