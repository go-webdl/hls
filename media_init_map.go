package hls

import (
	"fmt"
	"net/url"
)

type MediaInitMap struct {
	Tag       *Tag
	URI       *url.URL
	ByteRange *ByteRange

	// the following are computed values
	Key *Key
}

func (m *MediaInitMap) ParseTag(tag *Tag) (err error) {
	if tag.Name != "EXT-X-MAP" {
		err = fmt.Errorf("parsing media init map using the wrong tag: %s: %w", tag.Name, ErrFormat)
		return
	}
	m.Tag = tag
	if _, err = tag.ParseAttributeList(); err != nil {
		err = fmt.Errorf("failed parsing media init map attribute list: %w", err)
		return
	}
	return m.ParseAttributeList(tag.AttributeList)
}

func (m *MediaInitMap) ParseAttributeList(attrs *AttributeList) (err error) {
	if attr := attrs.GetLast("URI"); attr == nil {
		err = fmt.Errorf("%s tag is missing URI attribute: %w", m.Tag.Name, ErrFormat)
		return
	} else {
		var value string
		if value, err = attr.String(); err != nil {
			err = fmt.Errorf("failed getting URI attribute: %w", err)
			return
		}
		if m.URI, err = url.Parse(value); err != nil {
			err = fmt.Errorf("failed parsing URI attribute value as URL: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("BYTERANGE"); attr != nil {
		var value string
		if value, err = attr.String(); err != nil {
			err = fmt.Errorf("failed getting BYTERANGE attribute: %w", err)
			return
		}
		br := &ByteRange{}
		if err = br.ParseString(value, 0); err != nil {
			err = fmt.Errorf("failed parsing BYTERANGE attribute value as ByteRange: %w", err)
			return
		}
		m.ByteRange = br
	}
	return
}
