package main

import (
	"bytes"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"os/exec"
)

type Config struct {
	Type   string `yaml:"type"`
	Source struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"source"`

	Destination struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"destination"`
}

func runCommandExec(cmdinput string) (string, error) {
	fmt.Println("[DEBUG] Executing " + cmdinput)
	cmd := exec.Command("/bin/bash", "-c", cmdinput)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", errors.New(fmt.Sprint(err) + ": " + stderr.String())
	} else {
		return out.String(), nil
	}
}

func main() {
	config := &Config{}

	// Open config file
	file, err := os.Open("config.yaml")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		fmt.Println(err)
	}

	if config.Type == "mysql" {
		command := fmt.Sprintf("MYSQL_PWD=%v mysqldump -h %v -u %v -P %v %v > /tmp/dump.sql", config.Source.Password, config.Source.Host, config.Source.Username, config.Source.Port, config.Source.Database)
		out, err := runCommandExec(command)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(out)

		command = fmt.Sprintf("MYSQL_PWD=%v mysql -h %v -u %v -P %v -e \"CREATE DATABASE IF NOT EXISTS \\`%v\\` \"", config.Destination.Password, config.Destination.Host, config.Destination.Username, config.Destination.Port, config.Destination.Database)
		out, err = runCommandExec(command)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(out)

		command = fmt.Sprintf("MYSQL_PWD=%v mysql -h %v -u %v -P %v %v < /tmp/dump.sql", config.Destination.Password, config.Destination.Host, config.Destination.Username, config.Destination.Port, config.Destination.Database)
		out, err = runCommandExec(command)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(out)
	}
}
