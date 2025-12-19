package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"time"

	"github.com/imroc/req/v3"
)

var config *Config

// Kode utama, auto run runTunnelCycle() tiap 1 jam. Gak akan berhenti kalau ga di kill
func main() {
	var err error
	config, err = LoadConfig("config/tunnel.json")
	if err != nil {
		fmt.Println("Failed to load config:", err)
		return
	}

	for {
		runTunnelCycle()
		fmt.Println("Auto run in 1 hour...")

		RunWebhook()

		time.Sleep(47 * time.Minute)
	}
}

func runTunnelCycle() {
	fmt.Println("Running Pinggy SSH...")

	cmd := exec.Command(
		"ssh",
		"-p",
		"443",
		"-R0:127.0.0.1:25565",
		"tcp@free.pinggy.io",
	)

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	cmd.Start()

	// Auto-skip jika Pinggy minta password
	// Ga workkkk
	go func() {
		for i := 0; i < 5; i++ {
			stdin.Write([]byte("\n"))
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	scanner := bufio.NewScanner(stdout)
	re := regexp.MustCompile(`tcp://([a-zA-Z0-9.-]+):([0-9]+)`)

	fmt.Println("Waiting output...")

	var host string
	var port string

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)

		match := re.FindStringSubmatch(line)
		if len(match) == 3 {
			host = match[1]
			port = match[2]
			break
		}
	}

	// Auto reconnecting kalau failed to connect
	if host == "" {
		fmt.Println("Failed to connect Pinggy. Reconnecting in 3 sec...")
		time.Sleep(3 * time.Second)
		runTunnelCycle()
		return
	}

	fmt.Println("Public Host:", host)
	fmt.Println("Public Port:", port)

	updateDNS(host, port)

	// Run cmd ssh tanpa close
	go func() {
		cmd.Wait()
	}()

	// Just in case ssh mati mendadak
	go func() {
		time.Sleep(10 * time.Second)
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			fmt.Println("Unexpected shutdown SHH. Restartng...")
			runTunnelCycle()
		}
	}()
}

func updateDNS(host, port string) {
	client := req.C().SetCommonBearerAuthToken(config.APIToken)

	// Update CNAME sesuai dengan host Pinggy
	updateRecord(client, "CNAME", config.DomainCNAME, host)

	// Ini untuk Port yg mengarah ke CNAME
	srvName := fmt.Sprintf("%s.%s", config.ServiceName, config.DomainCNAME)
	srvData := map[string]interface{}{
		"service":  "_minecraft",
		"proto":    "_tcp",
		"name":     config.DomainCNAME,
		"priority": 0,
		"weight":   5,
		"port":     portToInt(port),
		"target":   config.DomainCNAME + ".",
	}

	// Update Port/SRV
	updateSRV(client, srvName, srvData)
}

func updateSRV(client *req.Client, name string, data map[string]interface{}) {
	r := client.R().SetQueryParams(map[string]string{
		"type": "SRV",
		"name": name,
	})

	resp, err := r.Get(fmt.Sprintf("%s/zones/%s/dns_records", config.CloudflareAPI, config.ZoneID))
	if err != nil {
		log.Fatal(err)
	}

	var result struct {
		Result []struct {
			ID string `json:"id"`
		} `json:"result"`
	}

	resp.Unmarshal(&result)

	body := map[string]interface{}{
		"type": "SRV",
		"name": name,
		"ttl":  1,
		"data": data,
	}

	if len(result.Result) > 0 {
		id := result.Result[0].ID
		fmt.Println("Updating SRV:", name)
		client.R().SetBody(body).
			Put(fmt.Sprintf("%s/zones/%s/dns_records/%s", config.CloudflareAPI, config.ZoneID, id))
		return
	}

	fmt.Println("Creating SRV:", name)
	client.R().SetBody(body).
		Post(fmt.Sprintf("%s/zones/%s/dns_records", config.CloudflareAPI, config.ZoneID))
}

func updateRecord(client *req.Client, recordType, name string, content interface{}) {
	r := client.R().SetQueryParams(map[string]string{
		"type": recordType,
		"name": name,
	})

	resp, err := r.Get(fmt.Sprintf("%s/zones/%s/dns_records", config.CloudflareAPI, config.ZoneID))
	if err != nil {
		log.Fatal(err)
	}

	var result struct {
		Result []struct {
			ID string `json:"id"`
		} `json:"result"`
	}

	resp.Unmarshal(&result)

	body := map[string]interface{}{
		"type":    recordType,
		"name":    name,
		"content": content,
		"ttl":     1,
		"proxied": false,
	}

	if len(result.Result) > 0 {
		id := result.Result[0].ID
		fmt.Println("Updating:", name)
		client.R().SetBody(body).
			Put(fmt.Sprintf("%s/zones/%s/dns_records/%s", config.CloudflareAPI, config.ZoneID, id))
		return
	}

	fmt.Println("Creating:", name)
	client.R().SetBody(body).
		Post(fmt.Sprintf("%s/zones/%s/dns_records", config.CloudflareAPI, config.ZoneID))
}

func portToInt(p string) int {
	var x int
	fmt.Sscanf(p, "%d", &x)
	return x
}
