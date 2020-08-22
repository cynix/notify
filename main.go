package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)


func main() {
	hostname, _ := os.Hostname()
	hostname = strings.TrimSuffix(strings.TrimSuffix(hostname, ".h.cynix.org"), ".g.cynix.org")

	c := pflag.StringP("config", "c", "/usr/local/etc/notify.yaml", "Config file")
	t := pflag.StringP("title", "t", hostname, "Title")
	u := pflag.StringP("url", "u", "", "URL")

	pflag.Parse()

	var cfg Config
	if err := cfg.Load(*c); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse config %s: %v\n", *c, err)
		os.Exit(1)
	}

	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read stdin: %v\n", err)
		os.Exit(1)
	}

	data := url.Values{}
	data.Set("value1", *t)
	data.Set("value2", strings.TrimSpace(string(b)))

	if len(*u) > 0 {
		data.Set("value3", *u)
	}

	resp, err := http.PostForm("https://maker.ifttt.com/trigger/notify/with/key/" + cfg.Ifttt.Key, data)

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to send notification: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "failed to send notification: %d\n", resp.StatusCode)
		os.Exit(1)
	}
}


type Config struct {
	Ifttt struct {
		Key string
	}
}

func (c *Config) Load(file string) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, c)
	if err != nil {
		return err
	}

	return nil
}
