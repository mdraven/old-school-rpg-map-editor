package configuration

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
)

type Config struct {
	MainWindowHeight int `json:"main-window-height"`
	MainWindowWidth  int `json:"main-window-width"`

	MainWindowX int `json:"main-window-x"`
	MainWindowY int `json:"main-window-y"`

	ContentOffset float32 `json:"content-offset"`
	ToolsOffset   float32 `json:"tools-offset"`
}

type ImageConfig struct {
	FloorSize uint
	WallWidth uint
}

func GetConfigFile(fileName string) (*os.File, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(dir, "old-school-rpg-map-editor")

	err = os.Mkdir(configDir, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return nil, err
	}

	return os.OpenFile(filepath.Join(configDir, fileName), os.O_CREATE|os.O_RDWR, 0o700)
}

func LoadConfig(f *os.File) (*Config, error) {
	_, err := f.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var config Config

	if len(data) > 0 {
		err = json.Unmarshal(data, &config)
		if err != nil {
			return nil, err
		}
	} else {
		config = Config{
			MainWindowWidth:  800,
			MainWindowHeight: 600,
			ContentOffset:    0.7,
			ToolsOffset:      0.5,
		}
	}

	return &config, nil
}

func SaveConfig(f *os.File, c *Config) error {
	_, err := f.Seek(0, 0)
	if err != nil {
		return err
	}

	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	err = f.Truncate(0)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	err = f.Sync()
	if err != nil {
		return err
	}

	return nil
}
