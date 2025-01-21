package config

type Http struct {
	URL  string            `json:"url"`
	With map[string]string `json:"with"`
}
