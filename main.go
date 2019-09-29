package main

import (
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
	"github.com/urfave/cli"
	"github.com/wcharczuk/go-chart"
	"os"
	"os/user"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
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

func getLiverNames(url string) (map[string]string, error) {
	doc, err := goquery.NewDocument(url)
	namesMap := map[string]string{}
	var regxNewline = regexp.MustCompile(`\r\n|\r|\n`)
	doc.Find(".roundcorner").Each(func(index int, s *goquery.Selection) {
		if strings.TrimSpace(s.Text()) != "" {
			href, _ := s.Find("a").Attr("href")
			href = strings.Split(href, "/")[len(strings.Split(href, "/"))-2]
			namesMap[strings.TrimSpace(regxNewline.ReplaceAllString(s.Find("span").Text(), ""))] = href
		}
	})
	return namesMap, err
}

func getLiverAliasMapFromCsv(filepath string) [][]string {
	f, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	c := csv.NewReader(f)
	result := [][]string{}
	crec, er := c.ReadAll()
	if er != nil {
		panic(er)
	}
	for _, r := range crec[1:] {
		result = append(result, []string{r[0], r[1], r[2]})
	}
	return result
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

func saveLiverNamesToCsv(headers []string, nm map[string]string) {
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
	for k, v := range nm {
		w.Write([]string{k, k, v})
		// processing of retention point
		if strings.Index(k, "・") >= 0 {
			sp_names := strings.Split(k, "・")
			w.Write([]string{k, sp_names[0], v})
			w.Write([]string{k, sp_names[1], v})
		} else if inHiraganaAndKatakana(k) && utf8.RuneCountInString(k) <= 3 {
			continue
		} else if inHiraganaAndKatakana(k) {
			a, b := splitHiraganaAndKatakana(k)
			w.Write([]string{k, a, v})
			w.Write([]string{k, b, v})
		}
	}

	// I want to be first and last name separated if possible...!
	// but I can't it.
	// I want a force...!
}

func determineWhatLiverFromString(str string, nameMap map[string]string) string {
	for k, _ := range nameMap {
		if strings.Index(str, k) >= 0 {
			return k
		}
	}
	return "other"
}

func createChartVals(pathquerys []string, namesCsvpath string) []chart.Value {
	csvList := getLiverAliasMapFromCsv(namesCsvpath)
	nameAliasMap := map[string]string{}
	for _, r := range csvList {
		nameAliasMap[r[0]] = r[1]
	}
	nameRomajiMap := map[string]string{}
	for _, r := range csvList {
		nameRomajiMap[r[0]] = r[2]
	}
	nameCnt := map[string]int{}
	vals := []chart.Value{}
	for _, f := range pathquerys {
		f, _ := os.Open(f)
		defer f.Close()
		c := csv.NewReader(f)
		allRec, er := c.ReadAll()
		if er != nil {
			panic(er)
		}
		for _, row := range allRec[1:] {
			name := determineWhatLiverFromString(row[0], nameAliasMap)
			if name != "other" {
				nameCnt[nameRomajiMap[name]] += 1
			} else {
				nameCnt[name] += 1
			}
		}
		allRec = [][]string{}
	}
	for k, v := range nameCnt {
		vals = append(vals, chart.Value{Value: float64(v), Label: k})
	}
	return vals
}

func graph(vals []chart.Value, savepath, title string) {
	bar := chart.BarChart{
		Title: title,
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		Width:  1500,
		Height: 512,
		XAxis:  chart.StyleShow(),
		YAxis: chart.YAxis{
			Style:     chart.StyleShow(),
			NameStyle: chart.Style{Show: true},
		},
		Bars: vals,
	}

	f, _ := os.Create(savepath)
	defer f.Close()
	bar.Render(chart.PNG, f)
}
func main() {
	app := cli.NewApp()
	app.Name = "nico-ranking-processing"
	app.Usage = "my niconico api processing tool"
	app.Version = "0.0.1"
	nijisanji_url := "https://nijisanji.ichikara.co.jp/member/"

	app.Commands = []cli.Command{
		{
			Name:    "getLiverNames",
			Aliases: []string{},
			Usage:   "get liver names",
			Action: func(c *cli.Context) error {
				nm, e := getLiverNames(nijisanji_url)
				if e != nil {
					panic("failed get name")
				}
				for k, v := range nm {
					fmt.Println(k + ":" + v)
				}
				return nil
			},
		},
		{
			Name:    "saveLiverNamesToCsv",
			Aliases: []string{},
			Usage:   "get liver names and save csv",
			Action: func(c *cli.Context) error {
				nm, e := getLiverNames(nijisanji_url)
				if e != nil {
					panic("failed get name")
				}
				saveLiverNamesToCsv([]string{"name", "alias", "romaji"}, nm)
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
				querypaths := []string{}
				for _, path := range c.Args()[0:(c.NArg() - 1)] {
					usr, _ := user.Current()
					querypath := strings.Replace(path, "~", usr.HomeDir, 1)
					querypaths = append(querypaths, querypath)
				}
				namesCsvpath := c.Args().Get(c.NArg() - 1)
				vals := createChartVals(querypaths, namesCsvpath)
				sort.Slice(vals, func(i, j int) bool {
					return vals[i].Value > vals[j].Value
				})
				total := 0.0
				for _, v := range vals {
					if v.Label != "other" {
						total += v.Value
					}
				}
				if len(vals) > 10 {
					vals = vals[0:11]
				}
				delptr := -1
				for i, v := range vals {
					if v.Label == "other" {
						delptr = i
					} else {
						v_s := fmt.Sprintf("%f", v.Value)
						vals[i].Label += "\n" + v_s
						fmt.Println(v.Label + ":" + v_s)
					}
				}
				if delptr >= 0 {
					vals = append(vals[:delptr], vals[delptr+1:]...)
				}
				graph(vals, "liver_cuts_top_10.png", "liver cuts top 10")
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
