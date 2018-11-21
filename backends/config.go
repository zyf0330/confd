package backends

import (
	"github.com/zyf0330/confd/util"
)

type Config struct {
	AuthToken    string     `toml:"auth_token"`
	AuthType     string     `toml:"auth_type"`
	Backend      string     `toml:"backend"`
	BasicAuth    bool       `toml:"basic_auth"`
	ClientCaKeys string     `toml:"client_cakeys"`
	ClientCert   string     `toml:"client_cert"`
	ClientKey    string     `toml:"client_key"`
	BackendNodes util.Nodes `toml:"nodes"`
	Password     string     `toml:"password"`
	Scheme       string     `toml:"scheme"`
	Table        string     `toml:"table"`
	Username     string     `toml:"username"`
	AppID        string     `toml:"app_id"`
	UserID       string     `toml:"user_id"`
	YAMLFile     util.Nodes `toml:"file"`
}
