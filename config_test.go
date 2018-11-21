package main

import (
	"reflect"
	"testing"

	"github.com/zyf0330/confd/log"
)

func TestInitConfigDefaultConfig(t *testing.T) {
	log.SetLevel("warn")
	want := Config{
		Backend:       "etcdv3",
		BackendNodes:  []string{"127.0.0.1:2379"},
		ClientCaKeys:  "",
		ClientCert:    "",
		ClientKey:     "",
		ConfDir:       "/etc/confd",
		Interval:      600,
		Noop:          false,
		Prefix:        "",
		SRVDomain:     "",
		Scheme:        "http",
		SecretKeyring: "",
		Table:         "",
	}
	if err := initConfig(); err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(want, config) {
		t.Errorf("initConfig() = %v, want %v", config, want)
	}
}
