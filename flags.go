package main

import (
	"github.com/codegangsta/cli"
)

var (
	install = cli.StringFlag{
		Name: "install",
		Value: "",
		Usage: "kickstart/preinstall template",
	}
	installUsername = cli.StringFlag{
		Name: "username",
		Value: "vagrant",
		Usage: "username",
	}
	installPassword = cli.StringFlag{
		Name: "password",
		Value: "vagrant",
		Usage: "username",
	}
	installWindowsKey = cli.StringFlag{
		Name: "windows-key",
		Value: "",
		Usage: "key",
	}
	
)


var (
  	preinstall = cli.StringFlag{
		Name: "preinstall",
		Value: "",
		Usage: "iPxe template",
	}
	preinstallMac = cli.StringFlag{
		Name: "mac",
		Value: "",
		Usage: "Mac address to use, Example: 00-00-00-00-00-00",
	}
	preinstallAppend = cli.StringFlag{
		Name: "append",
		Value: "",
		Usage: "Append string",
	}

	preinstallKernel = cli.StringFlag{
		Name: "kernel",
		Value: "",
		Usage: "Kernel string",
	}
	preinstallInitrd = cli.StringFlag{
		Name: "initrd",
		Value: "",
		Usage: "Initrd string",
	}
)

var (
	postinstall = cli.StringFlag{
		Name: "postinstall",
		Value: "",
		Usage: "Uri to script to execute after installation.",
	}
	
)