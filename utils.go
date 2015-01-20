package main

import (
	"io/ioutil"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"html/template"
	"os"
	"regexp"
	"path/filepath"
)

func ValidateHostName(name string) (string, error) {
	validHostNamePattern := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validHostNamePattern.MatchString(name) {
		log.Errorf("Invalid host name %q, it must match %s", name, validHostNamePattern)
	}
	return name, nil
}

func ValidateMacaddr(mac string) (string, error) {
	validMacaddrPattern := regexp.MustCompile(`^([0-9a-fA-F]{2}[-]){5}([0-9a-fA-F]{2})+$`)
	if !validMacaddrPattern.MatchString(mac) {
		return mac, fmt.Errorf("Invalid mac address %q, it must match %s", mac, validMacaddrPattern)

	}
	return mac, nil
}

func ValidateTemplates(path, extension string) {
	filenames := []string{}
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == extension {
			filenames = append(filenames, path)
		}
		return nil
	})

	if err != nil {
		log.Fatalln(err)
	}

	if len(filenames) == 0 {
		log.Errorf("There is no %s templates in: %q", extension, path)
		os.Exit(1)
	}

	templates, err = template.ParseFiles(filenames...)
	if err != nil {
		log.Fatalln(err)
	}
}


func listTemplates(path string) {
	files, _ := ioutil.ReadDir(path)
    for _, f := range files {
            fmt.Println(f.Name())
    }
}