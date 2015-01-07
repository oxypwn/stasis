package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	log "github.com/Sirupsen/logrus"
	"github.com/pandrew/stasis/drivers"

)


type Host struct {
	Name       string `json:"-"`
	DriverName string
	Driver     drivers.Driver
	Macaddress string
	Template    string
	Append		string
	Mirror		string
	Kernel		string
	Initrd		string
	Status		string 
	storePath  string
}

type hostConfig struct {
	DriverName string
}

func NewHost(name, driverName, mac, template, append, mirror, kernel, initrd, status, storePath string) (*Host, error) {
	driver, err := drivers.NewDriver(driverName, storePath)
	if err != nil {
		return nil, err
	}
	//status = "INACTIVE"

	return &Host{
		Name:       name,
		DriverName:	driverName,
		Driver:		driver,
		Macaddress:	mac,
		Template:	template,
		Append:		append,
		Mirror:		mirror,
		Kernel:		kernel,
		Initrd:		initrd,
		Status:		status,
		storePath:  storePath,
	}, nil
}

func (h *Host) Create() error {
	if err := h.Driver.Create(); err != nil {
		return err
	}
	if err := h.SaveConfig(); err != nil {
		return err
	}
	return nil
}


func (h *Host) removeStorePath() error {
	file, err := os.Stat(h.storePath)
	if err != nil {
		return err
	}
	if !file.IsDir() {
		return fmt.Errorf("%q is not a directory", h.storePath)
	}
	return os.RemoveAll(h.storePath)
}




func (h *Host) SaveConfig() error {
	data, err := json.Marshal(h)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(h.storePath, "config.json"), data, 0600); err != nil {
		return err
	}
	return nil
}

func LoadHost(name string, storePath string) (*Host, error) {
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Host %q does not exist", name)
	}
	host := &Host{Name: name, storePath: storePath}
	if err := host.LoadConfig(); err != nil {
		return nil, err
	}
	return host, nil
}

func (h *Host) LoadConfig() error {
	data, err := ioutil.ReadFile(filepath.Join(h.storePath, "config.json"))
	if err != nil {
		return err
	}

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

	// Second pass: unmarshal driver config into correct driver
	if err := json.Unmarshal(data, &h); err != nil {
		return err
	}

	return nil

}

func getHost(c *cli.Context) *Host {
	name := c.Args().First()
	store := NewStore(c.GlobalString("storage-path"))

	if name == "" {
		host, err := store.GetActive()
		if err != nil {
			log.Fatalf("unable to get active host: %v", err)
		}
		return host
	}

	host, err := store.Load(name)
	if err != nil {
		log.Fatalf("unable to load host: %v", err)
	}
	return host
}