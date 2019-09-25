package main

import (
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/urfave/cli"
	"os"
	"regexp"
	"strings"
)

func getLiverNames(url string) ([]string, error) {
	doc, err := goquery.NewDocument(url)
	var namesAry []string
	var regxNewline = regexp.MustCompile(`\r\n|\r|\n`)
	doc.Find(".roundcorner span").Each(func(index int, s *goquery.Selection) {
		if strings.TrimSpace(s.Text()) != "" {
			namesAry = append(namesAry, strings.TrimSpace(regxNewline.ReplaceAllString(s.Text(), "")))
		}
	})
	return namesAry, err
}

func saveLiverNamesToCsv(headers, names []string) {
	csvFile, err := os.Create("liverNamesAndAlias.csv")
	if err != nil {
		panic("failed create csv file")
	}
	defer csvFile.Close()
	w := csv.NewWriter(csvFile)
	defer w.Flush()
	w.Write(headers)
	// insert to fullname
	for _, name := range names {
		w.Write([]string{name, name})
	}
	// I want to be first and last name separated if possible...!
	// but I can't it.
	// I want a force...!

}
func main() {
	app := cli.NewApp()
	app.Name = "nico-ranking-processing"
	app.Usage = "my niconico api processing tool"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		{
			Name:    "getLiverNames",
			Aliases: []string{},
			Usage:   "get liver names",
			Action: func(c *cli.Context) error {
				names, e := getLiverNames("https://nijisanji.ichikara.co.jp/member/")
				if e != nil {
					panic("failed get name")
				}
				for _, name := range names {
					fmt.Println(name)
				}
				return nil
			},
		},
		{
			Name:    "saveLiverNamesToCsv",
			Aliases: []string{},
			Usage:   "get liver names and save csv",
			Action: func(c *cli.Context) error {
				names, e := getLiverNames("https://nijisanji.ichikara.co.jp/member/")
				if e != nil {
					panic("failed get name")
				}
				saveLiverNamesToCsv([]string{"name", "alias"}, names)
				return nil
			},
		},
	}
	er := app.Run(os.Args)
	if er != nil {
		fmt.Println(er)
		panic("error app run.")
	}
}
