package main

import (
	"reflect"
	"testing"

	"github.com/zyf0330/confd/log"
)

func TestInitConfigDefaultConfig(t *testing.T) {
	log.SetLevel("warn")
	want := Config{
		BackendsConfig: BackendsConfig{
			Backend:      "etcdv3",
			BackendNodes: []string{"127.0.0.1:2379"},
			Scheme:       "http",
		},
		TemplateConfig: TemplateConfig{
			ConfDir:     "/etc/confd",
			ConfigDir:   "/etc/confd/conf.d",
			TemplateDir: "/etc/confd/templates",
			Noop:        false,
		},
		ConfigFile: "/etc/confd/confd.toml",
		Interval:   600,
	}
	if err := initConfig(); err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(want, config) {
		t.Errorf("initConfig() = %v, want %v", config, want)
	}
}
