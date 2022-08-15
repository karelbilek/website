package main

// this created "stuff.json" and spits it into stdout

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// this is a scraper for SuperMegaMonkey chronocomics page,
// tailor specifically just for chronocomics

// need to have: archive from here
// https://archive.org/details/megamonkey.tar
// as megamonkey.tar.gz, untarred

func getNameOrdered() map[string]int {

	yearNames := strings.Split(lines, "\n")
	comicsMap := map[string][]string{}

	file, err := os.Open("megamonkey.tar.gz")
	if err != nil {
		panic(err)
	}

	archive, err := gzip.NewReader(file)

	if err != nil {
		panic(err)
	}
	tr := tar.NewReader(archive)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		yRegex := regexp.MustCompile("/([0-9]{4})/index.html$")

		if yRegex.MatchString(hdr.Name) ||
			hdr.Name == "www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/golden_age/index.html" ||
			hdr.Name == "www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/monster_age/index.shtml.html" ||
			hdr.Name == "www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/hero_gap/index.shtml.html" {

			//fmt.Println(hdr.Name)
			doc, err := goquery.NewDocumentFromReader(tr)

			if err != nil {
				panic(err)
			}

			doc.Find(".dragcheck").Each(func(i int, s *goquery.Selection) {
				h, _ := s.Find("a").Attr("href")

				eRegex := regexp.MustCompile("entries/(.*)$")
				if !eRegex.MatchString(h) {
					fmt.Println(h)
					panic("no")
				}

				submatches := eRegex.FindStringSubmatch(h)
				// if strings.Contains(hdr.Name, "1994") {
				// 	fmt.Println(submatches[1])
				// }
				//fmt.Println(submatches[1])

				comicsMap[hdr.Name] = append(comicsMap[hdr.Name], submatches[1])
			})
		}
	}

	resArr := []string{}
	for _, yName := range yearNames {
		resArr = append(resArr, comicsMap[yName]...)
	}
	resMap := map[string]int{}
	for i, r := range resArr {
		resMap[r] = i + 1
		//fmt.Println(r, i+1)
	}
	return resMap
}

func getNameMap() map[string]bool {
	file, err := os.Open("megamonkey.tar.gz")
	if err != nil {
		panic(err)
	}

	archive, err := gzip.NewReader(file)

	if err != nil {
		panic(err)
	}
	tr := tar.NewReader(archive)

	nameMap := map[string]bool{}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		yRegex := regexp.MustCompile("/([0-9]{4})/index.html$")

		if yRegex.MatchString(hdr.Name) ||
			hdr.Name == "www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/golden_age/index.html" ||
			hdr.Name == "www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/monster_age/index.shtml.html" ||
			hdr.Name == "www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/hero_gap/index.shtml.html" {

			//fmt.Println(hdr.Name)
			doc, err := goquery.NewDocumentFromReader(tr)

			if err != nil {
				panic(err)
			}

			doc.Find(".dragcheck").Each(func(i int, s *goquery.Selection) {
				h, _ := s.Find("a").Attr("href")

				eRegex := regexp.MustCompile("entries/(.*)$")
				if !eRegex.MatchString(h) {
					fmt.Println(h)
					panic("no")
				}

				submatches := eRegex.FindStringSubmatch(h)
				//fmt.Println(submatches[1])

				nameMap[submatches[1]] = true
			})
		}
	}
	return nameMap
}

const lines = `www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/golden_age/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/monster_age/index.shtml.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/hero_gap/index.shtml.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/silver_age/1962/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/silver_age/1963/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/silver_age/1964/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/silver_age/1965/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/silver_age/1966/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/silver_age/1967/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/silver_age/1968/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/silver_age/1969/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/silver_age/1970/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/silver_age/1971/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_roy_thomas/1972/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_roy_thomas/1973/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_roy_thomas/1974/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_len_wein/1975/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_len_wein/1976/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_archie_goodwin/1977/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_archie_goodwin/1978/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_jim_shooter/1979/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_jim_shooter/1980/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_jim_shooter/1981/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_jim_shooter/1982/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_jim_shooter/1983/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_jim_shooter/1984/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_jim_shooter/1985/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_jim_shooter/1986/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_jim_shooter/1987/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_tom_defalco/1988/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_tom_defalco/1989/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_tom_defalco/1990/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_tom_defalco/1991/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_tom_defalco/1992/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_tom_defalco/1993/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_tom_defalco/1994/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_silos/1995/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_bob_harras/1996/index.html
www.supermegamonkey.net/www.supermegamonkey.net/chronocomic/eic_bob_harras/1997/index.html`

func multiSplit(s string, cuts []string) []string {
	res := []string{s}
	for _, c := range cuts {
		newRes := []string{}
		for _, inS := range res {
			newRes = append(newRes, strings.Split(inS, c)...)
		}
		res = newRes
	}
	return res
}

type res struct {
	Year       int      `json:"year"`
	Characters []string `json:"characters"`
	Name       string   `json:"name"`
	Link       string   `json:"link"`
	Credits    []string `json:"credits"`
	I          int      `json:"i"`
}

func main() {
	nameMap := getNameMap()
	nameOrdered := getNameOrdered()

	file, err := os.Open("megamonkey.tar.gz")
	if err != nil {
		panic(err)
	}

	archive, err := gzip.NewReader(file)

	if err != nil {
		panic(err)
	}
	tr := tar.NewReader(archive)

	eRegex := regexp.MustCompile(`/entries/(.*\.shtml\.html)$`)
	coverDateR := regexp.MustCompile(`<b>Cover Date:</b>([^<]*)(<br />|</p>)`)
	topDateR := regexp.MustCompile(`^(\d\d\d\d)-\d\d-\d\d`)
	endNumR := regexp.MustCompile(`(\d\d)$`)

	sspecialCases := map[string]int{}
	for k, v := range specialCases {
		sspecialCases[k+".shtml.html"] = v
	}
	sskipped := map[string]bool{}
	for k, v := range skipped {
		sskipped[k+".shtml.html"] = v
	}

	allComics := []res{}

COMICS:

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		if eRegex.MatchString(hdr.Name) {

			submatches := eRegex.FindStringSubmatch(hdr.Name)

			fname := submatches[1]
			if sskipped[fname] {
				continue COMICS
			}

			if nameOrdered[fname] == 0 {
				panic("hm oh no" + fname)
			}

			if !nameMap[fname] {
				panic("WHAT " + fname)
			}
			//fmt.Println(nameOrdered[fname])

			r, err := io.ReadAll(tr)
			if err != nil {
				panic(err)
			}
			s := string(r)
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(r))
			if err != nil {
				panic(err)
			}

			getYear := func() int {
				if special, isSpecial := sspecialCases[fname]; isSpecial {
					return special
				}

				coverDateSM := coverDateR.FindStringSubmatch(s)
				if len(coverDateSM) != 0 {
					complexDate := strings.TrimSpace(coverDateSM[1])
					if complexDate != "" {
						complexDateSpl := multiSplit(complexDate, []string{"-", " - ", "/", " / "})
						complexDateLast := complexDateSpl[len(complexDateSpl)-1]
						//fmt.Println(complexDate)
						//fmt.Println(complexDateLast)
						yearSM := endNumR.FindStringSubmatch(complexDateLast)
						//fmt.Println(complexDateLast)

						realYear := yearSM[1]
						realYearI, err := strconv.Atoi(realYear)
						if err != nil {
							panic(err)
						}
						if realYearI < 17 {
							realYearI += 2000
						} else if realYearI > 38 {
							realYearI += 1900
						} else {
							panic(realYear)
						}
						return realYearI
					}

				}

				k := doc.Find(".content-nav .smalltext").Text()
				topCoverDateSM := topDateR.FindStringSubmatch(k)
				topCoverDateY := topCoverDateSM[1]
				topCoverDateYI, err := strconv.Atoi(topCoverDateY)
				if err != nil {
					panic(err)
				}
				return topCoverDateYI
			}

			year := getYear()

			credits := []string{}

			doc.Find(".bodytext>p").Each(func(i int, s *goquery.Selection) {
				h, e := s.Html()
				if e != nil {
					panic(e)
				}
				_, af, has := strings.Cut(h, "<b>Credits:</b>")
				if has {
					lines := multiSplit(af, []string{"<br>", "<br/>", "<br />"})
					names := []string{}

				LINES:
					for _, l := range lines {
						if strings.Contains(l, "Review") {
							break LINES
						}
						almostname := multiSplit(l, []string{"-", " - "})
						allNames := multiSplit(almostname[0], []string{"/", " / ", "&amp;", " and ", " with ", " &amp; ", ",", ", "})
						for _, n := range allNames {
							ff := strings.TrimSpace(n)
							ff, _, _ = strings.Cut(ff, "(")
							ff = html.UnescapeString(ff)

							if ff != "" {
								names = append(names, ff)
							}
						}
					}

					credits = append(credits, names...)
					//panic("hm")
				}
			})
			sname := strings.TrimSuffix(fname, ".shtml.html")

			characters := []string{}

			doc.Find(".bodytext").Each(func(i int, s *goquery.Selection) {
				h, e := s.Html()
				if e != nil {
					panic(e)
				}

				if strings.Contains(h, "<b>Characters Appearing:</b>") {
					s.Find("a").Each(func(i int, s *goquery.Selection) {
						characters = append(characters, s.Text())
					})
				}
			})

			he, er := doc.Find("h3.header").Html()
			if er != nil {
				panic(er)
			}

			titleSplit := multiSplit(he, []string{"<br>", "<br/>", "<br />"})
			titleAll := ""
			for _, s := range titleSplit {
				if titleAll != "" {
					titleAll += ", "
				}
				titleAll += html.UnescapeString(s)
			}

			re := res{
				Year:       year,
				Characters: characters,
				Name:       titleAll,
				Link:       sname,
				Credits:    credits,
				I:          nameOrdered[fname],
			}

			allComics = append(allComics, re)
		}
	}

	sort.Slice(allComics, func(i, j int) bool {
		return allComics[i].I < allComics[j].I
	})

	js, _ := json.MarshalIndent(allComics, "", "    ")
	fmt.Printf("%s\n", js)
}

var specialCases = map[string]int{
	"hulk_194":                       1975,
	"monsters_on_the_prowl_24":       1973,
	"tomb_of_dracula_10vampire_tale": 1973,
	"hulk_407-408_pantheon":          1993,
	"marvel_spotlight_14":            1973,
	"where_monsters_dwell_25":        1973,
	"marvel_two-in-one_16-17":        1976,
	"x-terminators_3":                1988,
	"fantastic_four_326-328":         1989,
	"quasar_37":                      1992,
	// "tales_to_astonish_20tales_to_a.shtml.html": 1961,
	// "marvel_mystery_comics_17captai.shtml.html": 1944,
}

var skipped = map[string]bool{
	"captain_america_450-454":    true,
	"End_Of_Line":                true,
	"cage_1-2":                   true,
	"blackwulf_2-5":              true,
	"punisher_war_journal_70-75": true,

	///
	"tales_of_the_marvels_blockbust": true,
	"knights_of_pendrage_1-6":        true,
	"batman_384":                     true,
}
