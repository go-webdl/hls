package hls

import (
	"fmt"
	"net/url"
	"strings"
)

type Rendition struct {
	Tag               *Tag
	URI               *url.URL           // [OPTIONAL] identifies the Media Playlist file, only nil for CLOSED-CAPTIONS
	Type              RenditionType      // [REQUIRED] valid strings are AUDIO, VIDEO, SUBTITLES, and CLOSED-CAPTIONS
	GroupID           string             // [REQUIRED] specifies the group to which the Rendition belongs
	Name              string             // [REQUIRED] human-readable description of the Rendition
	Language          *string            // [OPTIONAL] containing one of the standard Tags for Identifying Languages [RFC5646], which identifies the primary language used in the Rendition
	AssocLanguage     *string            // [OPTIONAL] identifies a language that is associated with the Rendition
	StableRenditionID *string            // [OPTIONAL] stable identifier for the URI within the Master Playlist
	Default           bool               // [OPTIONAL][DEFAULT=false] If the value is YES, then the client SHOULD play this Rendition of the content in the absence of information from the user indicating a different choice
	Autoselect        bool               // [OPTIONAL][DEFAULT=false] If the value is YES, then the client MAY choose to play this Rendition in the absence of explicit user preference because it matches the current playback environment, such as chosen system language
	Forced            bool               // [OPTIONAL][DEFAULT=false] A value of YES indicates that the Rendition contains content that is considered essential to play
	InstreamID        *string            // [OPTIONAL] specifies a Rendition within the segments in the Media Playlist
	Characteristics   []string           // [OPTIONAL] one or more Media Characteristic Tags (MCTs)
	Channels          *RenditionChannels // [OPTIONAL] specifies an ordered, slash-separated ("/") list of parameters
}

type RenditionType string

const (
	Audio          RenditionType = "AUDIO"
	Video          RenditionType = "VIDEO"
	Subtitles      RenditionType = "SUBTITLES"
	ClosedCaptions RenditionType = "CLOSED-CAPTIONS"
)

func (r *Rendition) ParseTag(tag *Tag) (err error) {
	if tag.Name != "EXT-X-MEDIA" {
		err = fmt.Errorf("parsing rendition using the wrong tag: %s: %w", tag.Name, ErrFormat)
		return
	}
	r.Tag = tag
	if _, err = tag.ParseAttributeList(); err != nil {
		err = fmt.Errorf("failed parsing rendition attribute list: %w", err)
		return
	}
	return r.ParseAttributeList(tag.AttributeList)
}

func (r *Rendition) ParseAttributeList(attrs *AttributeList) (err error) {
	if attr := attrs.GetLast("TYPE"); attr == nil {
		err = fmt.Errorf("%s tag is missing TYPE attribute: %w", r.Tag.Name, ErrFormat)
		return
	} else {
		var value string
		if value, err = attr.Enum(); err != nil {
			err = fmt.Errorf("failed getting TYPE attribute: %w", err)
			return
		}
		renditionType := RenditionType(value)
		switch renditionType {
		case Audio, Video, Subtitles, ClosedCaptions:
			r.Type = renditionType
		default:
			err = fmt.Errorf("%s tag has invalid TYPE enum value: %s: %w", r.Tag.Name, value, ErrFormat)
			return
		}
	}
	if attr := attrs.GetLast("URI"); attr == nil {
		if r.Type != ClosedCaptions {
			err = fmt.Errorf("%s tag is missing URI attribute: %w", r.Tag.Name, ErrFormat)
			return
		}
	} else {
		var value string
		if value, err = attr.String(); err != nil {
			err = fmt.Errorf("failed getting URI attribute: %w", err)
			return
		}
		if r.URI, err = url.Parse(value); err != nil {
			err = fmt.Errorf("failed parsing URI attribute value as URL: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("GROUP-ID"); attr == nil {
		err = fmt.Errorf("%s tag is missing GROUP-ID attribute: %w", r.Tag.Name, ErrFormat)
		return
	} else {
		if r.GroupID, err = attr.String(); err != nil {
			err = fmt.Errorf("failed getting GROUP-ID attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("NAME"); attr == nil {
		err = fmt.Errorf("%s tag is missing NAME attribute: %w", r.Tag.Name, ErrFormat)
		return
	} else {
		if r.Name, err = attr.String(); err != nil {
			err = fmt.Errorf("failed getting NAME attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("LANGUAGE"); attr != nil {
		if r.Language, err = attr.StringPtr(); err != nil {
			err = fmt.Errorf("failed getting LANGUAGE attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("ASSOC-LANGUAGE"); attr != nil {
		if r.AssocLanguage, err = attr.StringPtr(); err != nil {
			err = fmt.Errorf("failed getting ASSOC-LANGUAGE attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("STABLE-RENDITION-ID"); attr != nil {
		if r.AssocLanguage, err = attr.StringPtr(); err != nil {
			err = fmt.Errorf("failed getting STABLE-RENDITION-ID attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("DEFAULT"); attr != nil {
		if r.Default, err = attr.YesNo(); err != nil {
			err = fmt.Errorf("failed getting DEFAULT attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("AUTOSELECT"); attr != nil {
		if r.Autoselect, err = attr.YesNo(); err != nil {
			err = fmt.Errorf("failed getting AUTOSELECT attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("FORCED"); attr != nil {
		if r.Forced, err = attr.YesNo(); err != nil {
			err = fmt.Errorf("failed getting FORCED attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("INSTREAM-ID"); attr != nil {
		if r.InstreamID, err = attr.StringPtr(); err != nil {
			err = fmt.Errorf("failed getting INSTREAM-ID attribute: %w", err)
			return
		}
	}
	if attr := attrs.GetLast("CHARACTERISTICS"); attr != nil {
		var value string
		if value, err = attr.String(); err != nil {
			err = fmt.Errorf("failed getting CHARACTERISTICS attribute: %w", err)
			return
		}
		r.Characteristics = strings.Split(value, ",")
	}
	if attr := attrs.GetLast("CHANNELS"); attr != nil {
		var value string
		if value, err = attr.String(); err != nil {
			err = fmt.Errorf("failed getting CHANNELS attribute: %w", err)
			return
		}
		r.Channels = &RenditionChannels{}
		if err = r.Channels.ParseString(r.Type, value); err != nil {
			err = fmt.Errorf("failed parsing CHANNELS attribute value as RenditionChannels: %w", err)
			return
		}
	}
	return
}
