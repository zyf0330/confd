package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/zyf0330/confd/backends"
	"github.com/zyf0330/confd/log"
	"github.com/zyf0330/confd/resource/template"
)

type TemplateConfig = template.Config
type BackendsConfig = backends.Config

// A Config structure is used to configure confd.
type Config struct {
	TemplateConfig
	BackendsConfig
	Interval      int    `toml:"interval"`
	SecretKeyring string `toml:"secret_keyring"`
	SRVDomain     string `toml:"srv_domain"`
	SRVRecord     string `toml:"srv_record"`
	LogLevel      string `toml:"log-level"`
	Watch         bool   `toml:"watch"`
	PrintVersion  bool
	ConfigFile    string
	OneTime       bool
	PProf         bool
}

var config Config

func init() {
	flag.StringVar(&config.AuthToken, "auth-token", "", "Auth bearer token to use")
	flag.StringVar(&config.Backend, "backend", "etcdv3", "backend to use")
	flag.BoolVar(&config.BasicAuth, "basic-auth", false, "Use Basic Auth to authenticate (only used with -backend=consul and -backend=etcd)")
	flag.StringVar(&config.ClientCaKeys, "client-ca-keys", "", "client ca keys")
	flag.StringVar(&config.ClientCert, "client-cert", "", "the client cert")
	flag.StringVar(&config.ClientKey, "client-key", "", "the client key")
	flag.StringVar(&config.ConfDir, "confdir", "/etc/confd", "confd conf directory")
	flag.StringVar(&config.ConfigFile, "config-file", "/etc/confd/confd.toml", "the confd config file")
	flag.Var(&config.YAMLFile, "file", "the YAML file to watch for changes (only used with -backend=file)")
	flag.IntVar(&config.Interval, "interval", 600, "backend polling interval")
	flag.BoolVar(&config.KeepStageFile, "keep-stage-file", false, "keep staged files")
	flag.StringVar(&config.LogLevel, "log-level", "", "level which confd should log messages")
	flag.BoolVar(&config.PProf, "pprof", false, "enable pprof debug")
	flag.Var(&config.BackendNodes, "node", "list of backend nodes")
	flag.BoolVar(&config.Noop, "noop", false, "only show pending changes")
	flag.BoolVar(&config.OneTime, "onetime", false, "run once and exit")
	flag.StringVar(&config.Prefix, "prefix", "", "key path prefix")
	flag.BoolVar(&config.PrintVersion, "version", false, "print version and exit")
	flag.StringVar(&config.Scheme, "scheme", "http", "the backend URI scheme for nodes retrieved from DNS SRV records (http or https)")
	flag.StringVar(&config.SecretKeyring, "secret-keyring", "", "path to armored PGP secret keyring (for use with crypt functions)")
	flag.StringVar(&config.SRVDomain, "srv-domain", "", "the name of the resource record")
	flag.StringVar(&config.SRVRecord, "srv-record", "", "the SRV record to search for backends nodes. Example: _etcd-client._tcp.example.com")
	flag.BoolVar(&config.SyncOnly, "sync-only", false, "sync without check_cmd and reload_cmd")
	flag.StringVar(&config.AuthType, "auth-type", "", "Vault auth backend type to use (only used with -backend=vault)")
	flag.StringVar(&config.AppID, "app-id", "", "Vault app-id to use with the app-id backend (only used with -backend=vault and auth-type=app-id)")
	flag.StringVar(&config.UserID, "user-id", "", "Vault user-id to use with the app-id backend (only used with -backend=value and auth-type=app-id)")
	flag.StringVar(&config.Table, "table", "", "the name of the DynamoDB table (only used with -backend=dynamodb)")
	flag.StringVar(&config.Username, "username", "", "the username to authenticate as (only used with vault and etcd backends)")
	flag.StringVar(&config.Password, "password", "", "the password to authenticate with (only used with vault and etcd backends)")
	flag.BoolVar(&config.Watch, "watch", false, "enable watch support")
}

// initConfig initializes the confd configuration by first setting defaults,
// then overriding settings from the confd config file, then overriding
// settings from environment variables, and finally overriding
// settings from flags set on the command line.
// It returns an error if any.
func initConfig() error {
	_, err := os.Stat(config.ConfigFile)
	if os.IsNotExist(err) {
		log.Debug("Skipping confd config file.")
	} else {
		log.Debug("Loading " + config.ConfigFile)
		configBytes, err := ioutil.ReadFile(config.ConfigFile)
		if err != nil {
			return err
		}

		_, err = toml.Decode(string(configBytes), &config)
		if err != nil {
			return err
		}
	}

	// Update config from environment variables.
	processEnv()

	if config.SecretKeyring != "" {
		kr, err := os.Open(config.SecretKeyring)
		if err != nil {
			log.Fatal(err.Error())
		}
		defer kr.Close()
		config.PGPPrivateKey, err = ioutil.ReadAll(kr)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	if config.LogLevel != "" {
		log.SetLevel(config.LogLevel)
	}

	if config.SRVDomain != "" && config.SRVRecord == "" {
		config.SRVRecord = fmt.Sprintf("_%s._tcp.%s.", config.Backend, config.SRVDomain)
	}

	// Update BackendNodes from SRV records.
	if config.SRVRecord != "" {
		log.Info("SRV record set to " + config.SRVRecord)
		srvNodes, err := getBackendNodesFromSRV(config.SRVRecord)
		if err != nil {
			return errors.New("Cannot get nodes from SRV records " + err.Error())
		}

		config.BackendNodes = srvNodes
	}
	if len(config.BackendNodes) == 0 {
		config.BackendNodes = []string{"127.0.0.1:2379"}
	}
	// Initialize the storage client
	log.Info("Backend set to " + config.Backend)

	config.ConfigDir = filepath.Join(config.ConfDir, "conf.d")
	config.TemplateDir = filepath.Join(config.ConfDir, "templates")
	return nil
}

func getBackendNodesFromSRV(record string) ([]string, error) {
	nodes := make([]string, 0)

	// Ignore the CNAME as we don't need it.
	_, addrs, err := net.LookupSRV("", "", record)
	if err != nil {
		return nodes, err
	}
	for _, srv := range addrs {
		host := strings.TrimRight(srv.Target, ".")
		port := strconv.FormatUint(uint64(srv.Port), 10)
		nodes = append(nodes, net.JoinHostPort(host, port))
	}
	return nodes, nil
}

func processEnv() {
	cakeys := os.Getenv("CONFD_CLIENT_CAKEYS")
	if len(cakeys) > 0 && config.ClientCaKeys == "" {
		config.ClientCaKeys = cakeys
	}

	cert := os.Getenv("CONFD_CLIENT_CERT")
	if len(cert) > 0 && config.ClientCert == "" {
		config.ClientCert = cert
	}

	key := os.Getenv("CONFD_CLIENT_KEY")
	if len(key) > 0 && config.ClientKey == "" {
		config.ClientKey = key
	}
}
