
package main

import (
	"testing"
)

func TestParseConfig(t *testing.T) {
	_, err := ParseConfigFile("frontline.conf")
	if err != nil {
		t.Errorf("parse config: %s", err)
	}
}
