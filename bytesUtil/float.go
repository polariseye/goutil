package bytesUtil

import (
	"encoding/binary"
	"math"
)

// 浮点转换成字节
// n：float32型数字
// order：大、小端的枚举
// 返回值：对应的字节数组
func Float32ToByte(val float32, order binary.ByteOrder) []byte {
	bits := math.Float32bits(val)
	bytes := make([]byte, 4)
	order.PutUint32(bytes, bits)
	return bytes
}

//浮点转换成字节
// n：float64型数字
// order：大、小端的枚举
// 返回值：对应的字节数组
func Float64ToByte(val float64, order binary.ByteOrder) []byte {
	bits := math.Float64bits(val)
	bytes := make([]byte, 8)
	order.PutUint64(bytes, bits)
	return bytes
}
