package config

import (
	"bytes"
	"fmt"

	"github.com/hyperledger/burrow/logging/structure"

	"encoding/json"

	"github.com/BurntSushi/toml"
)

type LoggingConfig struct {
	RootSink *SinkConfig `toml:",omitempty"`
}

// For encoding a top-level '[logging]' TOML table
type LoggingConfigWrapper struct {
	Logging *LoggingConfig `toml:",omitempty"`
}

func DefaultNodeLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		RootSink: Sink().SetOutput(StderrOutput()),
	}
}

func DefaultClientLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		// No output
		RootSink: Sink().
			SetTransform(FilterTransform(ExcludeWhenAnyMatches,
				structure.CapturedLoggingSourceKey, "")).
			SetOutput(StderrOutput()),
	}
}

// Returns the TOML for a top-level logging config wrapped with [logging]
func (lc *LoggingConfig) RootTOMLString() string {
	return TOMLString(LoggingConfigWrapper{lc})
}

func (lc *LoggingConfig) TOMLString() string {
	return TOMLString(lc)
}

func (lc *LoggingConfig) RootJSONString() string {
	return JSONString(LoggingConfigWrapper{lc})
}

func (lc *LoggingConfig) JSONString() string {
	return JSONString(lc)
}

func LoggingConfigFromMap(loggingRootMap map[string]interface{}) (*LoggingConfig, error) {
	lc := new(LoggingConfig)
	buf := new(bytes.Buffer)
	enc := toml.NewEncoder(buf)
	// TODO: [Silas] consider using strongly typed config/struct mapping everywhere
	// (!! unfortunately the way we are using viper
	// to pass around a untyped bag of config means that we don't get keys mapped
	// according to their metadata `toml:"Name"` tags. So we are re-encoding to toml
	// and then decoding into the strongly type struct as a work-around)
	// Encode the map back to TOML
	err := enc.Encode(loggingRootMap)
	if err != nil {
		return nil, err
	}
	// Decode into struct into the LoggingConfig struct
	_, err = toml.Decode(buf.String(), lc)
	if err != nil {
		return nil, err
	}
	return lc, nil
}

func TOMLString(v interface{}) string {
	buf := new(bytes.Buffer)
	encoder := toml.NewEncoder(buf)
	err := encoder.Encode(v)
	if err != nil {
		// Seems like a reasonable compromise to make the string function clean
		return fmt.Sprintf("Error encoding TOML: %s", err)
	}
	return buf.String()
}

func JSONString(v interface{}) string {
	bs, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return fmt.Sprintf("Error encoding JSON: %s", err)
	}
	return string(bs)
}
