package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"
	"log"
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
		log.Fatal(err)
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		log.Fatal(err)
	}

	if config.Type == "mysql" {
		command := fmt.Sprintf("MYSQL_PWD=%v mysqldump -h %v -u %v -P %v %v > /tmp/dump.sql", config.Source.Password, config.Source.Host, config.Source.Username, config.Source.Port, config.Source.Database)
		out, err := runCommandExec(command)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(out)

		command = fmt.Sprintf("MYSQL_PWD=%v mysql -h %v -u %v -P %v -e \"CREATE DATABASE IF NOT EXISTS \\`%v\\` \"", config.Destination.Password, config.Destination.Host, config.Destination.Username, config.Destination.Port, config.Destination.Database)
		out, err = runCommandExec(command)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(out)

		command = fmt.Sprintf("MYSQL_PWD=%v mysql -h %v -u %v -P %v %v < /tmp/dump.sql", config.Destination.Password, config.Destination.Host, config.Destination.Username, config.Destination.Port, config.Destination.Database)
		out, err = runCommandExec(command)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(out)
	} else if config.Type == "psql" {
		f, err := os.Create("/tmp/.pgpass")

		if err != nil {
			log.Fatal(err)
		}

		defer f.Close()

		dsource := fmt.Sprintf("%v:%v:*:%v:%v\n", config.Source.Host, config.Source.Port, config.Source.Username, config.Source.Password)
		ddest := fmt.Sprintf("%v:%v:*:%v:%v\n", config.Destination.Host, config.Destination.Port, config.Destination.Username, config.Destination.Password)
		_, err = f.WriteString(dsource + ddest)
		if err != nil {
			log.Fatal(err)
		}

		// Change permissions Linux.
		err = os.Chmod("/tmp/.pgpass", 0600)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Writing pgpass done")

		command := fmt.Sprintf("PGPASSFILE='/tmp/.pgpass' pg_dump -h %v -p %v -U %v %v > /tmp/dump.sql", config.Source.Host, config.Source.Port, config.Source.Username, config.Source.Database)
		out, err := runCommandExec(command)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(out)

		conninfo := fmt.Sprintf("host=%v port=%v user=%v password=%v sslmode=disable", config.Destination.Host, config.Destination.Port, config.Destination.Username, config.Destination.Password)
		db, err := sql.Open("postgres", conninfo)

		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE \"%v\"", config.Destination.Database))
		if err != nil {
			log.Println(err)
		}

		command = fmt.Sprintf("PGPASSFILE='/tmp/.pgpass' psql -h %v -p %v -d %v -U %v -f /tmp/dump.sql", config.Destination.Host, config.Destination.Port, config.Destination.Database, config.Destination.Username)
		out, err = runCommandExec(command)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(out)
	} else if config.Type == "mongo" {
		var URIsrc string
		var URIdest string
		if config.Source.Username == "" {
			URIsrc = fmt.Sprintf("mongodb://%v:%v/%v", config.Source.Host, config.Source.Port, config.Source.Database)
		} else {
			URIsrc = fmt.Sprintf("mongodb://%v:%v@%v:%v/%v?authSource=admin", config.Source.Username, config.Source.Password, config.Source.Host, config.Source.Port, config.Source.Database)
		}

		command := fmt.Sprintf("mongodump -v --forceTableScan --uri %v -o /tmp/dump/", URIsrc)
		out, err := runCommandExec(command)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(out)

		if config.Destination.Username == "" {
			URIdest = fmt.Sprintf("mongodb://%v:%v/%v", config.Destination.Host, config.Destination.Port, config.Destination.Database)
		} else {
			URIdest = fmt.Sprintf("mongodb://%v:%v@%v:%v/%v?authSource=admin", config.Destination.Username, config.Destination.Password, config.Source.Host, config.Source.Port, config.Destination.Database)
		}

		command = fmt.Sprintf("mongorestore --uri %v /tmp/dump/", URIdest)
		out, err = runCommandExec(command)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(out)

	} else {
		fmt.Println("Not supported")
	}
}
