
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"json"
	"os"
	"strings"
	"sync"
)

type Config struct {
	sync.Mutex

	Timeout int64 // Keep-alive timeout in nanoseconds
	FDLimit int   // Maximum number of file descriptors
	hosts	map[string][]string	// virtual host name -> array of actual net addr of server
}

func (c *Config) String() string {
	c.Lock()
	defer c.Unlock()

	var w bytes.Buffer
	fmt.Fprintf(&w, "Timeout=%g ns, FDLimit=%d\n", float64(c.Timeout), c.FDLimit)
	for v, aa := range c.hosts {
		fmt.Fprintf(&w, "%s->\n    ", v)
		for _, a := range aa {
			fmt.Fprintf(&w, "%s, ", a)
		}
		fmt.Fprintln(&w, "")
	}
	return string(w.Bytes())
}

func ParseConfigFile(filename string) (*Config, os.Error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	return ParseConfigMap(m)
}

func ParseConfigMap(m map[string]interface{}) (*Config, os.Error) {
	c := &Config{
		Timeout: 5e9,
		FDLimit: 200,
		hosts:   make(map[string][]string),
	}
	// Timeout
	tmo_, ok := m["Timeout"]
	if ok {
		if tmo, ok := tmo_.(float64); ok {
			c.Timeout = int64(tmo)
		}
	}
	// FDLimit
	fdl_, ok := m["FDLimit"]
	if ok {
		if fdl, ok := fdl_.(float64); ok {
			c.FDLimit = int(fdl)
		}
	}
	// Virtual hosts
	for _, w_ := range getSliceInterface(m["Virtual"]) {
		w := getMapStringInterface(w_)
		vhosts := getSliceInterface(w["VHosts"])
		ahosts := getSliceInterface(w["AHosts"])
		a := []string{}
		for _, ah_ := range ahosts {
			ah := strings.TrimSpace(getString(ah_))
			if ah != "" {
				a = append(a, strings.ToLower(ah))
			}
		}
		if len(a) == 0 {
			continue
		}
		for _, vh_ := range vhosts {
			vh := strings.TrimSpace(getString(vh_))
			if vh != "" {
				c.hosts[strings.ToLower(vh)] = a
			}
		}
	}
	return c, nil
}

func getString(s_ interface{}) string {
	if s_ == nil {
		return ""
	}
	if s, ok := s_.(string); ok {
		return s
	}
	return ""
}

func getMapStringInterface(v_ interface{}) map[string]interface{} {
	if v_ == nil {
		return make(map[string]interface{})
	}
	if v, ok := v_.(map[string]interface{}); ok {
		return v
	}
	return make(map[string]interface{})
}

func getSliceInterface(v_ interface{}) []interface{} {
	if v_ == nil {
		return []interface{}{}
	}
	if v, ok := v_.([]interface{}); ok {
		return v
	}
	return []interface{}{}
}

func (c *Config) ActualHost(vhost string) string {
	c.Lock()
	defer c.Unlock()

	aa, ok := c.hosts[vhost]
	if !ok {
		return ""
	}
	return aa[0]
}
