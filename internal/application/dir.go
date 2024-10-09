package application

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var cachedHugoverseDir string

func init() {
	cachedHugoverseDir = hugoverseDir()

	err := ensureDirExists(cachedHugoverseDir)
	if err != nil {
		log.Fatalln(err)
	}
}

func TLSDir() string {
	return filepath.Join(DataDir(), "tls")
}

func UploadDir() string {
	return filepath.Join(DataDir(), "uploads")
}

func DataDir() string {
	return cachedHugoverseDir
}

func hugoverseDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err, "using current directory as working directory")

		return getWd()
	}

	// 构建目录路径 ~/.local/share/hugoverse
	hugoverseDir := filepath.Join(homeDir, ".local", "share", "hugoverse")

	return hugoverseDir
}

func getWd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Couldn't find working directory", err)
	}
	return wd
}

func ensureDirExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	} else if err != nil {
		// 其他错误
		return fmt.Errorf("failed to check directory: %w", err)
	}
	return nil
}
