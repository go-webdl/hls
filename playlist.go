package hls

type Playlist struct {
	Lines   []*Line
	Version uint64 // [OPTIONAL][DEFAULT=1] indicates the compatibility version of the Playlist file, its associated media, and its server
}

type MediaPlaylist struct {
	*Playlist
	MediaSegments         []*MediaSegment
	MediaSequence         uint64 // [OPTIONAL][DEFAULT=0] indicates the Media Sequence Number of the first Media Segment that appears in a Playlist file
	DiscontinuitySequence uint64 // [OPTIONAL][DEFAULT=0] allows synchronization between different Renditions of the same Variant Stream or different Variant Streams
}

type MasterPlaylist struct {
	*Playlist
	VariantStreams  []*VariantStream
	IframeStreams   []*IframeStream
	RenditionGroups map[RenditionType]map[string][]*Rendition
}

func (playlsit *Playlist) Format() (str string) {
	for _, line := range playlsit.Lines {
		str += line.Format() + "\n"
	}
	return
}
