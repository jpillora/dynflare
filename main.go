package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/jpillora/go-tld"
	"github.com/jpillora/ipflare"
	"github.com/jpillora/opts"
)

var c = struct {
	Interval time.Duration `opts:"help=polling interval"`
	Token    string        `opts:"short=t,env,help=cloudflare token"`
	Domain   string        `opts:"short=d,env,help=domain"`
	DryRun   bool          `opts:"help=only print dns updates"`
}{
	Interval: 5 * time.Minute,
}

func main() {
	opts.New(&c).Parse()
	z, d, err := prep()
	if err != nil {
		log.Fatal(err)
	}
	if err := run(z, d); err != nil {
		log.Fatal(err)
	}
}

func prep() (z zone, d record, err error) {
	if c.Token == "" {
		return z, d, errors.New("missing cloudflare --token")
	}
	if c.Domain == "" {
		return z, d, errors.New("missing target --domain")
	}
	u, err := tld.Parse("http://" + c.Domain)
	if err != nil {
		return z, d, errors.New("invalid domain")
	}
	root := u.Domain + "." + u.TLD
	target := root
	if s := u.Subdomain; s != "" {
		target = s + "." + root
	}
	//get all zones
	zones := []zone{}
	if err := cf("GET", "/zones", nil, &zones); err != nil {
		return z, d, err
	}
	//get my zone
	for _, zone := range zones {
		if zone.Name == root {
			z = zone
		}
	}
	if z.ID == "" {
		return z, d, errors.New("CF account is missing " + root)
	}
	//get zones records
	records := []record{}
	if err := cf("GET", "/zones/"+z.ID+"/dns_records", nil, &records); err != nil {
		return z, d, err
	}
	//get records, find target
	d = record{}
	for _, r := range records {
		if r.Name == target {
			d = r
			break
		}
	}
	if d.Name == "" {
		return z, d, errors.New("cannot DNS record for " + target)
	}
	log.Printf("found record %s (%s)", d.Name, d.Content)
	return z, d, nil
}

func run(z zone, d record) error {
	first := true
	for {
		//get public ip
		public, err := ipflare.My()
		if err != nil {
			return err
		}
		if public.To4() == nil {
			return fmt.Errorf("public ip is v6 (%s) only v4 supported", public)
		}
		ip := public.String()
		//status message
		if first {
			log.Printf("watching public ip (%s) for changes...", ip)
			first = false
		}
		// changed?
		if d.Content != ip {
			d.Content = ip
			if err := update(z, d, ip); err != nil {
				return err
			}
		}
		time.Sleep(c.Interval)
	}
}

func update(z zone, d record, newIP string) error {
	if c.DryRun {
		log.Printf("[DRYRUN] skip update record %s to %s", d.Name, newIP)
	} else {
		if err := cf("PUT", "/zones/"+z.ID+"/dns_records/"+d.ID, &d, nil); err != nil {
			return fmt.Errorf("updating %s to %s failed: %s", d.Name, newIP, err)
		}
		log.Printf("updated record %s to %s", d.Name, newIP)
	}
	return nil
}

type zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type record struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

//cf is a basic client for the v4 api
func cf(method, url string, input, output interface{}) error {
	wrapper := struct {
		Result  interface{} `json:"result"`
		Success bool
		Errors  []struct {
			Code    int
			Message string
		}
		Messages []string
	}{
		Result: output,
	}
	const base = "https://api.cloudflare.com/client/v4"
	var r io.Reader
	if input != nil {
		b, _ := json.Marshal(input)
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, base+url, r)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	tmp := bytes.Buffer{}
	r = io.TeeReader(resp.Body, &tmp)
	if err := json.NewDecoder(r).Decode(&wrapper); err != nil {
		return fmt.Errorf("json error: %s\n%s", err, tmp.String())
	}
	if !wrapper.Success {
		for _, e := range wrapper.Errors {
			if e.Message != "" {
				return errors.New(e.Message)
			}
		}
		return errors.New("unknown error")
	}
	return nil
}
