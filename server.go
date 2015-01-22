package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net"
	"os"
	"path/filepath"
	"io/ioutil"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/pandrew/stasis/drivers"
)

const (
	 extPreinstall string = ".preinstall"
	 extGohtml string = ".gohtml"
	 extInstall string = ".install"
)

func GetStasisDir() string {
	return fmt.Sprintf(filepath.Join(drivers.GetHomeDir(), ".stasis"))
}

func preinstallDir() string {
	return filepath.Join(GetStasisDir(), "preinstall")
}

func gohtmlDir() string {
	return filepath.Join(drivers.GetHomeDir(), ".stasis", "gohtml")
}

func installDir() string {
	return filepath.Join(GetStasisDir(), "install")
}

func postinstallDir() string {
	return filepath.Join(drivers.GetHomeDir(), ".stasis", "postinstall")
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

func init() {
	dirInstall := installDir()
	pathExist, _ := DirExists(dirInstall) 
	if !pathExist {
		if err := os.MkdirAll(dirInstall, 0700); err != nil {
			log.Println(err)
		}
	}


		
}



func initRouter(gather bool) {
	r := mux.NewRouter()
	// Prepend uri with v1 for version 1 api. This will help error responds
	// when using relative paths in links.
	r.HandleFunc("/v1/{id}/preinstall", ReturnPreinstall)
	r.HandleFunc("/v1/{id}/preinstall/raw", ReturnRawPreinstall)
	r.HandleFunc("/v1/{id}/preinstall/preview", ReturnPreviewPreinstall)
	r.HandleFunc("/v1/{id}/install", ReturnInstall)
	r.HandleFunc("/v1/{id}/install/raw", ReturnRawInstall)
	r.HandleFunc("/v1/info/stats", ReturnStats)
	r.HandleFunc("/v1/{id}/toggle", ReturnToggle)
	if gather {
		r.HandleFunc("/v1/{id}/announce", GatherMac)
	}
	http.Handle("/", r)

	port := os.Getenv("STASIS_HTTP_PORT")
	log.Info("Listening on: ", port)
	path := os.Getenv("STASIS_HOST_STORAGE_PATH")
	log.Info("Using path: ", path)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir(path)))

	log.Println("Listening...")
	http.ListenAndServe(":"+os.Getenv("STASIS_HTTP_PORT"), nil)
}

func ReturnStats(w http.ResponseWriter, r *http.Request) {
	store := NewHostStore(os.Getenv("STASIS_HOST_STORAGE_PATH"))

	hostList, err := store.List()
	if err != nil {
		log.Fatal(err)
	}

	items := []hostListItem{}
	hostListItems := make(chan hostListItem)

	for _, host := range hostList {
		go getHostState(host, *store, hostListItems)
	}

	for i := 0; i < len(hostList); i++ {
		items = append(items, <-hostListItems)
	}

	close(hostListItems)
	templates, err := template.New("stats").Parse(index)
	if err != nil {
        panic(err)
    }
    err = templates.Execute(w, items)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}


func ReturnInstall(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	macaddress := vars["id"]

	_, err := ValidateMacaddr(macaddress)
	if err != nil {
		http.NotFound(w, r)
	} else {

		store := NewHostStore(os.Getenv("STASIS_HOST_STORAGE_PATH"))
		host, err := store.GetMacaddress(macaddress)
		if err != nil {
			log.Fatal(err)
		}
		//inst := installDir()
		//ValidateTemplates(inst, extInstall)
		//test := host.Install
		if len(host.Install) != 0 {
			tmpl := host.Install + extInstall
			renderTemplate(w, tmpl, host)
		} else {
			http.NotFound(w, r)

		}

	}

}

func ReturnRawInstall(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	macaddress := vars["id"]

	_, err := ValidateMacaddr(macaddress)
	if err != nil {
		http.NotFound(w, r)
	} else {

		store := NewHostStore(os.Getenv("STASIS_HOST_STORAGE_PATH"))
		host, err := store.GetMacaddress(macaddress)
		if err != nil {
			log.Fatal(err)
		}
		if len(host.Install) != 0 {
			dir := installDir()
			returnRaw(w, dir, host.Install, extInstall)
		} else {
			http.NotFound(w, r)
		}
	}

}

func ReturnPreinstall(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	macaddress := vars["id"]
	if macaddress == "" {
		http.NotFound(w, r)
		return
	}

	store := NewHostStore(os.Getenv("STASIS_HOST_STORAGE_PATH"))
	host, err := store.GetMacaddress(macaddress)
	if err != nil {
		log.Fatal(err)
	}

	if host.Status == "ACTIVE" {
		pre := preinstallDir()
		ValidateTemplates(pre, extPreinstall)
		tmpl := host.Preinstall + extPreinstall
		renderTemplate(w, tmpl, host)
		host.Status = "INSTALLED"
		host.SaveConfig()
	} else if host.Status == "INSTALLED" {
		ip := GetIP(r)
		log.Errorf("%s requests %s: host is already installed!", ip, macaddress)
	} else {		
		ip := GetIP(r)
		log.Errorf("%s requests %s: not in database!", ip, macaddress)

	}
}

func ReturnRawPreinstall(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	macaddress := vars["id"]

	_, err := ValidateMacaddr(macaddress)
	if err != nil {
		http.NotFound(w, r)
	} else {

		store := NewHostStore(os.Getenv("STASIS_HOST_STORAGE_PATH"))
		host, err := store.GetMacaddress(macaddress)
		if err != nil {
			log.Println(err)
		}

		dir := preinstallDir()
		returnRaw(w, dir, host.Preinstall, extPreinstall)
	}
}

func ReturnPreviewPreinstall(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	macaddress := vars["id"]

	_, err := ValidateMacaddr(macaddress)
	if err != nil {
		http.NotFound(w, r)
	} else {

		store := NewHostStore(os.Getenv("STASIS_HOST_STORAGE_PATH"))
		host, err := store.GetMacaddress(macaddress)
		if err != nil {
			log.Println(err)
		}
		pre := preinstallDir()
		ValidateTemplates(pre, extPreinstall)
		tmpl := host.Preinstall + extPreinstall
		renderTemplate(w, tmpl, host)
	}
}

func ReturnToggle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	macaddress := vars["id"]
	store := NewHostStore(os.Getenv("STASIS_HOST_STORAGE_PATH"))
	host, err := store.GetMacaddress(macaddress)
	if err != nil {
		log.Println(err)
	}
	log.Println(host)
	if host.Status == "INACTIVE" {
		host.Status = "ACTIVE"
		log.Infof("%s is now ACTIVE", host.Name)
	} else if host.Status == "INSTALLED" {
		host.Status = "ACTIVE"
		log.Infof("%s is now ACTIVE", host.Name)
	} else {
		host.Status = "INACTIVE"
		log.Infof("%s is now INATIVE", host.Name)
	}

	host.SaveConfig()
	http.Redirect(w, r, "/v1/info/stats", http.StatusFound)

}

func GatherMac(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	macaddress := vars["id"]
	if macaddress == "" {
		http.NotFound(w, r)
		return
	}

	ValidateMacaddr(macaddress)
	
	store := NewHostStore(os.Getenv("STASIS_HOST_STORAGE_PATH"))

	host, err := store.GetActive()
	if err != nil {
		log.Println(err)
	}

	ip := GetIP(r)
	
	if macaddress == host.Macaddress {
		http.NotFound(w, r)
		log.Errorf("%s requests to modify %q with macaddress %s to %s: DENIED" , ip, host.Name, host.Macaddress, macaddress)
		return
	} else {	
		log.Printf("%s requests to modify %q with macaddress %s to %s: ACCEPTED" , ip, host.Name, host.Macaddress, macaddress)
	}

	host.Macaddress = macaddress
	host.SaveConfig()
}

var templates *template.Template


func renderTemplate(w http.ResponseWriter, tmpl string, vars interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func returnRaw(w http.ResponseWriter, dir string, tmpl string, ext string) {
	raw, err := ioutil.ReadFile(dir + "/" + tmpl + ext)
	if err != nil {
		log.Println(err)
	}
	fmt.Fprintf(w, string(raw))
}

func GetIP(r *http.Request) string {
    if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
        return ipProxy
    }
    ip, _, _ := net.SplitHostPort(r.RemoteAddr)
    return ip
}
