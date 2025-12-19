package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mcstatus-io/mcutil/v4/status"
)

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type Embed struct {
	Title  string  `json:"title"`
	Color  int     `json:"color"`
	Fields []Field `json:"fields"`
}

type WebhookMessage struct {
	Content string  `json:"content"`
	Embeds  []Embed `json:"embeds"`
}

func RunWebhook() {
	startTime := time.Now()

	for i := 0; i < 12; i++ {
		fmt.Println("Connecting to", config.DomainCNAME, "...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		response, err := status.Modern(
			ctx,
			config.DomainCNAME,
			25565,
		)

		cancel()

		now := time.Now().Format(time.DateTime)

		if err != nil {
			fmt.Println("Failed to connect:", err)

			if i == 0 {
				msg := WebhookMessage{
					Content: "**Ahli Suargo**",
					Embeds: []Embed{
						{
							Title: "**Server Status**",
							Color: 0xFF0000,
							Fields: []Field{
								{Name: "IP Address", Value: config.DomainCNAME, Inline: true},
								{Name: "Status", Value: "Restarting", Inline: true},
								{Name: "Date/Time", Value: now, Inline: false},
							},
						},
					},
				}

				body, _ := json.Marshal(msg)
				http.Post(config.WebhookURL, "application/json", bytes.NewBuffer(body))
			}

		} else {
			endTime := time.Now()

			msg := WebhookMessage{
				Content: "**Ahli Suargo**",
				Embeds: []Embed{
					{
						Title: "**Server Status**",
						Color: 0x00AAFF,
						Fields: []Field{
							{Name: "IP Address", Value: config.DomainCNAME, Inline: true},
							{Name: "Status", Value: "Online", Inline: true},
							{
								Name:   "Ping",
								Value:  fmt.Sprintf("%dms", response.Latency.Milliseconds()),
								Inline: true,
							},
							{
								Name:   "Version",
								Value:  response.Version.Name.Clean,
								Inline: true,
							},
							{Name: "Date/Time", Value: now, Inline: false},
							{
								Name:   "Delay",
								Value:  fmt.Sprintf("%dm %ds", int(endTime.Sub(startTime).Minutes()), int(endTime.Sub(startTime).Seconds())%60),
								Inline: true,
							},
							{
								Name:   "Info",
								Value:  config.Info,
								Inline: true,
							},
						},
					},
				},
			}

			body, _ := json.Marshal(msg)
			http.Post(config.WebhookURL, "application/json", bytes.NewBuffer(body))

			fmt.Println("Server online, webhook sent.")

			return
		}

		time.Sleep(1 * time.Minute)
	}
}
