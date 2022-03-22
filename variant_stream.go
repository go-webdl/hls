package hls

import (
	"fmt"
)

type VariantStream struct {
	BaseStream
	FrameRate          *float64 // [OPTIONAL] the maximum frame rate for all the video in the Variant Stream, rounded to three decimal places
	Audio              *string  // [OPTIONAL] indicates the set of audio Renditions that SHOULD be used when playing the presentation
	Video              *string  // [OPTIONAL] indicates the set of video Renditions that SHOULD be used when playing the presentation
	Subtitles          *string  // [OPTIONAL] indicates the set of subtitles Renditions that can be used when playing the presentation
	ClosedCaptions     *string  // [OPTIONAL] indicates the set of closed-captions Renditions that can be used when playing the presentation
	ClosedCaptionsNone bool     // indicates the CLOSED-CAPTIONS value is the NONE enum
}

func (s *VariantStream) ParseTag(tag *Tag) (err error) {
	if tag.Name != "EXT-X-STREAM-INF" {
		err = fmt.Errorf("parsing variant stream using the wrong tag: %s: %w", tag.Name, ErrFormat)
		return
	}
	s.Tag = tag
	if _, err = tag.ParseAttributeList(); err != nil {
		return
	}
	return s.ParseAttributeList(tag.AttributeList)
}

func (s *VariantStream) ParseAttributeList(attrs *AttributeList) (err error) {
	if err = s.BaseStream.ParseAttributeList(attrs); err != nil {
		return
	}
	if attr := attrs.GetLast("FRAME-RATE"); attr != nil {
		if s.FrameRate, err = attr.FloatPtr(); err != nil {
			return
		}
	}
	if attr := attrs.GetLast("AUDIO"); attr != nil {
		if s.Audio, err = attr.StringPtr(); err != nil {
			return
		}
	}
	if attr := attrs.GetLast("VIDEO"); attr != nil {
		if s.Video, err = attr.StringPtr(); err != nil {
			return
		}
	}
	if attr := attrs.GetLast("SUBTITLES"); attr != nil {
		if s.Subtitles, err = attr.StringPtr(); err != nil {
			return
		}
	}
	if attr := attrs.GetLast("CLOSED-CAPTIONS"); attr != nil {
		if s.ClosedCaptions, err = attr.StringOrEnumPtr(); err != nil {
			return
		}
		if attr.Type == EnumType && *s.ClosedCaptions == "NONE" {
			s.ClosedCaptionsNone = true
		}
	}
	return
}
