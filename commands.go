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

	"github.com/pandrew/stasis/drivers"
	_ "github.com/pandrew/stasis/drivers/none"
)

type hostListItem struct {
	Name       string
	Active     bool
	Status     string
	DriverName string
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
		log.Errorf("error determining whether host %q is active: %s",
			host.Name, err)
	}

	hostListItems <- hostListItem{
		Name:       host.Name,
		Active:     isActive,
		DriverName: host.Driver.DriverName(),
		Status:		host.Status,
	}
}

var Commands = []cli.Command{
	{
		Flags: append(
			drivers.GetCreateFlags(),
		cli.StringFlag{
			Name: "driver, d",
			Usage: fmt.Sprintf(
				"Driver to create machine with. Available drivers: %s",
				strings.Join(drivers.GetDriverNames(), ", "),
			),
			Value: "none",
		},
  		cli.StringFlag{
    		Name: "mac",
    		Value: "",
    		Usage: "Mac address to use, Example: 00-00-00-00-00-00",
  		},
  		cli.StringFlag{
    		Name: "template",
    		Value: "",
    		Usage: "iPxe template",
  		},
  		cli.StringFlag{
    		Name: "append",
    		Value: "",
    		Usage: "Append string",
  		},
  		cli.StringFlag{
    		Name: "mirror",
    		Value: "localhost",
    		Usage: "Location for static content",
    		EnvVar: "STASIS_HTTP_MIRROR",
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
    		Name: "status",
    		Value: "INACTIVE",
    		Usage: "Initial status of machine",
  		},
  	),
		Name:  "create",
		Usage: "Create a machine",
		Action: func(c *cli.Context) {
			driver := c.String("driver")
			mac := c.String("mac")
			template := c.String("template")
			append := c.String("append")
			mirror := c.String("mirror")
			os.Setenv("STASIS_HTTP_MIRROR", mirror)
			kernel := c.String("kernel")
			initrd := c.String("initrd")
			status := c.String("status")


			name := c.Args().First()

			if name == "" {
				cli.ShowCommandHelp(c, "create")
				os.Exit(1)
			}

			ValidateHostName(name)

			if mac == "" {
				cli.ShowCommandHelp(c, "create")
				os.Exit(1)
			}

			ValidateMacaddr(mac)

			if template == "" {
				cli.ShowCommandHelp(c, "create")
				os.Exit(1)
			}

			store := NewStore(c.GlobalString("storage-path"))


			host, err := store.Create(name, driver, mac, template, append, mirror, kernel, initrd, status, c)
			if err != nil {
				log.Fatal(err)
			}
			if err := store.SetActive(host); err != nil {
				log.Fatalf("error setting active host: %v", err)
			}

			log.Infof("%q has been created and is now the active machine. To point Docker at this machine, run: export DOCKER_HOST=$(machine url) DOCKER_AUTH=identity", name)
		},
	},
	{
		Name:  "inspect",
		Usage: "Inspect information about a machine",
		Action: func(c *cli.Context) {
			prettyJSON, err := json.MarshalIndent(getHost(c), "", "    ")
			if err != nil {
				log.Fatal(err)
			}
			log.Println(getHost(c))
			fmt.Println(string(prettyJSON))
		},
	},
	{
		Name:  "toggle",
		Usage: "Toggles hosts status between INACTIVE and ACTIVE ",
		Action: func(c *cli.Context) {
			fmt.Println(GetIpxeDir())

			host := getHost(c)
			
			if host.Status == "INACTIVE" {
				host.Status = "ACTIVE"
			} else if host.Status == "INSTALLED" {
				host.Status = "INACTIVE"
			} else {
				host.Status = "INACTIVE"
			}

			host.SaveConfig()

		},
	},
	{
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "quiet, q",
				Usage: "Enable quiet mode",
			},
		},
		Name:  "ls",
		Usage: "List machines",
		Action: func(c *cli.Context) {
			quiet := c.Bool("quiet")
			store := NewStore(c.GlobalString("storage-path"))

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

			if !quiet {
				for i := 0; i < len(hostList); i++ {
					items = append(items, <-hostListItems)
				}
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
		},
	},
	{
		Flags: []cli.Flag {
  		cli.StringFlag{
    		Name: "port",
    		//needs input validation
    		Value: "8080",
    		Usage: "default port to listen on",
    		EnvVar: "STASIS_HTTP_PORT",
  		},
  	},
		Name: "listen",
		Usage: "Listens on port",
		Action: func(c *cli.Context) {
			//store := NewStore()
			os.Setenv("STASIS_HTTP_PORT", c.String("port"))

			store := NewStore(c.GlobalString("storage-path"))
			_, err := os.Stat(store.Path)
			if os.IsNotExist(err) {
				log.Errorf("There is no machines or location to store them.")
				cli.ShowCommandHelp(c, "create")
				os.Exit(1)	
			} else if err == nil {
				initRouter()
			
			}
		},
	},
}


