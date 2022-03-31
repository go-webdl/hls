package hls

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrWrongType = errors.New("consuming the value as a wrong data type")

type Type string

const (
	StringType     Type = "String"
	EnumType       Type = "Enum"
	IntegerType    Type = "Integer"
	FloatType      Type = "Float"
	BytesType      Type = "Bytes"
	ResolutionType Type = "Resolution"
)

type Value struct {
	Type            Type
	StringValue     *string
	EnumValue       *string
	IntegerValue    *int64
	FloatValue      *float64
	BytesValue      []byte
	ResolutionValue *Resolution
}

type Resolution struct {
	Width  int
	Height int
}

func (r Resolution) Format() string {
	return fmt.Sprintf("%dx%d", r.Width, r.Height)
}

func String(value string) *Value {
	return &Value{Type: StringType, StringValue: &value}
}

func (v *Value) String() (value string, err error) {
	if v.Type == StringType && v.StringValue != nil {
		value = *v.StringValue
	} else {
		err = fmt.Errorf("consuming %s value as String: %w", string(v.Type), ErrWrongType)
	}
	return
}

func Enum(value string) *Value {
	return &Value{Type: EnumType, EnumValue: &value}
}

func (v *Value) Enum() (value string, err error) {
	if v.Type == EnumType && v.EnumValue != nil {
		value = *v.EnumValue
	} else {
		err = fmt.Errorf("consuming %s value as Enum: %w", string(v.Type), ErrWrongType)
	}
	return
}

func YesNo(value bool) *Value {
	enumValue := "YES"
	if !value {
		enumValue = "NO"
	}
	return &Value{Type: EnumType, EnumValue: &enumValue}
}

func (v *Value) YesNo() (value bool, err error) {
	yesno, err := v.Enum()
	if err != nil {
		return
	}
	switch yesno {
	case "YES":
		value = true
	case "NO":
		value = false
	default:
		err = fmt.Errorf("consuming %s value as YES/NO Enum: %w", yesno, ErrWrongType)
	}
	return
}

func (v *Value) StringOrEnum() (value string, err error) {
	if v.Type == StringType && v.StringValue != nil {
		value = *v.StringValue
	} else if v.Type == EnumType && v.EnumValue != nil {
		value = *v.EnumValue
	} else {
		err = fmt.Errorf("consuming %s value as String or Enum: %w", string(v.Type), ErrWrongType)
	}
	return
}

func Int(value int64) *Value {
	return &Value{Type: IntegerType, IntegerValue: &value}
}

func (v *Value) Int() (value int64, err error) {
	if v.Type == IntegerType && v.IntegerValue != nil {
		value = *v.IntegerValue
	} else {
		err = fmt.Errorf("consuming %s value as Int: %w", string(v.Type), ErrWrongType)
	}
	return
}

func (v *Value) Uint() (value uint64, err error) {
	if v.Type == IntegerType && v.IntegerValue != nil {
		if *v.IntegerValue < 0 {
			err = fmt.Errorf("consuming negative value as Uint: %w", ErrWrongType)
			return
		}
		value = uint64(*v.IntegerValue)
	} else {
		err = fmt.Errorf("consuming %s value as Uint: %w", string(v.Type), ErrWrongType)
	}
	return
}

func Float(value float64) *Value {
	return &Value{Type: FloatType, FloatValue: &value}
}

func (v *Value) Float() (value float64, err error) {
	if v.Type == FloatType && v.FloatValue != nil {
		value = *v.FloatValue
	} else {
		err = fmt.Errorf("consuming %s value as Float: %w", string(v.Type), ErrWrongType)
	}
	return
}

func (v *Value) Number() (value float64, err error) {
	if v.Type == IntegerType && v.IntegerValue != nil {
		value = float64(*v.IntegerValue)
	} else if v.Type == FloatType && v.FloatValue != nil {
		value = *v.FloatValue
	} else {
		err = fmt.Errorf("consuming %s value as Number: %w", string(v.Type), ErrWrongType)
	}
	return
}

func Bytes(value []byte) *Value {
	return &Value{Type: BytesType, BytesValue: value}
}

func (v *Value) Bytes() (value []byte, err error) {
	if v.Type == BytesType && v.BytesValue != nil {
		value = v.BytesValue
	} else {
		err = fmt.Errorf("consuming %s value as Bytes: %w", string(v.Type), ErrWrongType)
	}
	return
}

func ResolutionValue(value *Resolution) *Value {
	return &Value{Type: ResolutionType, ResolutionValue: value}
}

func (v *Value) Resolution() (value Resolution, err error) {
	if v.Type == ResolutionType && v.ResolutionValue != nil {
		value = *v.ResolutionValue
	} else {
		err = fmt.Errorf("consuming %s value as Resolution: %w", string(v.Type), ErrWrongType)
	}
	return
}

func (v *Value) StringPtr() (ptr *string, err error) {
	value, err := v.String()
	if err != nil {
		return
	}
	ptr = &value
	return
}

func (v *Value) EnumPtr() (ptr *string, err error) {
	value, err := v.Enum()
	if err != nil {
		return
	}
	ptr = &value
	return
}

func (v *Value) YesNoPtr() (ptr *bool, err error) {
	value, err := v.YesNo()
	if err != nil {
		return
	}
	ptr = &value
	return
}

func (v *Value) StringOrEnumPtr() (ptr *string, err error) {
	value, err := v.StringOrEnum()
	if err != nil {
		return
	}
	ptr = &value
	return
}

func (v *Value) IntPtr() (ptr *int64, err error) {
	value, err := v.Int()
	if err != nil {
		return
	}
	ptr = &value
	return
}

func (v *Value) UintPtr() (ptr *uint64, err error) {
	value, err := v.Uint()
	if err != nil {
		return
	}
	ptr = &value
	return
}

func (v *Value) FloatPtr() (ptr *float64, err error) {
	value, err := v.Float()
	if err != nil {
		return
	}
	ptr = &value
	return
}

func (v *Value) NumberPtr() (ptr *float64, err error) {
	value, err := v.Number()
	if err != nil {
		return
	}
	ptr = &value
	return
}

func (v *Value) ResolutionPtr() (ptr *Resolution, err error) {
	value, err := v.Resolution()
	if err != nil {
		return
	}
	ptr = &value
	return
}

func (v *Value) Format() string {
	switch v.Type {
	case StringType:
		return `"` + *v.StringValue + `"`
	case EnumType:
		return *v.EnumValue
	case IntegerType:
		return fmt.Sprintf("%d", *v.IntegerValue)
	case FloatType:
		return strconv.FormatFloat(*v.FloatValue, 'f', -1, 64)
	case BytesType:
		return "0x" + strings.ToUpper(hex.EncodeToString(v.BytesValue))
	case ResolutionType:
		return v.ResolutionValue.Format()
	}
	return "UNKNOWN_VALUE"
}
