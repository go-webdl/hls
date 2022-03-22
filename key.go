package hls

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Key struct {
	Tag               *Tag
	Method            KeyMethod // [REQUIRED] specifies the encryption method
	URI               *url.URL  // [OPTIONAL] specifies how to obtain the key
	IV                []byte    // [OPTIONAL] specifies a 128-bit unsigned integer Initialization Vector to be used with the key
	KeyFormat         *string   // [OPTIONAL] specifies how the key is represented in the resource identified by the URI
	KeyFormatVersions []uint64  // [OPTIONAL] indicate which version(s) this instance complies with
}

type KeyMethod string

const (
	KeyMethodNone      KeyMethod = "NONE"
	KeyMethodAES128    KeyMethod = "AES-128"
	KeyMethodSampleAES KeyMethod = "SAMPLE-AES"
)

func (k *Key) ParseTag(tag *Tag) (err error) {
	if tag.Name != "EXT-X-KEY" {
		err = fmt.Errorf("parsing key object using the wrong tag: %s: %w", tag.Name, ErrFormat)
		return
	}
	k.Tag = tag
	if _, err = tag.ParseAttributeList(); err != nil {
		err = fmt.Errorf("failed parsing key object attribute list: %w", err)
		return
	}
	return k.ParseAttributeList(tag.AttributeList)
}

func (k *Key) ParseAttributeList(attrs *AttributeList) (err error) {
	if attr := attrs.GetLast("METHOD"); attr == nil {
		err = fmt.Errorf("%s tag is missing METHOD attribute: %w", k.Tag.Name, ErrFormat)
		return
	} else {
		var value string
		if value, err = attr.Enum(); err != nil {
			err = fmt.Errorf("failed getting METHOD attribute: %w", err)
			return
		}
		method := KeyMethod(value)
		switch method {
		case KeyMethodNone, KeyMethodAES128, KeyMethodSampleAES:
			k.Method = method
		default:
			err = fmt.Errorf("%s tag has invalid METHOD enum value: %s: %w", k.Tag.Name, value, ErrFormat)
			return
		}
	}
	if attr := attrs.GetLast("URI"); attr != nil {
		var value string
		if value, err = attr.String(); err != nil {
			err = fmt.Errorf("failed getting URI attribute: %w", err)
			return
		}
		if k.URI, err = url.Parse(value); err != nil {
			err = fmt.Errorf("failed parsing URI attribute value as URL: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("IV"); attr != nil {
		if k.IV, err = attr.Bytes(); err != nil {
			err = fmt.Errorf("failed getting IV attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("KEYFORMAT"); attr != nil {
		if k.KeyFormat, err = attr.StringPtr(); err != nil {
			err = fmt.Errorf("failed getting KEYFORMAT attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("KEYFORMATVERSIONS"); attr != nil {
		var value string
		if value, err = attr.String(); err != nil {
			err = fmt.Errorf("failed getting KEYFORMATVERSIONS attribute: %w", err)
			return
		}
		// The value is a quoted-string containing one or more positive
		// integers separated by the "/" character (for example, "1", "1/2",
		// or "1/2/5").
		slashParts := strings.Split(value, "/")
		for _, part := range slashParts {
			var version uint64
			if version, err = strconv.ParseUint(part, 10, 64); err != nil {
				err = fmt.Errorf("%s tag failed to parse version part as integer: %s: %w", k.Tag.Name, part, ErrFormat)
				return
			}
			k.KeyFormatVersions = append(k.KeyFormatVersions, version)
		}
	}
	if len(k.KeyFormatVersions) == 0 {
		// if it is not present, its value is considered to be "1"
		k.KeyFormatVersions = append(k.KeyFormatVersions, 1)
	}
	return
}
