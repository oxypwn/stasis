package main

import (
	"fmt"
	"strings"
	"os"
	//"sync"
	"sort"
	"encoding/json"
	"text/tabwriter"
	"github.com/codegangsta/cli"
	log "github.com/Sirupsen/logrus"

	//"github.com/pandrew/stasis/drivers"
	//_ "github.com/pandrew/stasis/drivers/none"
)


type hostListItem struct {
	Name       string
	Active     bool
	Preinstall   string
	Install    string
	Postinstall string
	Status     string
	Append     string
	DriverName string
	Macaddress string
}

type hostListItemByName []hostListItem

func (h hostListItemByName) Len() int {
	return len(h)
}

func (h hostListItemByName) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h hostListItemByName) Less(i, j int) bool {
	return strings.ToLower(h[i].Name) < strings.ToLower(h[j].Name)
}

func getHostState(host Host, store Store, hostListItems chan<- hostListItem) {
	isActive, err := store.IsActive(&host)
	if err != nil {
		log.Debugf("error determining whether host %q is active: %s",
			host.Name, err)
	}

	hostListItems <- hostListItem{
		Name:       host.Name,
		Active:     isActive,
		Preinstall:	host.Preinstall,
		Install:	host.Install,
		//DriverName: host.Driver.DriverName(),
		Status:		host.Status,
		Macaddress: host.Macaddress,
	}
}

var Flags = []cli.Flag {
  cli.StringFlag{
    Name: "lang, l",
    Value: "english",
    Usage: "language for the greeting",
    EnvVar: "LEGACY_COMPAT_LANG,APP_LANG,LANG",
  },
}

var Commands = []cli.Command{
	{
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "quiet, q",
				Usage: "Enable quiet mode",
			},
		},
		Name:  "list",
		ShortName: "ls",
		Usage: "List machines",
		Action: cmdLs,
	},
	{
		Flags: []cli.Flag {
			cli.StringFlag{
				Name: "preinstall",
				Value: "",
				Usage: "iPxe template",
			},
			cli.StringFlag{
				Name: "mac",
				Value: "",
				Usage: "Mac address of host, Example: 00-00-00-00-00-00",
			},
			cli.StringFlag{
				Name: "append",
				Value: "",
				Usage: "Append string",
			},
			cli.StringFlag{
				Name: "kernel",
				Value: "",
				Usage: "Kernel string",
			},
			cli.StringFlag{
				Name: "initrd",
				Value: "",
				Usage: "Initrd string",
			},
			cli.StringFlag{
				Name: "install",
				Value: "",
				Usage: "kickstart/preseed/Autounattend.xml/... template",
			},
			cli.StringFlag{
				Name: "serial",
				Value: "",
				Usage: "Serial key; Windows...",
			},
			cli.StringFlag{
				Name: "username",
				Value: "stasis",
				Usage: "Username for default user",
			},
			cli.StringFlag{
				Name: "password",
				Value: "stasis",
				Usage: "Password to default user",
			},
			cli.StringFlag{
				Name: "postinstall",
				Value: "",
				Usage: "Uri to script to execute after installation.",
			},
		},
		Name: "create",
		ShortName: "c",
		Usage: "Create host installation profile",
		Action: cmdCreateHost,
	},
	{
		Flags: []cli.Flag {
  		cli.StringFlag{
    		Name: "port",
    		Value: "8080",
    		Usage: "default port to listen on",
    		EnvVar: "STASIS_HTTP_PORT",
  		},
  		cli.StringFlag{
    		Name: "static",
    		Value: staticDir(),
    		Usage: "default path for static content",
    		EnvVar: "STASIS_HTTP_STATIC_PATH",
  		},
  		cli.BoolFlag{
    		Name: "gather, g",
    		Usage: "Gather mac address",
  		},
  	},
		Name: "listen",
		ShortName: "l",
		Usage: "Listens on port",
		Action: cmdListen,
	},
}


func cmdNotFound(c *cli.Context, command string) {
	log.Fatalf(
		"%s: '%s' is not a %s command. See '%s --help'.",
		c.App.Name,
		command,
		c.App.Name,
		c.App.Name,
	)
}

func cmdInspect(c *cli.Context) {
	prettyJSON, err := json.MarshalIndent(getHost(c), "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(getHost(c))
	fmt.Println(string(prettyJSON))
}

func cmdCreateHost(c *cli.Context) {
	//driver := c.String("driver")
	mac := c.String("mac")
	preinstall := c.String("preinstall")
	install := c.String("install")
	username := c.String("username")
	password := c.String("password")
	postinstall := c.String("postinstall")
	windowsKey := c.String("windows-key")
	append := c.String("append")
	mirror := c.String("mirror")
	os.Setenv("STASIS_HTTP_MIRROR", mirror)
	kernel := c.String("kernel")
	initrd := c.String("initrd")
	status := c.String("status")

	// check for missing settings
	if preinstall == "" {
		log.Errorf("Missing required option --preinstall")
		os.Exit(1)
	}

	name := c.Args().First()

	if name == "" {
		cli.ShowCommandHelp(c, "create")
		os.Exit(1)
	}

	match := ValidateHostName(name)
	if match == false {
		log.Errorf("%q Is not a valid hostname.", name)
		cli.ShowCommandHelp(c, "create")
		os.Exit(1)
	}

	if mac != "" {
		ValidateMacaddr(mac)
	}

	announce := false

	store := NewHostStore(c.GlobalString("storage-path"))


	host, err := store.CreateHost(name, mac, preinstall, install, username, password, postinstall, windowsKey, append, mirror, kernel, initrd, status, announce)
	if err != nil {
		log.Fatal(err)
	}
	if err := store.SetActive(host); err != nil {
		log.Fatalf("error setting active host: %v", err)
	}

	log.Infof("%q has been created and is now the active machine. To point Docker at this machine, run: export DOCKER_HOST=$(machine url) DOCKER_AUTH=identity", name)
}

func cmdToggle(c *cli.Context) {

	host := getHost(c)

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

}
/*
var global string
func cmdGather(c *cli.Context) {
		host := getHost(c)
		global := host.Name
		log.Println("global", global)
		initRouter()
}
*/

func cmdLs(c *cli.Context) {
	quiet := c.Bool("quiet")
	store := NewHostStore(c.GlobalString("storage-path"))

	hostList, err := store.List()
	if err != nil {
		log.Fatal(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)

	if !quiet {
		fmt.Fprintln(w, "NAME\tACTIVE\tDRIVER\tSTATUS")
	}

	items := []hostListItem{}
	hostListItems := make(chan hostListItem)

	for _, host := range hostList {
		if !quiet {
			go getHostState(host, *store, hostListItems)
		} else {
			fmt.Fprintf(w, "%s\n", host.Name)
		}
	}

	for i := 0; i < len(hostList); i++ {
		items = append(items, <-hostListItems)
	}

	close(hostListItems)

	sort.Sort(hostListItemByName(items))

	for _, item := range items {
		activeString := ""
		if item.Active {
			activeString = "*"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			item.Name, activeString, item.DriverName, item.Status)
	}

	w.Flush()
}

func cmdListen(c *cli.Context) {
	gather := c.Bool("gather")
	os.Setenv("STASIS_HTTP_STATIC_PATH", c.String("static"))
	os.Setenv("STASIS_HTTP_PORT", c.String("port"))
	store := NewHostStore(c.GlobalString("storage-path"))
	_, err := os.Stat(store.Path)
	if os.IsNotExist(err) {
		log.Errorf("There is no machines or location to store them.")
		cli.ShowCommandHelp(c, "H c")
		os.Exit(1)
	} else if err == nil {
		if gather {
			name := c.Args().First()
			if name == "" {
				_, err := store.GetActive()
				if err != nil {
					log.Fatalf("unable to get active host: %v", err)
				}
			} else {
				host, err := store.Load(name)
				if err != nil {
					log.Fatalf("error loading host: %v", err)
				}

				if err := store.SetActive(host); err != nil {
					log.Fatalf("error setting active host: %v", err)
				}
			}
		}
		initRouter()

	}
}

func cmdListTemplates(c *cli.Context) {
	preinstall := preinstallDir()
	install := installDir()
	var paths []string
	paths = append(paths, preinstall, install)

	for _, path := range paths {
		listTemplates(path)
	}

}
func getHost(c *cli.Context) *Host {
	name := c.Args().First()
	store := NewHostStore(c.GlobalString("storage-path"))

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
