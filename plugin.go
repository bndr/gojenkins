package main

import (
	"strconv"
)

type Plugins struct {
	Jenkins *Jenkins
	Raw     *pluginResponse
	Base    string
	Depth   int
}

type pluginResponse struct {
	Plugins []Plugin `json:"plugins"`
}

type Plugin struct {
	Active              bool        `json:"active"`
	BackupVersion       interface{} `json:"backupVersion"`
	Bundled             bool        `json:"bundled"`
	Deleted             bool        `json:"deleted"`
	Dependencies        []struct{}  `json:"dependencies"`
	Downgradable        bool        `json:"downgradable"`
	Enabled             bool        `json:"enabled"`
	HasUpdate           bool        `json:"hasUpdate"`
	LongName            string      `json:"longName"`
	Pinned              bool        `json:"pinned"`
	ShortName           string      `json:"shortName"`
	SupportsDynamicLoad string      `json:"supportsDynamicLoad"`
	URL                 string      `json:"url"`
	Version             string      `json:"version"`
}

func (p *Plugins) Contains(name string) *Plugin {
	for _, p := range p.Raw.Plugins {
		if p.LongName == name || p.ShortName == name {
			return &p
		}
	}
	return nil
}

func (p *Plugins) Poll() int {
	qr := map[string]string{
		"depth": strconv.Itoa(p.Depth),
	}
	p.Jenkins.Requester.GetJSON(p.Base, p.Raw, qr)
	return p.Jenkins.Requester.LastResponse.StatusCode
}
