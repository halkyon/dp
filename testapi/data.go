package testapi

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed testdata/servers.yml
var serversData []byte

//go:embed testdata/locations.yml
var locationsData []byte

type serverEntry struct {
	Name        string `yaml:"name"`
	Alias       string `yaml:"alias"`
	Hostname    string `yaml:"hostname"`
	Uptime      int    `yaml:"uptime"`
	StatusV2    string `yaml:"statusV2"`
	PowerStatus string `yaml:"powerStatus"`
	Location    struct {
		Name   string `yaml:"name"`
		Region string `yaml:"region"`
	} `yaml:"location"`
	System struct {
		OperatingSystem struct {
			Name string `yaml:"name"`
		} `yaml:"operatingSystem"`
		Raid string `yaml:"raid"`
	} `yaml:"system"`
	Hardware struct {
		CPUs    []cpuInfo     `yaml:"cpus"`
		Storage []storageInfo `yaml:"storage"`
		Rams    []ramInfo     `yaml:"rams"`
	} `yaml:"hardware"`
	Network struct {
		IPAddresses        []ipAddress `yaml:"ipAddresses"`
		UplinkCapacity     int         `yaml:"uplinkCapacity"`
		HasBGP             bool        `yaml:"hasBgp"`
		HasLinkAggregation bool        `yaml:"hasLinkAggregation"`
		DDOSShieldLevel    string      `yaml:"ddosShieldLevel"`
		IPMI               ipmiInfo    `yaml:"ipmi"`
	} `yaml:"network"`
	TrafficPlan struct {
		Name      string  `yaml:"name"`
		Type      string  `yaml:"type"`
		Bandwidth float64 `yaml:"bandwidth"`
	} `yaml:"trafficPlan"`
	Billing struct {
		SubscriptionItem struct {
			Price                  float64 `yaml:"price"`
			Currency               string  `yaml:"currency"`
			SubscriptionItemDetail struct {
				Server struct {
					Name string `yaml:"name"`
				} `yaml:"server"`
			} `yaml:"subscriptionItemDetail"`
		} `yaml:"subscriptionItem"`
	} `yaml:"billing"`
	Tags []tagInfo `yaml:"tags"`
}

type cpuInfo struct {
	Name  string `yaml:"name"`
	Cores int    `yaml:"cores"`
}

type storageInfo struct {
	Size int    `yaml:"size"`
	Type string `yaml:"type"`
}

type ramInfo struct {
	Size int `yaml:"size"`
}

type ipAddress struct {
	IP          string `yaml:"ip"`
	IsPrimary   bool   `yaml:"isPrimary"`
	Type        string `yaml:"type"`
	IsBgpPrefix bool   `yaml:"isBgpPrefix"`
}

type ipmiInfo struct {
	IP       string `yaml:"ip"`
	Username string `yaml:"username"`
}

type tagInfo struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type locationEntry struct {
	Name   string `yaml:"name"`
	Region string `yaml:"region"`
}

type serversResponse struct {
	Entries []serverEntry `yaml:"entries"`
}

type locationsResponse []locationEntry

func loadTestData() ([]serverEntry, []locationEntry, error) {
	var servers serversResponse
	if err := yaml.Unmarshal(serversData, &servers); err != nil {
		return nil, nil, fmt.Errorf("unmarshaling servers: %w", err)
	}

	var locations locationsResponse
	if err := yaml.Unmarshal(locationsData, &locations); err != nil {
		return nil, nil, fmt.Errorf("unmarshaling locations: %w", err)
	}

	return servers.Entries, locations, nil
}
