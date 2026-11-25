package gowok

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gowok/fp/maps"
	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/json"
	"github.com/joho/godotenv"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

var configFiles []string

func ConfigFiles(files ...string) *Project {
	configFiles = append([]string(nil), files...)
	p := _project()
	return *p
}

func configFilesDiscover(dir string) string {
	for _, name := range configFiles {
		filePath := name
		if !filepath.IsAbs(filePath) {
			filePath = filepath.Join(dir, filePath)
		}

		info, err := os.Stat(filePath)
		if err != nil || info.IsDir() {
			continue
		}
		return filePath
	}

	return ""
}

func runningUnderTest() bool {
	return strings.Contains(filepath.Base(os.Args[0]), ".test")
}

var Config = &config.Config{}

func newConfigEmpty() (*config.Config, map[string]any) {
	conf := &config.Config{
		SQLs:   make(map[string]config.SQL),
		Smtp:   make(map[string]config.Smtp),
		Others: make(map[string]any),
	}

	confRaw := maps.FromStruct(conf)
	return conf, confRaw
}

func newConfig(pathConfig string, envFile string) (*config.Config, map[string]any, error) {
	fiConfig, err := os.Open(pathConfig)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = fiConfig.Close()
	}()

	fiContent, err := io.ReadAll(fiConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("can't read config file: %w", err)
	}

	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			return nil, nil, fmt.Errorf("can't load .env file: %w", err)
		}
	}

	cfgS := os.ExpandEnv(string(fiContent))
	fiContent = []byte(cfgS)

	confRaw, err := newConfigRaw(path.Ext(fiConfig.Name()), string(fiContent))
	if err != nil {
		return nil, nil, err
	}

	conf := &config.Config{}
	err = maps.ToStruct(confRaw, conf)
	if err != nil {
		return nil, nil, fmt.Errorf("can't decode config file: %w", err)
	}

	return conf, confRaw, nil
}

func newConfigRaw(filetype string, configString string) (map[string]any, error) {
	confRaw := map[string]any{}
	var err error

	switch strings.ToLower(filetype) {
	case ".json":
		err = json.Unmarshal([]byte(configString), &confRaw)
	case ".yaml", ".yml":
		err = yaml.Unmarshal([]byte(configString), &confRaw)
	case ".toml":
		err = toml.Unmarshal([]byte(configString), &confRaw)
	default:
		return nil, fmt.Errorf("unsupported config file type: %q", filetype)
	}
	if err != nil {
		return nil, err
	}

	return confRaw, nil
}
