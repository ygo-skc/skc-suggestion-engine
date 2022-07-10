package model

type Status struct {
	Version    string           `json:"version"`
	Downstream []DownstreamItem `json:"downstream"`
}

type DownstreamItem struct {
	ServiceName string `json:"serviceName"`
	Status      string `json:"status"`
}
