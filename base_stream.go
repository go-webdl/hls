package hls

import (
	"fmt"
	"net/url"
)

type BaseStream struct {
	Tag              *Tag        // The EXT-X-STREAM-INF or EXT-X-I-FRAME-STREAM-INF tag
	TagLine          *Line       // The Line for the EXT-X-STREAM-INF or EXT-X-I-FRAME-STREAM-INF tag
	URI              *url.URL    // [REQUIRED] stream URI
	URILine          *Line       // [OPTIONAL] for EXT-X-STREAM-INF, the Line for the URL
	Bandwidth        uint64      // [REQUIRED] peak segment bit rate of the Variant Stream, in bits per second
	AverageBandwidth *uint64     // [OPTIONAL] average segment bit rate of the Variant Stream, in bits per second
	Score            *float64    // [OPTIONAL] abstract, relative measure of the playback quality-of-experience of the Variant Stream
	Codecs           *string     // [OPTIONAL] comma-separated list of formats, where each format specifies a media sample type that is present in one or more Renditions specified by the Variant Stream
	Resolution       *Resolution // [OPTIONAL] pixel resolution at which to display all the video in the Variant Stream
	HDCPLevel        *string     // [OPTIONAL] valid strings are TYPE-0, TYPE-1, and NONE
	AllowedCPC       *string     // [OPTIONAL] indicate that the playback of a Variant Stream containing encrypted Media Segments is to be restricted to devices that guarantee a certain level of content protection robustness
	VideoRange       *string     // [OPTIONAL] valid strings are SDR, HLG and PQ
	StableVariantID  *string     // [OPTIONAL] stable identifier for the URI within the Master Playlist
	Video            *string     // [OPTIONAL] indicates the set of video Renditions that SHOULD be used when playing the presentation
}

func (s *BaseStream) ParseAttributeList(attrs *AttributeList) (err error) {
	if attr := attrs.GetLast("BANDWIDTH"); attr == nil {
		err = fmt.Errorf("%s tag is missing BANDWIDTH attribute: %w", s.Tag.Name, ErrFormat)
		return
	} else {
		if s.Bandwidth, err = attr.Uint(); err != nil {
			err = fmt.Errorf("failed getting BANDWIDTH attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("AVERAGE-BANDWIDTH"); attr != nil {
		if s.AverageBandwidth, err = attr.UintPtr(); err != nil {
			err = fmt.Errorf("failed getting AVERAGE-BANDWIDTH attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("SCORE"); attr != nil {
		if s.Score, err = attr.FloatPtr(); err != nil {
			err = fmt.Errorf("failed getting SCORE attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("CODECS"); attr != nil {
		if s.Codecs, err = attr.StringPtr(); err != nil {
			err = fmt.Errorf("failed getting CODECS attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("RESOLUTION"); attr != nil {
		if s.Resolution, err = attr.ResolutionPtr(); err != nil {
			err = fmt.Errorf("failed getting RESOLUTION attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("HDCP-LEVEL"); attr != nil {
		if s.HDCPLevel, err = attr.EnumPtr(); err != nil {
			err = fmt.Errorf("failed getting HDCP-LEVEL attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("ALLOWED-CPC"); attr != nil {
		if s.AllowedCPC, err = attr.StringPtr(); err != nil {
			err = fmt.Errorf("failed getting ALLOWED-CPC attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("VIDEO-RANGE"); attr != nil {
		if s.VideoRange, err = attr.EnumPtr(); err != nil {
			err = fmt.Errorf("failed getting VIDEO-RANGE attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("STABLE-VARIANT-ID"); attr != nil {
		if s.StableVariantID, err = attr.StringPtr(); err != nil {
			err = fmt.Errorf("failed getting STABLE-VARIANT-ID attribute: %w", err)
			return
		}
	}
	return
}
