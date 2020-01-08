package conf

import (
	"io/ioutil"
	"time"

	"github.com/BurntSushi/toml"
)

var Cfg Config

type (
	Config struct {
		Global    Global      `toml:"global"`
		Callbacks []*Callback `toml:"callback"`
	}

	Global struct {
		LogPath    string        `toml:"log_path"`
		LogDebug   bool          `toml:"log_debug"`
		Listen     string        `toml:"listen"`
		NSQAddr    string        `toml:"nsq_lookupd_addr"`
		TimerTopic string        `toml:"timer_topic"`
		TimerCycle time.Duration `toml:"timer_cycle"`
	}

	Callback struct {
		Route string `toml:"route"`
		Bash  string `toml:"verify_bash"`
	}

	RouteInfo struct {
		Bash string
	}
)

func LoadConfig(f string) error {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	if _, err := toml.Decode(string(data), &Cfg); err != nil {
		return err
	}

	return nil
}
