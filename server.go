package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net"
	"os"
	"path/filepath"
	//"io/ioutil"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/pandrew/stasis/drivers"
)

const (
	 extIpxe string = ".ipxe" 
)

func GetStasisDir() string {
	return fmt.Sprintf(filepath.Join(drivers.GetHomeDir(), ".stasis"))
}

func ipxeDir() string {
	return filepath.Join(drivers.GetHomeDir(), ".stasis", "ipxe")
}

func DirExists(dir string) (bool, error) {
	_, err := os.Stat(dir)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func installDir() string {
	return filepath.Join(drivers.GetHomeDir(), ".stasis", "install")
}

func postinstallDir() string {
	return filepath.Join(drivers.GetHomeDir(), ".stasis", "postinstall")
}


func initRouter(gather bool) {
	r := mux.NewRouter()
	r.HandleFunc("/{id}", ReturnIpxe)
	if gather {
		r.HandleFunc("/{id}/gather", GatherMac)
	}
	http.Handle("/", r)

	port := os.Getenv("STASIS_HTTP_PORT")
	log.Info("Listening on: ", port)
	path := os.Getenv("STASIS_STORAGE_PATH")
	log.Info("Using path: ", path)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir(path)))

	log.Println("Listening...")
	http.ListenAndServe(":"+os.Getenv("STASIS_HTTP_PORT"), nil)
}

func ReturnIpxe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	macaddress := vars["id"]
	if macaddress == "" {
		http.NotFound(w, r)
		return
	}

	store := NewStore(os.Getenv("STASIS_STORAGE_PATH"))
	host, err := store.GetMacaddress(macaddress)
	if err != nil {
		log.Fatal(err)
	}

	if host.Status == "ACTIVE" {
		renderTemplate(w, host.Template, extIpxe, host)
		host.Status = "INSTALLED"
		host.SaveConfig()
	} else {
		http.NotFound(w, r)
		return

	}
}


func GatherMac(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	macaddress := vars["id"]
	if macaddress == "" {
		http.NotFound(w, r)
		return
	}

	ValidateMacaddr(macaddress)
	
	store := NewStore(os.Getenv("STASIS_STORAGE_PATH"))

	host, err := store.GetActive()
	if err != nil {
		log.Println(err)
	}

	ip := GetIP(r)
	
	if macaddress == host.Macaddress {
		http.NotFound(w, r)
		log.Errorf("Request from %s to modify %q with macaddress %s to %s DENIED" , ip, host.Name, host.Macaddress, macaddress)
		return
	} else {	
		log.Printf("Request from %s to modify %q with macaddress %s to %s ACCEPTED" , ip, host.Name, host.Macaddress, macaddress)
	}

	host.Macaddress = macaddress
	host.SaveConfig()
}

var templates *template.Template

func init() {
	//filenames := []string{}

	//store := NewStore(os.Getenv("STASIS_STORAGE_PATH"))

	dirIpxe := ipxeDir()
	if err := os.MkdirAll(dirIpxe, 0700); err != nil {
		log.Println(err)
	}
	dirInstall := installDir()
	if err := os.MkdirAll(dirInstall, 0700); err != nil {
		log.Println(err)
	}
	dirPostinstall := postinstallDir()
	if err := os.MkdirAll(dirPostinstall, 0700); err != nil {
		log.Println(err)
	}

	ValidateTemplates(dirIpxe, extIpxe)

}

func renderTemplate(w http.ResponseWriter, tmpl string, ext string, vars interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+ext, vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetIP(r *http.Request) string {
    if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
        return ipProxy
    }
    ip, _, _ := net.SplitHostPort(r.RemoteAddr)
    return ip
}
