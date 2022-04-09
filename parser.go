package hls

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

type ParserHandler struct {
	HandleMediaSegment   func(segment *MediaSegment, playlist *MediaPlaylist) (next bool)
	HandleVariantStream  func(variantStream *VariantStream, playlist *MasterPlaylist) (next bool)
	HandleIframeStream   func(iframeStream *IframeStream, playlist *MasterPlaylist) (next bool)
	HandleMediaPlaylist  func(playlist *MediaPlaylist)
	HandleMasterPlaylist func(playlist *MasterPlaylist)
}

type LineType int

const (
	TagLineType LineType = iota
	URLLineType
	SpaceLineType
)

type Line struct {
	LineNum int
	Type    LineType
	Tag     *Tag
	URL     string
	Space   string
}

func (line Line) Format() string {
	switch line.Type {
	case TagLineType:
		return line.Tag.Format()
	case URLLineType:
		return line.URL
	case SpaceLineType:
		return line.Space
	}
	panic(fmt.Errorf("unknown line type %d", line.Type))
}

var IV128Pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 16)
	},
}

var ParserBufferPool = sync.Pool{
	New: func() interface{} {
		return bufio.NewReader(nil)
	},
}

func Parse(r io.Reader, baseURL *url.URL, handler *ParserHandler) (err error) {
	var (
		lineNum               int
		lineBytes             []byte
		lineStr               string
		lineTrimed            string
		isPrefix              bool
		buf                   *bufio.Reader
		key                   *Key
		isMaster              bool
		isMedia               bool
		stop                  bool
		byteRangeOffset       uint64
		mediaInitMap          *MediaInitMap
		mediaSequence         uint64
		discontinuitySequence uint64
		mediaSegmentBitrate   *uint64
	)

	buf = ParserBufferPool.Get().(*bufio.Reader)
	buf.Reset(r)

	playlist := &Playlist{Lines: make([]*Line, 0), Version: 1}
	mediaPlaylist := &MediaPlaylist{Playlist: playlist}
	mediaSegment := &MediaSegment{}
	masterPlaylist := &MasterPlaylist{Playlist: playlist}
	variantStream := &VariantStream{}
	ifrmaeStream := &IframeStream{}

	ensurePlaylist := func(cond bool, toSet *bool) error {
		if !cond {
			return fmt.Errorf("line %d: mixing media and master playlist tags: %w", lineNum, ErrFormat)
		}
		if toSet != nil {
			*toSet = true
		}
		return nil
	}

	finishMediaSegment := func() {
		mediaSegment.MediaSequence = mediaSequence
		mediaSegment.DiscontinuitySequence = discontinuitySequence
		mediaSegment.Key = key
		mediaSegment.MediaInitMap = mediaInitMap
		mediaSegment.Bitrate = mediaSegmentBitrate
		mediaSequence += 1
		mediaPlaylist.MediaSegments = append(mediaPlaylist.MediaSegments, mediaSegment)
		if handler.HandleMediaSegment != nil {
			stop = !handler.HandleMediaSegment(mediaSegment, mediaPlaylist)
		}
		mediaSegment = &MediaSegment{}
	}

	for !stop {
		lineNum++
		if lineBytes, isPrefix, err = buf.ReadLine(); err == io.EOF {
			err = nil
			break
		} else if isPrefix {
			err = fmt.Errorf("line %d: playlist line too long", lineNum)
			return
		} else if err != nil {
			err = fmt.Errorf("line %d: ReadLine failed: %w", lineNum, err)
			return
		}

		line := &Line{LineNum: lineNum}

		lineStr = string(lineBytes)
		lineTrimed = strings.TrimLeft(lineStr, " \t")

		if len(lineTrimed) == 0 {
			line.Type = SpaceLineType
			line.Space = lineStr
			playlist.Lines = append(playlist.Lines, line)
			continue
		}

		if lineTrimed[0] != '#' {
			var ref *url.URL
			if ref, err = url.Parse(lineTrimed); err != nil {
				err = fmt.Errorf("line %d: failed parsing line:\n%s\nas URL: %w", lineNum, lineTrimed, err)
				return
			}
			line.Type = URLLineType
			line.URL = lineTrimed
			playlist.Lines = append(playlist.Lines, line)
			resolvedURL := baseURL.ResolveReference(ref)
			if isMaster {
				if err = ensurePlaylist(!isMedia, nil); err != nil {
					return
				}
				variantStream.URI = resolvedURL
				variantStream.URILine = line
				masterPlaylist.VariantStreams = append(masterPlaylist.VariantStreams, variantStream)
				if handler.HandleVariantStream != nil {
					stop = !handler.HandleVariantStream(variantStream, masterPlaylist)
				}
				variantStream = &VariantStream{}
			} else {
				if err = ensurePlaylist(!isMaster, &isMedia); err != nil {
					return
				}
				mediaSegment.URI = resolvedURL
				mediaSegment.URILine = line
				finishMediaSegment()
			}
			continue
		}

		colonParts := strings.SplitN(lineTrimed, ":", 2)

		tag := &Tag{}
		tag.Name = colonParts[0][1:]
		if len(colonParts) == 2 {
			tag.Value = colonParts[1]
			tag.HasColon = true
		}
		line.Type = TagLineType
		line.Tag = tag
		playlist.Lines = append(playlist.Lines, line)

		switch tag.Name {
		case "EXT-X-VERSION":
			var e error
			if mediaPlaylist.Version, e = strconv.ParseUint(tag.Value, 10, 64); e != nil {
				err = fmt.Errorf("line %d: failed to parse EXT-X-VERSION value as integer: %s: %w", lineNum, e.Error(), ErrFormat)
				return
			}
		case "EXT-X-STREAM-INF":
			if err = ensurePlaylist(!isMedia, &isMaster); err != nil {
				return
			}
			variantStream.TagLine = line
			if err = variantStream.ParseTag(tag); err != nil {
				err = fmt.Errorf("line %d: %w", lineNum, err)
				return
			}
		case "EXT-X-I-FRAME-STREAM-INF":
			if err = ensurePlaylist(!isMedia, &isMaster); err != nil {
				return
			}
			ifrmaeStream.TagLine = line
			if err = ifrmaeStream.ParseTag(tag); err != nil {
				err = fmt.Errorf("line %d: %w", lineNum, err)
				return
			}
			masterPlaylist.IframeStreams = append(masterPlaylist.IframeStreams, ifrmaeStream)
			if handler.HandleIframeStream != nil {
				stop = !handler.HandleIframeStream(ifrmaeStream, masterPlaylist)
			}
			ifrmaeStream = &IframeStream{}
		case "EXT-X-MEDIA":
			if err = ensurePlaylist(!isMedia, &isMaster); err != nil {
				return
			}
			rendition := &Rendition{}
			if err = rendition.ParseTag(tag); err != nil {
				err = fmt.Errorf("line %d: %w", lineNum, err)
				return
			}
			if masterPlaylist.RenditionGroups == nil {
				masterPlaylist.RenditionGroups = make(map[RenditionType]map[string][]*Rendition)
			}
			if masterPlaylist.RenditionGroups[rendition.Type] == nil {
				masterPlaylist.RenditionGroups[rendition.Type] = make(map[string][]*Rendition)
			}
			masterPlaylist.RenditionGroups[rendition.Type][rendition.GroupID] = append(masterPlaylist.RenditionGroups[rendition.Type][rendition.GroupID], rendition)

		case "EXT-X-MEDIA-SEQUENCE":
			var e error
			if err = ensurePlaylist(!isMaster, &isMedia); err != nil {
				return
			}
			if mediaSequence, e = strconv.ParseUint(tag.Value, 10, 64); e != nil {
				err = fmt.Errorf("line %d: failed to parse EXT-X-MEDIA-SEQUENCE value as integer: %s: %w", lineNum, e.Error(), ErrFormat)
				return
			}
			mediaPlaylist.MediaSequence = mediaSequence
		case "EXT-X-DISCONTINUITY-SEQUENCE":
			var e error
			if err = ensurePlaylist(!isMaster, &isMedia); err != nil {
				return
			}
			if discontinuitySequence, e = strconv.ParseUint(tag.Value, 10, 64); e != nil {
				err = fmt.Errorf("line %d: failed to parse EXT-X-DISCONTINUITY-SEQUENCE value as integer: %s: %w", lineNum, e.Error(), ErrFormat)
				return
			}
			mediaPlaylist.DiscontinuitySequence = discontinuitySequence
		case "EXTINF":
			if err = ensurePlaylist(!isMaster, &isMedia); err != nil {
				return
			}
			if err = mediaSegment.ParseTag(tag); err != nil {
				err = fmt.Errorf("line %d: %w", lineNum, err)
				return
			}
		case "EXT-X-BYTERANGE":
			if err = ensurePlaylist(!isMaster, &isMedia); err != nil {
				return
			}
			if err = mediaSegment.ParseByteRangeTag(tag, byteRangeOffset); err != nil {
				err = fmt.Errorf("line %d: %w", lineNum, err)
				return
			}
			byteRangeOffset = mediaSegment.ByteRange.End()
		case "EXT-X-DISCONTINUITY":
			if err = ensurePlaylist(!isMaster, &isMedia); err != nil {
				return
			}
			mediaSegment.IsDiscontinuity = true
			discontinuitySequence += 1
		case "EXT-X-GAP":
			if err = ensurePlaylist(!isMaster, &isMedia); err != nil {
				return
			}
			mediaSegment.IsGap = true
		case "EXT-X-MAP":
			mediaInitMap = &MediaInitMap{Key: key}
			if err = mediaInitMap.ParseTag(tag); err != nil {
				err = fmt.Errorf("line %d: %w", lineNum, err)
				return
			}
		case "EXT-X-KEY":
			key = &Key{}
			if err = key.ParseTag(tag); err != nil {
				err = fmt.Errorf("line %d: %w", lineNum, err)
				return
			}
		}
	}

	if isMaster {
		if handler.HandleMasterPlaylist != nil {
			handler.HandleMasterPlaylist(masterPlaylist)
		}
	} else if isMedia {
		if handler.HandleMediaPlaylist != nil {
			handler.HandleMediaPlaylist(mediaPlaylist)
		}
	} else {
		err = fmt.Errorf("ambiguous playlist: %w", ErrFormat)
		return
	}

	return
}
