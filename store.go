package main

import (
	"os"
	"fmt"
	"path/filepath"
	"io/ioutil"

	"github.com/pandrew/stasis/drivers"
	log "github.com/Sirupsen/logrus"

)



// Store persists hosts on the filesystem
type Store struct {
	Path string
}

func NewHostStore(rootPath string) *Store {
	if rootPath == "" {
		rootPath = filepath.Join(drivers.GetHomeDir(), ".stasis", "machines")
		os.Setenv("STASIS_HOST_STORAGE_PATH", rootPath)
	}
	return &Store{Path: rootPath}
}

func (s *Store) CreateHost(
	name,
	driverName,
	mac,
	preinstall,
	install,
	windowsKey,
	append,
	mirror,
	kernel, 
	initrd, 
	status string, 
	flags drivers.DriverOptions) (*Host, error) {
	exists, err := s.Exists(name)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, fmt.Errorf("Host %q already exists", name)
	}

	hostPath := filepath.Join(s.Path, name)

	host, err := NewHost(
		name, 
		driverName, 
		mac, 
		preinstall, 
		install,
		windowsKey,
		append, 
		mirror, 
		kernel, 
		initrd, 
		status, 
		hostPath)
	if err != nil {
		return host, err
	}
	if flags != nil {
		if err := host.Driver.SetConfigFromFlags(flags); err != nil {
			return host, err
		}
	}

	if err := os.MkdirAll(hostPath, 0700); err != nil {
		return nil, err
	}

	if err := host.SaveConfig(); err != nil {
		return host, err
	}

	if err := host.Create(); err != nil {
		return host, err
	}
	return host, nil
}

func (s *Store) GetMacaddress(macaddress string) (*Host, error) {
	dir, err := ioutil.ReadDir(s.Path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	host := &Host{}

	for _, file := range dir {
		if file.IsDir() {
			host, err := s.Load(file.Name())
			if host.Macaddress == macaddress {
				return host, nil
			} else if err != nil {
				log.Errorf("error loading host %q: %s", file.Name(), err)
				continue
			}
		}
	}
	return host, nil
}

func (s *Store) GetHostname(hostname string) (*Host, error) {
	_, err := ioutil.ReadDir(s.Path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	host, err := s.Load(hostname)
	if err != nil {
		//log.Errorf("error loading host %q: %s", host, err)
		log.Println(host)
	}

	return host, nil
}

func (s *Store) Exists(name string) (bool, error) {
	_, err := os.Stat(filepath.Join(s.Path, name))
	if os.IsNotExist(err) {
		return false, nil
	} else if err == nil {
		return true, nil
	}
	return false, err
}

func (s *Store) List() ([]Host, error) {
	dir, err := ioutil.ReadDir(s.Path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	hosts := []Host{}

	for _, file := range dir {
		if file.IsDir() {
			host, err := s.Load(file.Name())
			if err != nil {
				log.Errorf("error loading host %q: %s", file.Name(), err)
				continue
			}
			hosts = append(hosts, *host)
		}
	}
	return hosts, nil
}

func (s *Store) Load(name string) (*Host, error) {
	hostPath := filepath.Join(s.Path, name)
	return LoadHost(name, hostPath)
}

func (s *Store) IsActive(host *Host) (bool, error) {
	active, err := s.GetActive()
	if err != nil {
		return false, err
	}
	if active == nil {
		return false, nil
	}
	return active.Name == host.Name, nil
}

func (s *Store) GetActive() (*Host, error) {
	hostName, err := ioutil.ReadFile(s.activePath())
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return s.Load(string(hostName))
}

func (s *Store) SetActive(host *Host) error {
	if err := os.MkdirAll(filepath.Dir(s.activePath()), 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(s.activePath(), []byte(host.Name), 0600)
}

func (s *Store) RemoveActive() error {
	return os.Remove(s.activePath())
}

// activePath returns the path to the file that stores the name of the
// active host
func (s *Store) activePath() string {
	return filepath.Join(s.Path, ".active")
}