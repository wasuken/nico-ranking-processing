package main

import (
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
	"github.com/urfave/cli"
	"os"
	"os/user"
	"regexp"
	"strings"
	"time"
	"unicode"
)

func rankingCsv(filepath string) {
	usr, _ := user.Current()
	f := strings.Replace(filepath, "~", usr.HomeDir, 1)
	wfile, err := os.Create(f + time.Now().Format("2006-01-02_15:04:05.000") + ".csv")
	if err != nil {
		fmt.Println(err)
		panic("file error.")
	}
	fp := gofeed.NewParser()
	url := "https://www.nicovideo.jp/ranking/genre/entertainment?tag=%E3%81%AB%E3%81%98%E3%81%95%E3%82%93%E3%81%98&rss=2.0&lang=ja-jp"
	feed, _ := fp.ParseURL(url)
	writer := csv.NewWriter(wfile)

	writer.Write([]string{"title", "link"})
	for _, item := range feed.Items {
		writer.Write([]string{item.Title, item.Link})
	}
	writer.Flush()
}

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

func inHiraganaAndKatakana(str string) bool {
	for _, r := range str {
		if unicode.In(r, unicode.Hiragana) || unicode.In(r, unicode.Katakana) {
			return true
		}
	}
	return false
}

// Boundary line is abnormal behavior and are two or moreBoundary line is abnormal behavior and are two or more!
func splitHiraganaAndKatakana(str string) (string, string) {
	var a, b string
	flg := true
	for _, r := range str {
		if unicode.In(r, unicode.Hiragana) || unicode.In(r, unicode.Katakana) {
			flg = false
		}
		if flg {
			a += string([]rune{r})
		} else {
			b += string([]rune{r})
		}
	}
	return a, b
}

func saveLiverNamesToCsv(headers, names []string) {
	_, err := os.Stat("liverNamesAndAlias.csv")
	if err == nil {
		e := os.Remove("liverNamesAndAlias.csv")
		if e != nil {
			panic(e)
		}
	}
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
		// processing of retention point
		if strings.Index(name, "・") >= 0 {
			sp_names := strings.Split(name, "・")
			w.Write([]string{name, sp_names[0]})
			w.Write([]string{name, sp_names[1]})
		} else if inHiraganaAndKatakana(name) {
			a, b := splitHiraganaAndKatakana(name)
			w.Write([]string{name, a})
			w.Write([]string{name, b})
		}
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
		{
			Name:    "rankingCsv",
			Aliases: []string{},
			Usage:   "save ranking csv",
			Action: func(c *cli.Context) error {
				filepath := c.Args().First()
				rankingCsv(filepath)
				return nil
			},
		},
		{
			Name:    "graph",
			Aliases: []string{},
			Usage:   "create graph",
			Action: func(c *cli.Context) error {
				fmt.Println("Under construction")
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
