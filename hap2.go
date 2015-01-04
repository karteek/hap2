package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/codegangsta/cli"
	"os"
)

type site struct {
	User   string `json:"user"`
	Salt   string `json:"salt"`
	Domain string `json:"domain"`
	Length int    `json:"length"`
	Suffix string `json:"suffix"`
	Notes  string `json:"notes"`
	// check  string `json:"check"`
}

type siteCollection struct {
	list map[string]site
}

var SiteListFile string
var sitelist siteCollection

func main() {
	SiteListFile = fmt.Sprintf("%s/.config/hap2.sitelist.json", os.Getenv("HOME"))
	app := cli.NewApp()
	app.Name = "hap2"
	app.Usage = "Calculate secure passwords"
	app.Version = "0.0.1"
	app.Author = "Karteek E"
	app.Email = "me [AT] karteek [DOT] net"
	app.HideHelp = true
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "Be verbose",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:      "list",
			ShortName: "l",
			Usage:     "List sites in hap2 config file",
			Action:    listSites,
		},
		{
			Name:      "site",
			ShortName: "s",
			Usage:     "Manage sites in the hap2 config file",
			HideHelp:  true,
			Subcommands: []cli.Command{
				{
					Name:      "add",
					ShortName: "a",
					Usage:     "Add a site to hap2 config file",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "force, f",
							Usage: "Force the change",
						},
						cli.StringFlag{
							Name:  "nick",
							Value: "test",
							Usage: "Nickname for the site",
						},
						cli.StringFlag{
							Name:  "user",
							Value: "alice",
							Usage: "Username for the site",
						},
						cli.StringFlag{
							Name:  "salt",
							Value: "1",
							Usage: "Salt for the site, default is 1",
						},
						cli.StringFlag{
							Name:  "domain",
							Value: "foo.com",
							Usage: "Domain name which is being added",
						},
						cli.IntFlag{
							Name:  "length",
							Value: 12,
							Usage: "Length of calculated password, default is 12",
						},
						cli.StringFlag{
							Name:  "suffix",
							Usage: "Suffix for password to be added [optional]",
						},
						cli.StringFlag{
							Name:  "notes",
							Usage: "Notes for the site [optional]",
						},
					},
					Action: addSiteToList,
				},
				{
					Name:      "remove",
					ShortName: "r",
					Usage:     "Remove a site from hap2 config file",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "nick",
							Value: "test",
							Usage: "Nickname for the site",
						},
					},
					Action: removeSiteFromList,
				},
			},
			Action: listSites,
		},
	}
	app.Action = func(c *cli.Context) {
		println("Main Place")
	}
	app.Run(os.Args)
}

func readSiteList(c *cli.Context) *siteCollection {
	if c.GlobalBool("verbose") {
		fmt.Printf("[Debug] Reading site list from %s\n", SiteListFile)
	}
	sl := new(siteCollection)
	// Read from existing config file
	data, _ := ioutil.ReadFile(SiteListFile)
	json.Unmarshal(data, &sl.list)
	return sl
}

func addSiteToList(c *cli.Context) {
	if len(c.String("nick")) < 3 {
		fmt.Print("[Error] Site nickname should be atleast 3 characters\n")
		return
	}

	sl := readSiteList(c)
	s := site{}
	s.User = c.String("user")
	s.Salt = c.String("salt")
	s.Domain = c.String("domain")
	s.Length = c.Int("length")
	s.Suffix = c.String("suffix")
	s.Notes = c.String("notes")

	if sl.list == nil {
		sl.list = make(map[string]site)
	}

	if _, ok := sl.list[c.String("nick")]; ok {
		if !c.Bool("force") {
			fmt.Printf("[Error] Site `%s` is already in the config file, use --force | -f flag to modify existing site\n", c.String("nick"))
			return
		}
	} else {
		if c.String("user") == "alice" {
			if !c.Bool("force") {
				fmt.Print("[Error] Username left as default value `alice`. If you mean it, use --force | -f to make this change\n")
				return
			}
		}

		if c.String("domain") == "foo.com" {
			if !c.Bool("force") {
				fmt.Print("[Error] Domain left as default value `foo.com`. If you mean it, use --force | -f to make this change\n")
				return
			}
		}
	}

	sl.list[c.String("nick")] = s
	writeSiteList(sl)
	fmt.Printf("Added site `%s` with data `%s+%s@%s` to the config\n", c.String("nick"), c.String("user"), c.String("salt"), c.String("domain"))
}

func writeSiteList(slist *siteCollection) {
	data, _ := json.Marshal(slist.list)
	ioutil.WriteFile(SiteListFile, data, 0644)
}

func listSites(c *cli.Context) {
	fmt.Println("List of sites in the config")
	sl := readSiteList(c)
	for key, _ := range sl.list {
		fmt.Printf("%s => %s+%s@%s\n", key, sl.list[key].User, sl.list[key].Salt, sl.list[key].Domain)
	}
}

func removeSiteFromList(c *cli.Context) {
	sl := readSiteList(c)
	if _, ok := sl.list[c.String("nick")]; ok {
		delete(sl.list, c.String("nick"))
		writeSiteList(sl)
		fmt.Printf("Deleted `%s` from config\n", c.String("nick"))
	} else {
		fmt.Printf("[Error] Site `%s` not preset in the config. Can't delete\n", c.String("nick"))
	}
}
