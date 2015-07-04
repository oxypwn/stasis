package main

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Store persists hosts on the filesystem
type Store struct {
	Path string
}

func NewHostStore(rootPath string) *Store {
	if rootPath == "" {
		rootPath = filepath.Join(GetHomeDir(), ".stasis", "machines")
		os.Setenv("STASIS_HOST_STORAGE_PATH", rootPath)
	}
	return &Store{Path: rootPath}
}

func (s *Store) Save(host *Host) error {
	data, err := json.Marshal(host)
	if err != nil {
		return err
	}

	hostPath := filepath.Join(GetStasisDir(), host.Name)

	if err := os.MkdirAll(hostPath, 0700); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(hostPath, "config.json"), data, 0600); err != nil {
		return err
	}

	return nil
}

func (s *Store) CreateHost(name, storePath, mac, preinstall, install, username, password, postinstall, windowsKey, append, mirror, kernel, initrd, status string, disabledpreinstall, allowinstall, allowpostinstall, announce bool) (*Host, error) {
	exists, err := s.Exists(name)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, fmt.Errorf("Host %q already exists", name)
	}

	hostPath := filepath.Join(s.Path, name)

	host, err := NewHost(name, storePath, mac, preinstall, install, username, password, postinstall, windowsKey, append, mirror, kernel, initrd, status, disabledpreinstall, allowinstall, allowpostinstall, announce)
	if err != nil {
		return host, err
	}

	if err := os.MkdirAll(hostPath, 0700); err != nil {
		return nil, err
	}

	if err := host.SaveConfig(); err != nil {
		return host, err
	}

	//if err := host.Create(); err != nil {
	//	return host, err
	//}
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
			isActive, err := s.IsActive(host)
			if err != nil {
				log.Debugf("error determining whether host %q is active: %s",
					host.Name, err)
			}
			if isActive {
				if host.Macaddress == macaddress {
					return host, nil
				} else if err != nil {
					log.Errorf("error loading host %q: %s", file.Name(), err)
					continue
				}
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
	return LoadHost(name)
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
		return nil, err
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

func (s *Store) RemoveActive(name string) error {
	return os.Remove(s.activePath())
}

func (s *Store) Remove(name string) error {
	_, err := LoadHost(name)
	if err != nil {
		return err
	}
	hostPath := filepath.Join(hostDir(), name)
	return os.RemoveAll(hostPath)

}

// activePath returns the path to the file that stores the name of the
// active host
func (s *Store) activePath() string {
	return filepath.Join(s.Path, ".active")
}
