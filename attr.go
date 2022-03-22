package hls

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

type Attribute struct {
	Name string
	Value
}

func (attr *Attribute) Format() string {
	return fmt.Sprintf("%s=%s", attr.Name, attr.Value.Format())
}

type AttributeList struct {
	attrs   []*Attribute
	mapping map[string][]*Attribute
}

func (attrs AttributeList) List() []*Attribute {
	return attrs.attrs
}

func (attrs *AttributeList) Append(attr *Attribute) {
	attrs.attrs = append(attrs.attrs, attr)
	if attrs.mapping == nil {
		attrs.mapping = make(map[string][]*Attribute)
	}
	attrs.mapping[attr.Name] = append(attrs.mapping[attr.Name], attr)
}

func (attrs AttributeList) Get(name string) []*Attribute {
	return attrs.mapping[name]
}

func (attrs AttributeList) GetFirst(name string) *Attribute {
	if list := attrs.mapping[name]; len(list) > 0 {
		return list[0]
	}
	return nil
}

func (attrs AttributeList) GetLast(name string) *Attribute {
	if list := attrs.mapping[name]; len(list) > 0 {
		return list[len(list)-1]
	}
	return nil
}

func (attrs *AttributeList) Remove(name string) {
	newAttrs := []*Attribute{}
	for _, attr := range attrs.attrs {
		if attr.Name != name {
			newAttrs = append(newAttrs, attr)
		}
	}
	attrs.attrs = newAttrs
	delete(attrs.mapping, name)
}

func (attrs *AttributeList) Set(name string, value *Value) {
	newAttrs := []*Attribute{}
	var found *Attribute
	for _, attr := range attrs.attrs {
		if attr.Name != name {
			newAttrs = append(newAttrs, attr)
		} else if found == nil {
			found = attr
			attr.Value = *value
			newAttrs = append(newAttrs, attr)
		}
	}
	if found == nil {
		newAttrs = append(newAttrs, &Attribute{name, *value})
	}
	attrs.attrs = newAttrs
	attrs.mapping[name] = []*Attribute{found}
}

func (attrs AttributeList) Format() string {
	var attrStrings []string
	for _, attr := range attrs.attrs {
		attrStrings = append(attrStrings, attr.Format())
	}
	return strings.Join(attrStrings, ",")
}

func isNumericChar(c byte) bool {
	return c >= '0' && c <= '9'
}

func isSpaceChar(c byte) bool {
	return c == ' ' || c == '\t'
}

func isAttributeNameChar(c byte) bool {
	return (c >= 'A' && c <= 'Z') || isNumericChar(c) || (c == '-') || (c == '_')
}

func isBytesChar(c byte) bool {
	return (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') || isNumericChar(c)
}

const (
	attrStateStart = iota
	attrStateName
	attrStateNameEnd
	attrStateString
	attrStateEnum
	attrStateInteger
	attrStateFloat
	attrStateBytes
	attrStateResolution
	attrStateValueEnd
)

func ParseAttributeList(listStr string) (attrs *AttributeList, err error) {
	var (
		pos    int
		start  int
		signed bool
	)
	attr := &Attribute{}
	state := attrStateStart
	attrs = &AttributeList{}

	appendAttr := func(t Type) {
		attr.Type = t
		attrs.Append(attr)
		attr = &Attribute{}
		signed = false
	}

	finishEnum := func(c byte) (err error) {
		value := listStr[start:pos]
		attr.EnumValue = &value
		appendAttr(EnumType)
		if c == ',' {
			state = attrStateStart
		} else {
			state = attrStateValueEnd
		}
		return
	}

	finishInteger := func(c byte) (err error) {
		if value, e := strconv.ParseInt(listStr[start:pos], 10, 64); e != nil {
			err = fmt.Errorf("parsing Integer, strconv error: %s %w", e.Error(), ErrFormat)
			return
		} else {
			attr.IntegerValue = &value
		}
		appendAttr(IntegerType)
		if c == ',' {
			state = attrStateStart
		} else {
			state = attrStateValueEnd
		}
		return
	}

	finishFloat := func(c byte) (err error) {
		if value, e := strconv.ParseFloat(listStr[start:pos], 64); e != nil {
			err = fmt.Errorf("parsing Float, strconv error: %s %w", e.Error(), ErrFormat)
			return
		} else {
			attr.FloatValue = &value
		}
		appendAttr(FloatType)
		if c == ',' {
			state = attrStateStart
		} else {
			state = attrStateValueEnd
		}
		return
	}

	finishBytes := func(c byte) (err error) {
		if (pos-start)%2 != 0 {
			err = fmt.Errorf("parsing Bytes, value length must be multiples of 2: %w", ErrFormat)
		} else if pos-start <= 2 {
			attr.BytesValue = make([]byte, 0)
		} else {
			attr.BytesValue, err = hex.DecodeString(listStr[start+2 : pos])
		}
		if err != nil {
			return
		}
		appendAttr(BytesType)
		if c == ',' {
			state = attrStateStart
		} else {
			state = attrStateValueEnd
		}
		return
	}

	finishResolution := func(c byte) (err error) {
		if start == pos {
			err = fmt.Errorf("paring Resolution, missing Height: %w", ErrFormat)
			return
		}
		if height, e := strconv.ParseUint(listStr[start:pos], 10, 64); e != nil {
			err = fmt.Errorf("parsing Resolution.Height, strconv error: %s %w", e.Error(), ErrFormat)
			return
		} else {
			attr.ResolutionValue.Height = int(height)
		}
		appendAttr(ResolutionType)
		if c == ',' {
			state = attrStateStart
		} else {
			state = attrStateValueEnd
		}
		return
	}

	var c byte
	for pos = 0; pos < len(listStr); pos++ {
		c = listStr[pos]
		switch state {
		case attrStateStart:
			if isSpaceChar(c) || c == ',' {
			} else if isAttributeNameChar(c) {
				start = pos
				state = attrStateName
			} else {
				err = fmt.Errorf("before name start, expecting space, comma or start of attribute name, but got '%c': %w", c, ErrFormat)
			}
		case attrStateName:
			if c == '=' || isSpaceChar(c) {
				attr.Name = listStr[start:pos]
				state = attrStateNameEnd
			} else if isAttributeNameChar(c) {
			} else {
				err = fmt.Errorf("inside name, expecting space, '=', or attribute name character, but got '%c': %w", c, ErrFormat)
			}
		case attrStateNameEnd:
			if isSpaceChar(c) {
			} else if c == '"' {
				start = pos + 1
				state = attrStateString
			} else if c == '-' {
				start = pos
				signed = true
				state = attrStateInteger
			} else if c == '.' {
				start = pos
				state = attrStateFloat
			} else if c == '0' {
				start = pos
				state = attrStateBytes
			} else if c >= '1' && c <= '9' {
				start = pos
				state = attrStateInteger
			} else if c != ',' {
				start = pos
				state = attrStateEnum
			} else {
				err = fmt.Errorf("at value start, got invalid comma character: %w", ErrFormat)
			}
		case attrStateString:
			if c == '"' {
				value := listStr[start:pos]
				attr.StringValue = &value
				appendAttr(StringType)
				state = attrStateValueEnd
			}
		case attrStateEnum:
			if c == ',' || isSpaceChar(c) {
				err = finishEnum(c)
			} else if c == '"' {
				err = fmt.Errorf("inside Enum, expecting non-double-quote characters, but got '%c': %w", c, ErrFormat)
			}
		case attrStateInteger:
			if isNumericChar(c) {
			} else if c == '.' {
				state = attrStateFloat
			} else if c == 'x' || c == 'X' {
				if signed {
					err = fmt.Errorf("parsing Resolution, Width cannot be signed: %w", ErrFormat)
					return
				}
				value := Resolution{}
				if width, e := strconv.ParseUint(listStr[start:pos], 10, 64); e != nil {
					err = fmt.Errorf("parsing Resolution.Width, strconv error: %s %w", e.Error(), ErrFormat)
					return
				} else {
					value.Width = int(width)
				}
				attr.ResolutionValue = &value
				start = pos + 1
				state = attrStateResolution
			} else if c == ',' || isSpaceChar(c) {
				err = finishInteger(c)
			}
		case attrStateFloat:
			if isNumericChar(c) {
			} else if c == ',' || isSpaceChar(c) {
				err = finishFloat(c)
			}
		case attrStateBytes:
			if pos == start+1 {
				if c == '.' {
					state = attrStateFloat
				} else if isNumericChar(c) {
					state = attrStateInteger
				} else if c == ',' || isSpaceChar(c) {
					err = finishInteger(c)
				} else if c != 'x' && c != 'X' {
					state = attrStateEnum
					pos--
				}
			} else if pos > start+1 && (pos-start)%2 == 0 {
				if isBytesChar(c) {
				} else if c == ',' || isSpaceChar(c) {
					err = finishBytes(c)
				} else {
					state = attrStateEnum
					pos--
				}
			} else {
				if isBytesChar(c) {
				} else if c == ',' || isSpaceChar(c) {
					err = finishEnum(c)
				} else {
					state = attrStateEnum
					pos--
				}
			}
		case attrStateResolution:
			if isNumericChar(c) {
			} else if c == ',' || isSpaceChar(c) {
				err = finishResolution(c)
			}
		case attrStateValueEnd:
			if c == ',' {
				state = attrStateStart
			} else if isSpaceChar(c) {
			} else {
				err = fmt.Errorf("expecting space or comma, but got '%c': %w", c, ErrFormat)
			}
		}
		if err != nil {
			return
		}
	}

	c = ' '
	switch state {
	case attrStateEnum:
		err = finishEnum(c)
	case attrStateInteger:
		err = finishInteger(c)
	case attrStateFloat:
		err = finishFloat(c)
	case attrStateBytes:
		if pos == start+1 {
			err = finishInteger(c)
		} else if pos > start+1 && (pos-start)%2 == 0 {
			err = finishBytes(c)
		} else {
			err = finishEnum(c)
		}
	case attrStateResolution:
		err = finishResolution(c)
	case attrStateStart, attrStateValueEnd:
	default:
		err = fmt.Errorf("invalid attribute list format:\n%s\nError: %w", listStr, ErrFormat)
	}

	return
}
