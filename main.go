package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// packaging https://medium.com/@mattholt/packaging-a-go-application-for-macos-f7084b00f6b5
// apple events: https://stackoverflow.com/questions/29896408/open-with-application-how-to-detect-filename-in-go
// -na "/Applications/GoLand.app"
var (
	prettify     = flag.Bool("p", true, "prettify json")
	updateEditor = flag.String("default-editor", "", "set default file editor to open files with")
)

var defaultConfig = &Config{
	Editor: "TextEdit",
}

func main() {
	flag.Parse()

	if *updateEditor != "" {
		newConfig := &Config{
			Editor: *updateEditor,
		}
		if err := updateConfig(newConfig); err != nil {
			log.Fatal(err)
		}
		return
	}

	if len(os.Args) < 2 {
		log.Fatal("missing filename - arg[1]")
	}
	config, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	filename := os.Args[1]
	if err := UnzipAndOpen(config, filename, *prettify); err != nil {
		log.Fatal(err)
	}
}

type Config struct {
	Editor string `json:"editor"`
}

func getConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig, nil
		}
		return nil, err
	}
	defer f.Close()
	var config Config
	if err = json.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func updateConfig(config *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(config)
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "jsonx.conf"), nil
}

func UnzipAndOpen(config *Config, filename string, format bool) error {
	if config == nil {
		return fmt.Errorf("no config provided")
	}
	in, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer in.Close()

	r, err := gzip.NewReader(in)
	if err != nil {
		return err
	}
	uncompresed := strings.TrimSuffix(filename, ".gz")
	out, err := os.Create(uncompresed)
	if err != nil {
		return err
	}
	defer out.Close()

	var data interface{}
	if err = json.NewDecoder(r).Decode(&data); err != nil {
		return err
	}
	enc := json.NewEncoder(out)
	if format {
		enc.SetIndent("", "\t")
	}
	if err = enc.Encode(data); err != nil {
		return err
	}
	out.Close()
	in.Close()
	return exec.Command("open", "-a", config.Editor, uncompresed).Run()
}
