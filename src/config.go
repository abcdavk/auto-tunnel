package main

type Config struct {
	APIToken        string `json:"apiToken"`
	ZoneID          string `json:"zoneID"`
	WebhookURL      string `json:"webhookURL"`
	DomainCNAME     string `json:"domainCNAME"`
	CloudflareAPI   string `json:"cloudflareAPI"`
	ServiceName     string `json:"serviceName"`
	IntervalMinutes int    `json:"intervalMinutes"`
	Info            string `json:"info"`
}
