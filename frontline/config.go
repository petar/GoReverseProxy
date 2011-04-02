
package main

import (
	"fmt"
	"io/ioutil"
	"json"
	"os"
	"sync"
)

type Config struct {
	sync.Mutex

	Timeout int64 // Keep-alive timeout in nanoseconds
	FDLimit int   // Maximum number of file descriptors
	hosts	map[string][]string	// virtual host name -> array of actual net addr of server
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
	fmt.Printf("%#v\n", m)
	return nil, nil
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
