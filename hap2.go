package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"github.com/atotto/clipboard"
	"github.com/codegangsta/cli"
	"github.com/howeyc/gopass"
)

type site struct {
	User     string `json:"user"`
	Salt     string `json:"salt"`
	Domain   string `json:"domain"`
	Length   int    `json:"length"`
	Suffix   string `json:"suffix"`
	Notes    string `json:"notes"`
	Security string `json:"security"`
	Check    string `json:"check"`
}

type siteCollection struct {
	list map[string]site
}

var SiteListFile string

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
							Value: "test",
							Usage: "Username for the site",
						},
						cli.StringFlag{
							Name:  "salt",
							Value: "1",
							Usage: "Salt for the site, default is 1",
						},
						cli.StringFlag{
							Name:  "domain",
							Value: "example.com",
							Usage: "Domain name which is being added",
						},
						cli.IntFlag{
							Name:  "length",
							Value: 14,
							Usage: "Length of calculated password, default is 14",
						},
						cli.StringFlag{
							Name:  "suffix",
							Usage: "Suffix for password to be added [optional]",
						},
						cli.StringFlag{
							Name:  "notes",
							Usage: "Notes for the site [optional]",
						},
						cli.StringFlag{
							Name:  "security",
							Usage: "Security note for the site [optional]",
						},
						cli.StringFlag{
							Name:  "check",
							Usage: "Password Check for the site [optional]",
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
				{
					Name:      "list",
					ShortName: "l",
					Usage:     "List sites in hap2 config file",
					Action:    listSites,
				},
			},
			Action: func(c *cli.Context) {
				cli.HelpPrinter(cli.SubcommandHelpTemplate, c.App)
			},
		},
	}
	app.Action = revealPassword
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

	if c.String("nick") == "test" {
		if !c.Bool("force") {
			fmt.Println("[Error] Site nick left as default value `test`")
			fmt.Println("Use --force | -f to modify the existing site or use --help | -h to see all the options")
			return
		}
	}

	sl := readSiteList(c)

	if sl.list == nil {
		sl.list = make(map[string]site)
	}

	var s site
	if _, ok := sl.list[c.String("nick")]; ok {
		if !c.Bool("force") {
			fmt.Printf("[Error] Site `%s` is already in the config file\n", c.String("nick"))
			fmt.Println("Use --force | -f to modify the existing site or use --help | -h to see all the options")
			return
		}
		s = sl.list[c.String("nick")]
	} else {
		s = site{}
	}

	if c.String("user") != "test" || len(s.User) == 0 {
		s.User = c.String("user")
	}
	if c.String("salt") != "1" || len(s.Salt) == 0 {
		s.Salt = c.String("salt")
	}
	if c.String("domain") != "example.com" || len(s.Domain) == 0 {
		s.Domain = c.String("domain")
	}
	if c.Int("length") != 14 || s.Length == 0 {
		s.Length = c.Int("length")
	}
	if len(c.String("suffix")) > 0 {
		s.Suffix = c.String("suffix")
	}
	if len(c.String("notes")) > 0 {
		s.Notes = c.String("notes")
	}
	if len(c.String("security")) > 0 {
		s.Security = c.String("security")
	}
	if len(c.String("check")) > 0 {
		s.Check = c.String("check")
	}

	sl.list[c.String("nick")] = s
	writeSiteList(sl)
	fmt.Printf("Added site `%s` with data `%s+%s@%s` to the config\n", c.String("nick"), s.User, s.Salt, s.Domain)
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

func listSites(c *cli.Context) {
	fmt.Println("List of sites in the config")
	sl := readSiteList(c)
	for key, _ := range sl.list {
		fmt.Printf("%s => %s+%s@%s\n", key, sl.list[key].User, sl.list[key].Salt, sl.list[key].Domain)
	}
}

func writeSiteList(slist *siteCollection) {
	data, _ := json.MarshalIndent(slist.list, "", "    ")
	ioutil.WriteFile(SiteListFile, data, 0644)
}

func revealPassword(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("[Error] Cannot generate password without knowing the site nickname\n")
		return
	}

	nick := c.Args().First()
	sl := readSiteList(c)
	if s, ok := sl.list[nick]; ok {
		if len(s.Notes) > 0 {
			fmt.Printf("Notes: %s\n", s.Notes)
		}
		if len(s.Security) > 0 {
			fmt.Printf("Security: %s\n", s.Security)
		}
		fmt.Printf("Enter Master password. Hit enter to abort: ")
		masterPass := string(gopass.GetPasswdMasked()[:])
		if len(masterPass) == 0 {
			return
		}
		fmt.Printf("Computing password for %s <%s+%s@%s>\n", nick, s.User, s.Salt, s.Domain)
		password := generatePassword(s, masterPass)
		clipboard.WriteAll(password)
		fmt.Println(password, "copied to your clipboard")

	} else {
		fmt.Printf("[Error] Requested site `%s` not found the config. Adding it first using `site add`\n", nick)
	}
}

func generatePassword(s site, masterPassword string) string {
	key := []byte(masterPassword)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(fmt.Sprintf("%s+%s@%s", s.User, s.Salt, s.Domain)))
	pass := base64.StdEncoding.EncodeToString(h.Sum(nil))[:s.Length]
	if len(s.Suffix) > 0 {
		pass = pass + s.Suffix
	}

	ch := sha256.New()
	ch.Write([]byte(pass))
	check := base64.StdEncoding.EncodeToString(ch.Sum(nil))

	if len(s.Check) > 0 {
		if s.Check != check {
			fmt.Println("[error] Check value mismatch; Wrong master password entered")
			os.Exit(1)
		}
	} else {
		fmt.Printf("Check not found in the config. Consider adding `%s` to config\n", check)
	}
	return pass
}
