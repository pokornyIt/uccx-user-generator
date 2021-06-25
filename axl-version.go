package main

import (
	"encoding/xml"
	log "github.com/sirupsen/logrus"
	"strings"
)

type version struct {
	XMLName xml.Name `xml:"return"`
	Version string   `xml:"componentVersion>version"`
}

func VersionData(data string) (*version, error) {
	var ver version
	d := []byte(data)
	err := xml.Unmarshal(d, &ver)
	if err != nil {
		log.Errorf("problem Unmarshal XML source data to Version structure - [%s]", err)
		ver = version{
			Version: "",
		}
		return &ver, err
	}
	return &ver, nil
}

func (v *version) GetDbVersion() string {
	if v.Version == "" {
		return ""
	}
	ver := v.Version[:strings.Index(v.Version, ".")] + ".0"
	return ver
}
