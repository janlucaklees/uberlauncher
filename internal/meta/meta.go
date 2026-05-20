package meta

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	AppName        = "uberlauncher"
	ConfigFileName = "config.toml"
)

func GetConfigPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(base, AppName, ConfigFileName)

	err = ensureFileExists(path)
	if err != nil {
		return "", err
	}

	return path, nil
}

func GetCacheRootPath() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(base, AppName)

	err = ensurePathExists(path)
	if err != nil {
		return "", err
	}

	return path, nil
}

func ensureFileExists(path string) error {
	// Make sure the path to the file exists.
	err := ensurePathExists(filepath.Dir(path))
	if err != nil {
		return err
	}

	// Try stat the file.
	info, err := os.Stat(path)

	// Exit early if the given path is a dir.
	if err == nil && info.IsDir() {
		return fmt.Errorf("%s exists and is not a regular file", path)
	}

	// Create file if it doesn't exist
	if err != nil && os.IsNotExist(err) {
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0o644)
		if err != nil {
			return err
		}
		return file.Close()
	}

	if err != nil {
		return err
	}

	return nil
}

func ensurePathExists(path string) error {
	return os.MkdirAll(path, 0o755)
}
