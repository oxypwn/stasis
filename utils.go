package main

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"regexp"
	//"path/filepath"
)

func ValidateHostName(name string) (string, error) {
	validHostNamePattern := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validHostNamePattern.MatchString(name) {
		log.Errorf("Invalid host name %q, it must match %s", name, validHostNamePattern)
		os.Exit(1)
	}
	return name, nil
}

func ValidateMacaddr(mac string) (string, error) {
	validMacaddrPattern := regexp.MustCompile(`^([0-9A-F]{2}[-]){5}([0-9A-F]{2})+$`)
	if !validMacaddrPattern.MatchString(mac) {
		log.Errorf("Invalid mac address %q, it must match %s", mac, validMacaddrPattern)
		os.Exit(1)
	} 
	return mac, nil
}

