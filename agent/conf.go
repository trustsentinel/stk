package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type configuration struct {
	Hostname string `yaml:"hostname"`
	Port     int    `yaml:"port"`
}

func getEnvironment() string {
	env := os.Getenv("SENV")
	if env == "" {
		env = "development"
	}
	return env
}

func (c *configuration) getConfiguration() {

	filename := fmt.Sprintf("conf.%s.yaml", getEnvironment())
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("[conf] Error #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("[conf] configuration bad built: %v", err)
	}
	log.Println("[conf] configuration values read from", filename)
}

/*func main() {
	var c conf
	c.getConf()

	fmt.Println(c)
}*/
