package hls

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAttributeList(t *testing.T) {
	lineStr := `TYPE=AUDIO,GROUP-ID="aac",BITRATE=600000,SCORE=-3.14159,KEY=0xDEADBEEF00112233,RESOLUTION=1920x1080,NAME="English",DEFAULT=YES,AUTOSELECT=YES,LANGUAGE="en",URI="main/english-audio.m3u8",EXTRA=0`
	attrs, err := ParseAttributeList(lineStr)
	if err != nil {
		t.Fatal(err)
	}
	for _, attr := range attrs.List() {
		switch attr.Name {
		case "TYPE":
			assert.Equal(t, EnumType, attr.Type)
			assert.Equal(t, "AUDIO", *attr.EnumValue)
		case "GROUP-ID":
			assert.Equal(t, StringType, attr.Type)
			assert.Equal(t, "aac", *attr.StringValue)
		case "BITRATE":
			assert.Equal(t, IntegerType, attr.Type)
			assert.EqualValues(t, 600000, *attr.IntegerValue)
		case "SCORE":
			assert.Equal(t, FloatType, attr.Type)
			assert.InEpsilon(t, -3.14159, *attr.FloatValue, 0.00001)
		case "KEY":
			assert.Equal(t, BytesType, attr.Type)
			assert.Equal(t, []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x00, 0x11, 0x22, 0x33}, attr.BytesValue)
		case "RESOLUTION":
			assert.Equal(t, ResolutionType, attr.Type)
			assert.EqualValues(t, &Resolution{1920, 1080}, attr.ResolutionValue)
		case "NAME":
			assert.Equal(t, StringType, attr.Type)
			assert.Equal(t, "English", *attr.StringValue)
		case "DEFAULT":
			assert.Equal(t, EnumType, attr.Type)
			assert.Equal(t, "YES", *attr.EnumValue)
		case "AUTOSELECT":
			assert.Equal(t, EnumType, attr.Type)
			assert.Equal(t, "YES", *attr.EnumValue)
		case "LANGUAGE":
			assert.Equal(t, StringType, attr.Type)
			assert.Equal(t, "en", *attr.StringValue)
		case "URI":
			assert.Equal(t, StringType, attr.Type)
			assert.Equal(t, "main/english-audio.m3u8", *attr.StringValue)
		case "EXTRA":
			assert.Equal(t, IntegerType, attr.Type)
			assert.EqualValues(t, 0, *attr.IntegerValue)
		}
	}
	formatted := attrs.Format()
	if formatted != lineStr {
		t.Logf("formatted attributes: %s\n", formatted)
		t.Error("formatted attributes doesn't match original line")
	}
}
