package utils

import (
	"errors"
	"github.com/hyt-hz/wxOpenID/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
)

var (
	ErrConfigFilePathInvalid = errors.New("Invalid config file path given")
)

// parse YAML config file into struct
// configPath can be either directory of file
// if directory, filename will be appended
// if file, filename is ignored
func ParseConfigFile(configPath string, filename string, option interface{}) error {

	filepath, err := GetFilePath(configPath, filename)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Error("Failed to read file %s: %s", filepath, err)
		return err
	}

	err = yaml.Unmarshal([]byte(data), option)
	if err != nil {
		log.Error("Failed to parse config file %s: %s", filepath, err)
		return err
	}

	return nil
}

func GetFilePath(configPath string, filename string) (string, error) {

	if configPath == "" {
		return "", ErrConfigFilePathInvalid
	}

	f, err := os.Open(configPath)
	if err != nil {
		log.Error("Failed to open '%s': %s", configPath, err)
		return "", err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Error("Failed to stat '%s': %s", configPath, err)
		return "", err
	}

	switch mode := fi.Mode(); {
	case mode.IsDir():
		return path.Join(configPath, filename), nil
	case mode.IsRegular():
		return configPath, nil
	default:
		return "", ErrConfigFilePathInvalid
	}
}
