### How to use

Rename `config/tunnel.json.temp` to `config/tunnel.json`. Then fill the properties correctly.

```json
{
    "apiToken": "CLOUDFLARE API TOKEN",
	"zoneID": "CLOUDFLARE ZONE ID",
    "webhookURL": "DISCORD WEB HOOK URL",
    "domainCNAME": "play.example.com",
    "cloudflareAPI": "https://api.cloudflare.com/client/v4",
	"serviceName": "_test._play",
    "intervalMinutes": 5,

    "info": "Info message"
}
```

### Run, Build, Execute on Linux

To run: `go run */**.go`

To build: `go build -o build/auto-tunnel ./src`

To execute:
    You need config file to run this program. You can just copy the `config/tunnel.json` on the same folder of `auto-tunnel` executable. Then run `./auto-tunnel`