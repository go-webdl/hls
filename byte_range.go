package hls

import (
	"fmt"
	"strconv"
	"strings"
)

type ByteRange struct {
	Tag    *Tag
	Offset uint64 // [OPTIONAL][DEFAULT=inferred from previous byte range] indicating the start of the sub-range, as a byte offset from the beginning of the resource
	Length uint64 // [REQUIRED] the length of the sub-range in bytes
}

func (br *ByteRange) ParseString(value string, defaultOffset uint64) (err error) {
	atPars := strings.SplitN(value, "@", 2)
	if len(atPars) == 2 {
		br.Offset, err = strconv.ParseUint(atPars[1], 10, 64)
		if err != nil {
			err = fmt.Errorf("EXT-X-BYTERANGE offset has invalid integer format: %s: %w", atPars[1], ErrFormat)
			return
		}
	} else {
		br.Offset = defaultOffset
	}
	br.Length, err = strconv.ParseUint(atPars[0], 10, 64)
	if err != nil {
		err = fmt.Errorf("EXT-X-BYTERANGE length has invalid integer format: %s: %w", atPars[0], ErrFormat)
		return
	}
	return
}

func (br *ByteRange) ParseTag(tag *Tag, defaultOffset uint64) (err error) {
	if tag.Name != "EXT-X-BYTERANGE" {
		err = fmt.Errorf("parsing byte range using the wrong tag: %s: %w", tag.Name, ErrFormat)
		return
	}
	br.Tag = tag
	return br.ParseString(tag.Value, defaultOffset)
}

func (br ByteRange) End() uint64 {
	return br.Offset + br.Length
}
