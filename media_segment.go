package hls

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type MediaSegment struct {
	Tag             *Tag
	URI             *url.URL      // [REQUIRED] Media Segment URI
	URILine         *Line         // [REQUIRED] The Line for the URI
	Duration        time.Duration // [REQUIRED] specifies the duration of the Media Segment
	Title           string        // [OPTIONAL][DEFAULT=""] human-readable informative title of the Media Segment
	ByteRange       *ByteRange    // [OPTIONAL] indicates that a Media Segment is a sub-range of the resource identified by its URI
	IsDiscontinuity bool          // [OPTIONAL] indicates a discontinuity between the Media Segment that follows it and the one that preceded it
	IsGap           bool          // [OPTIONAL] indicates that the segment URL to which it applies does not contain media data and SHOULD NOT be loaded by clients

	// the following are computed/inherited values
	MediaSequence         uint64        // [OPTIONAL][DEFAULT=start at 0 and increment]
	DiscontinuitySequence uint64        // [OPTIONAL][DEFAULT=start at 0 and increment]
	Key                   *Key          // [OPTIONAL]
	MediaInitMap          *MediaInitMap // [OPTIONAL]
	Bitrate               *uint64       // [OPTIONAL]
}

func (s *MediaSegment) ParseTag(tag *Tag) (err error) {
	if tag.Name != "EXTINF" {
		err = fmt.Errorf("parsing media segment using the wrong tag: %s: %w", tag.Name, ErrFormat)
		return
	}
	s.Tag = tag
	commaParts := strings.SplitN(tag.Value, ",", 2)
	if len(commaParts) == 2 {
		s.Title = strings.TrimSpace(commaParts[1])
	}
	duration, err := strconv.ParseFloat(commaParts[0], 64)
	if err != nil {
		err = fmt.Errorf("EXTINF duration has invalid float or integer format: %s: %w", commaParts[0], ErrFormat)
		return
	}
	s.Duration = time.Duration(duration * float64(time.Second))
	return
}

func (s *MediaSegment) ParseByteRangeTag(tag *Tag, defaultOffset uint64) (err error) {
	br := &ByteRange{}
	if err = br.ParseTag(tag, defaultOffset); err != nil {
		return
	}
	s.ByteRange = br
	return
}
