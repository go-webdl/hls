package hls

import (
	"fmt"
	"net/url"
)

type IframeStream struct {
	BaseStream
}

func (s *IframeStream) ParseTag(tag *Tag) (err error) {
	if tag.Name != "EXT-X-I-FRAME-STREAM-INF" {
		err = fmt.Errorf("parsing iframe stream using the wrong tag: %s: %w", tag.Name, ErrFormat)
		return
	}
	s.Tag = tag
	if _, err = tag.ParseAttributeList(); err != nil {
		err = fmt.Errorf("failed parsing iframe stream attribute list: %w", err)
		return
	}
	return s.ParseAttributeList(tag.AttributeList)
}

func (s *IframeStream) ParseAttributeList(attrs *AttributeList) (err error) {
	if err = s.BaseStream.ParseAttributeList(attrs); err != nil {
		return
	}
	if attr := attrs.GetLast("URI"); attr == nil {
		err = fmt.Errorf("%s tag is missing URI attribute: %w", s.Tag.Name, ErrFormat)
		return
	} else {
		var value string
		if value, err = attr.String(); err != nil {
			err = fmt.Errorf("failed getting URI attribute: %w", err)
			return
		}
		if s.URI, err = url.Parse(value); err != nil {
			err = fmt.Errorf("failed parsing URI attribute value as URL: %w", err)
			return
		}
	}
	return
}
