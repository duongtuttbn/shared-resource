package corefx

import (
	"bytes"
	"encoding/json"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
	"strings"
)

func CreateConfigLoader[T any](initialValue T) func() (T, error) {
	return func() (T, error) {
		return LoadConfigFromEnv(initialValue)
	}
}

func LoadConfigFromEnv[T any](initialValue T) (T, error) {
	c := initialValue
	// --- hacking to load reflect structure config into env ----//
	viper.SetConfigType("json")
	configBuffer, err := json.Marshal(c)
	if err != nil {
		var zero T
		return zero, err
	}

	if err := viper.ReadConfig(bytes.NewBuffer(configBuffer)); err != nil {
		panic(err)
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// -- end of hacking --//
	viper.AutomaticEnv()
	err = viper.Unmarshal(&c, func(config *mapstructure.DecoderConfig) {
		config.TagName = "json"
		config.Squash = true
	})
	if err != nil {
		var zero T
		return zero, err
	}
	return c, nil
}
