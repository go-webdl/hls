package hls

import (
	"fmt"
)

type Tag struct {
	Name     string // without the leading #
	Value    string
	HasColon bool // false implies empty Value
	*AttributeList
}

func (tag *Tag) ParseAttributeList() (attrs *AttributeList, err error) {
	if tag.AttributeList != nil {
		return tag.AttributeList, nil
	}
	if tag.AttributeList, err = ParseAttributeList(tag.Value); err != nil {
		err = fmt.Errorf("error in parsing tag %s attribute list: %w", tag.Name, err)
	}
	return tag.AttributeList, err
}

func (tag *Tag) UpdateValue() {
	tag.Value = tag.AttributeList.Format()
}

func (tag Tag) Format() string {
	if !tag.HasColon {
		return "#" + tag.Name
	}
	return "#" + tag.Name + ":" + tag.Value
}
