package utils

import (
	"encoding/xml"
	"io"
	"os"
)

type NmapRun struct {
	Hosts []Host `xml:"host"`
}

type Host struct {
	Addresses []Address `xml:"address"`
	Ports     []Port    `xml:"ports>port"`
}

type Address struct {
	Addr string `xml:"addr" attr:"addr"`
}

type Port struct {
	PortID   int     `xml:"portid" attr:"portid"`
	Protocol string  `xml:"protocol" attr:"protocol"`
	State    State   `xml:"state"`
	Service  Service `xml:"service"`
}

type State struct {
	State string `xml:"state" attr:"state"`
}

type Service struct {
	Name string `xml:"name" attr:"name"`
}

func ParseNmapXML(filePath string) ([]string, error) {
	xmlFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer xmlFile.Close()

	byteValue, _ := io.ReadAll(xmlFile)
	var nmapRun NmapRun
	err = xml.Unmarshal(byteValue, &nmapRun)
	if err != nil {
		return nil, err
	}

	var targets []string
	for _, host := range nmapRun.Hosts {
		var addr string
		for _, a := range host.Addresses {
			addr = a.Addr // Usually the last one is IP
		}
		for _, p := range host.Ports {
			if p.State.State == "open" {
				targets = append(targets, addr)
			}
		}
	}

	return targets, nil
}
