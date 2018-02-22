package backends

import (
	"strings"

	"github.com/zyf0330/confd/backends/etcdv3"
	"github.com/zyf0330/confd/log"
)

// The StoreClient interface is implemented by objects that can retrieve
// key/value pairs from a backend store.
type StoreClient interface {
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error)
}

// New is used to create a storage client based on our configuration.
func New(config Config) (StoreClient, error) {
	if config.Backend == "" {
		config.Backend = "etcdv3"
	}
	backendNodes := config.BackendNodes

	log.Info("Backend source(s) set to " + strings.Join(backendNodes, ", "))

	return etcdv3.NewEtcdClient(backendNodes, config.ClientCert, config.ClientKey, config.ClientCaKeys, config.BasicAuth, config.Username, config.Password)
}
