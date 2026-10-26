// Command gen_license generates a signed南天门 AI 网关 enterprise license for
// local testing. It uses the DEMO vendor private key (must match the public key
// embedded in module/license/signer.go). For production, replace both keys with
// the vendor's real Ed25519 keypair and keep the private key secret.
//
// Usage:
//
//	go run ./scripts/gen_license.go [-days 365] [-customer "Acme"] [-modules advanced_policy,sso,audit,multi_cluster,sla] [-nodes 0]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/RiemannChow/nantianmen-ai-gateway/module/license"
)

// demoPrivateKey matches the public key baked into module/license/signer.go.
const demoPrivateKey = "c3UgqulxPN7d53wrWY9btsv9foHi7xhJ76BwXOA7XpifqMSbtjF+z/DoMXzSeDlQLqgzFMKdG+rirfAKq9xlzw=="

func main() {
	days := flag.Int("days", 365, "validity in days (0 = never expire)")
	customer := flag.String("customer", "Demo Enterprise", "licensee name")
	modules := flag.String("modules", "advanced_policy,sso,audit,multi_cluster,sla", "comma-separated modules")
	nodes := flag.Int("nodes", 0, "max gateway nodes (0 = unlimited)")
	sla := flag.String("sla", "Gold (99.9%)", "SLA tier label")
	flag.Parse()

	now := time.Now()
	expiry := int64(0)
	if *days > 0 {
		expiry = now.AddDate(0, 0, *days).Unix()
	}

	l := &license.License{
		Edition:  license.EditionEnterprise,
		Customer: *customer,
		Modules:  splitModules(*modules),
		MaxNodes: *nodes,
		IssuedAt: now.Unix(),
		Expiry:   expiry,
		SLA:      *sla,
	}

	signed, err := license.Sign(demoPrivateKey, l)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sign error:", err)
		os.Exit(1)
	}
	out, _ := json.MarshalIndent(signed, "", "  ")
	fmt.Println(string(out))
}

func splitModules(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	cur := ""
	for _, r := range s {
		if r == ',' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
