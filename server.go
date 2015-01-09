package main

import (
	"path/filepath"
	"html/template"
	"net/http"
	"os"
	"fmt"
	//"io/ioutil"
	"github.com/gorilla/mux"
	log "github.com/Sirupsen/logrus"
	"github.com/pandrew/stasis/drivers"

)

func GetStasisDir() string {
	return fmt.Sprintf(filepath.Join(drivers.GetHomeDir(), ".stasis"))
}

func GetIpxeDir() string {
	return filepath.Join(drivers.GetHomeDir(), ".stasis", "ipxe")
}

func IpxeDirExists() (bool, error) {
	_, err := os.Stat(GetIpxeDir())
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}


func initRouter() {
	r := mux.NewRouter()
	r.HandleFunc("/{id}", ReturnIpxe)
	http.Handle("/", r)


	port := os.Getenv("STASIS_HTTP_PORT")
	log.Info("Listening on: ", port)
	path := os.Getenv("STASIS_STORAGE_PATH")
	log.Info("Using path: ", path)



	r.PathPrefix("/").Handler(http.FileServer(http.Dir(path)))

	log.Println("Listening...")
	http.ListenAndServe(":" + os.Getenv("STASIS_HTTP_PORT"), nil)
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
		renderTemplate(w, host.Template, host)
		host.Status = "INSTALLED"
		host.SaveConfig()
	} else {
		http.NotFound(w, r)
		return

	}
}

var templates *template.Template

func init() {
	filenames := []string{}

	dirIpxe := GetIpxeDir()

	err := filepath.Walk(dirIpxe, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".ipxe" {
			filenames = append(filenames, path)
		}
		return nil
	})

	if err != nil {
		log.Fatalln(err)
	}

	if len(filenames) == 0 {
		log.Errorf("There is no ipxe templates in: %q", dirIpxe )
		os.Exit(1)
	}
		
	templates, err = template.ParseFiles(filenames...)
	if err != nil {
		log.Fatalln(err)
	}
	

}

func renderTemplate(w http.ResponseWriter, tmpl string, vars interface{}) {
        err := templates.ExecuteTemplate(w, tmpl+".ipxe", vars)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
        }
}