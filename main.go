package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/jpillora/go-tld"
	"github.com/jpillora/opts"
)

var c = struct {
	Token  string `opts:"short=t,env,help=cloudflare token"`
	Domain string `opts:"short=d,env,help=domain"`
}{}

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
	domain := record{}
	for _, r := range records {
		if r.Name == target {
			domain = r
			break
		}
	}
	if domain.Name == "" {
		return z, d, errors.New("cannot DNS record for " + target)
	}
	return z, d, nil
}

func run(z zone, d record) error {
	//get public ip
	last, err := myIP()
	if err != nil {
		return err
	}
	log.Printf("watching public ip (%s) for changes", last)
	for {
		time.Sleep(5 * time.Minute)

		//get public ip again,

		//changed?

		// if err := update(domain, ""); err != nil {
		// 	return err
		// }
	}
}

func update(z zone, d record, newIP string) error {
	d.Content = newIP
	if err := cf("PUT", "/zones/"+z.ID+"/dns_records/"+d.ID, &d, nil); err != nil {
		return err
	}
	//TODO check output
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

//cf is a basic client
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
		return errors.New("err")
	}
	return nil
}

var myIPre = regexp.MustCompile(`<span>Your IP</span>: ([\w+\.:]+)</span>`)

func myIP() (string, error) {
	resp, err := http.Get("https://www.cloudflare.com/learning/dns/glossary/what-is-my-ip-address/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	m := myIPre.FindSubmatch(b)
	if len(m) == 0 {
		return "", errors.New("my-ip not found on page")
	}
	s := string(m[1])
	ip := net.ParseIP(s)
	if ip == nil {
		return "", fmt.Errorf("my-ip invalid (%s)", s)
	}
	if ip.To4() == nil {
		return "", fmt.Errorf("my-ip invalid ipv4 (%s)", s)
	}
	return ip.String(), nil
}
