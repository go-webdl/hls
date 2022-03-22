package hls

import (
	"strconv"
	"strings"
)

type RenditionChannels struct {
	AudioChannelsCount           *uint64
	AudioObjectCodingIdentifiers []string
}

func (c *RenditionChannels) ParseString(renditionType RenditionType, value string) (err error) {
	slashParts := strings.Split(value, "/")
	if renditionType == Audio {
		if len(slashParts) >= 1 {
			var channelsCount uint64
			if channelsCount, err = strconv.ParseUint(slashParts[0], 10, 64); err != nil {
				return
			}
			c.AudioChannelsCount = &channelsCount
		}
		if len(slashParts) >= 2 {
			c.AudioObjectCodingIdentifiers = strings.Split(slashParts[1], ",")
		}
	}
	return
}
