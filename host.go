package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Host struct
type Host struct {
	Name               string
	StorePath          string
	Macaddress         string
	Preinstall         string
	PermitPreinstall bool
	Install            string
	PermitInstall    bool
	Username           string
	Password           string
	Postinstall        string
	PermitPostinstall   bool
	WindowsKey         string
	Append             string
	Mirror             string
	Kernel             string
	Initrd             string
	Status             string
	Announce           bool
}

func NewHost(
	name,
	storePath,
	mac,
	preinstall,
	install,
	username,
	password,
	postinstall,
	windowsKey,
	append,
	mirror,
	kernel,
	initrd,
	status string,
	permitpreinstall,
	permitinstall,
	allowpostinstall,
	announce bool) (*Host, error) {
	return &Host{
		Name:               name,
		StorePath:          storePath,
		Macaddress:         mac,
		Preinstall:         preinstall,
		PermitPreinstall: permitpreinstall,
		Install:            install,
		PermitInstall:    permitinstall,
		Username:           username,
		Password:           password,
		Postinstall:        postinstall,
		PermitPostinstall:   allowpostinstall,
		WindowsKey:         windowsKey,
		Append:             append,
		Mirror:             mirror,
		Kernel:             kernel,
		Initrd:             initrd,
		Status:             status,
		Announce:           announce,
	}, nil
}

/*func (h *Host) Create() error {
	if err := h.SaveConfig(); err != nil {
		return err
	}
	return nil
}*/

func (h *Host) removeStorePath() error {
	file, err := os.Stat(h.StorePath)
	if err != nil {
		return err
	}
	if !file.IsDir() {
		return fmt.Errorf("%q is not a directory", h.StorePath)
	}
	return os.RemoveAll(h.StorePath)
}

func (h *Host) SaveConfig() error {
	data, err := json.Marshal(h)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(h.StorePath, "config.json"), data, 0600); err != nil {
		return err
	}
	return nil
}

func LoadHost(name string) (*Host, error) {
	hostPath := filepath.Join(hostDir(), name)
	if _, err := os.Stat(hostPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Host %q does not exist", name)
	}
	host := &Host{Name: name, StorePath: hostPath}
	if err := host.LoadConfig(); err != nil {
		return nil, err
	}
	return host, nil
}

func (h *Host) LoadConfig() error {
	data, err := ioutil.ReadFile(filepath.Join(h.StorePath, "config.json"))
	if err != nil {
		return err
	}
	/*
		// First pass: find the driver name and load the driver
		var config hostConfig
		if err := json.Unmarshal(data, &config); err != nil {
			return err
		}

		driver, err := drivers.NewDriver(config.DriverName, h.storePath)
		if err != nil {
			return err
		}
		h.Driver = driver
	*/
	// Second pass: unmarshal driver config into correct driver
	if err := json.Unmarshal(data, &h); err != nil {
		return err
	}

	return nil

}
