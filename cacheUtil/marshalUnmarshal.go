package cacheUtil

import (
	"fmt"
	"reflect"
	"strconv"
)

type Marshaler interface {
	Marshal(val interface{}) (bytesData []byte, err error)
}

type marshalBase struct {
	actualMarshaler Marshaler
}

func (m *marshalBase) Marshal(val interface{}) (bytesData []byte, err error) {
	valTp := reflect.ValueOf(val)
	for valTp.Kind() != reflect.Ptr {
		valTp = valTp.Elem()
	}

	switch valTp.Kind() {
	case reflect.Array, reflect.Map, reflect.Struct:
		return m.Marshal(val)
	case reflect.Int, reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		resultStr := fmt.Sprintf("%v", val)
		bytesData = []byte(resultStr)
		return
	default:
		err = fmt.Errorf("not supported type:%v", valTp.Kind().String())
		return
	}
}

type Unmarshaler interface {
	Unmarshal(bytesData []byte, val interface{}) error
}

type unmarshalBase struct {
	actualunmarshaler Unmarshaler
}

func (u *unmarshalBase) Unmarshal(bytesData []byte, val interface{}) error {
	valTp := reflect.ValueOf(val)
	for valTp.Kind() != reflect.Ptr {
		if valTp.IsNil() {
			valTp.Set(reflect.New(valTp.Type()))
		}
		valTp = valTp.Elem()
	}

	switch valTp.Kind() {
	case reflect.Array, reflect.Map, reflect.Struct:
		return u.Unmarshal(bytesData, val)
	case reflect.Bool:
		{
			tmpResult, err := strconv.ParseBool(string(bytesData))
			if err != nil {
				return err
			}
			valTp.Set(reflect.ValueOf(tmpResult))
		}
	case reflect.Int:
		{
			tmpResult, err := strconv.ParseInt(string(bytesData), 10, 32)
			if err != nil {
				return err
			}
			valTp.Set(reflect.ValueOf(int(tmpResult)))
		}
	case reflect.Int8:
		{
			tmpResult, err := strconv.ParseInt(string(bytesData), 10, 8)
			if err != nil {
				return err
			}
			valTp.Set(reflect.ValueOf(int8(tmpResult)))
		}
	case reflect.Int16:
		{
			tmpResult, err := strconv.ParseInt(string(bytesData), 10, 16)
			if err != nil {
				return err
			}
			valTp.Set(reflect.ValueOf(int16(tmpResult)))
		}
	case reflect.Int32:
		{
			tmpResult, err := strconv.ParseInt(string(bytesData), 10, 32)
			if err != nil {
				return err
			}
			valTp.Set(reflect.ValueOf(int32(tmpResult)))
		}
	case reflect.Int64:
		{
			tmpResult, err := strconv.ParseInt(string(bytesData), 10, 64)
			if err != nil {
				return err
			}
			valTp.Set(reflect.ValueOf(int64(tmpResult)))
		}
	case reflect.Uint:
		{
			tmpResult, err := strconv.ParseUint(string(bytesData), 10, 32)
			if err != nil {
				return err
			}
			valTp.Set(reflect.ValueOf(uint(tmpResult)))
		}
	case reflect.Uint16:
		{
			tmpResult, err := strconv.ParseUint(string(bytesData), 10, 16)
			if err != nil {
				return err
			}
			valTp.Set(reflect.ValueOf(uint16(tmpResult)))
		}
	case reflect.Uint32:
		{
			tmpResult, err := strconv.ParseUint(string(bytesData), 10, 32)
			if err != nil {
				return err
			}
			valTp.Set(reflect.ValueOf(uint32(tmpResult)))
		}
	case reflect.Uint64:
		{
			tmpResult, err := strconv.ParseUint(string(bytesData), 10, 64)
			if err != nil {
				return err
			}
			valTp.Set(reflect.ValueOf(uint64(tmpResult)))
		}
	case reflect.Float32:
		{
			tmpResult, err := strconv.ParseFloat(string(bytesData), 32)
			if err != nil {
				return err
			}
			valTp.Set(reflect.ValueOf(float32(tmpResult)))
		}
	case reflect.Float64:
		{
			tmpResult, err := strconv.ParseFloat(string(bytesData), 32)
			if err != nil {
				return err
			}
			valTp.Set(reflect.ValueOf(float64(tmpResult)))
		}
	default:
		return fmt.Errorf("not supported type:%v", valTp.Kind().String())
	}

	return nil
}

// NewMarshal extend base marshal func when not implement base data type
func NewMarshal(marshalObj Marshaler) Marshaler {
	return &marshalBase{
		actualMarshaler: marshalObj,
	}
}

// NewUnmarhsal extend base marshal func when not implement base data type
func NewUnmarhsal(unmarshalObj Unmarshaler) Unmarshaler {
	return &unmarshalBase{
		actualunmarshaler: unmarshalObj,
	}
}
