package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	entry = "http://www.stats.gov.cn/tjsj/tjbz/tjyqhdmhcxhfdm/2018/index.html" // 入口
)

var (
	allowedDomains = []string{"www.stats.gov.cn"}

	provinceCollector *colly.Collector
	cityCollector     *colly.Collector
	countyCollector   *colly.Collector

	baseCollector *colly.Collector
)

func main() {
	flag.Parse()

	baseCollector = colly.NewCollector(
		colly.Async(true),
	)
	baseCollector.AllowedDomains = allowedDomains
	baseCollector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 10,              // 并行任务数
		RandomDelay: 5 * time.Second, // 5s内随机延迟
	})

	initProvinceCollector()
	initCityCollector()
	initCountyCollector()

	InitDb()
	defer db.Close()

	provinceCollector.Visit(entry)

	// 使用异步请求,需等待线程结束
	provinceCollector.Wait()
	cityCollector.Wait()
	countyCollector.Wait()
}

func GbkToUtf8(s string) string {
	b := []byte(s)
	reader := transform.NewReader(bytes.NewReader(b), simplifiedchinese.GBK.NewDecoder())
	d, _ := ioutil.ReadAll(reader)
	return string(d)
}

func logRequest(r *colly.Request) {
	fmt.Println(time.Now(), "Visiting", r.URL.String())
}

func logResponse(r *colly.Response) {
	fmt.Println("response = ", string(r.Body))
}

func logError(r *colly.Response, e error) {
	fmt.Println("error = ", e)
	// 出错时重试3次
	var retry int
	if r.Request.Ctx.GetAny("retry") == nil {
		retry = 0
	} else {
		retry = r.Request.Ctx.GetAny("retry").(int)
	}
	if retry < 3 {
		r.Request.Ctx.Put("retry", retry+1)
		r.Request.Retry()
	}
}

func collectorFromBase() *colly.Collector {
	collector := baseCollector.Clone()
	collector.OnRequest(logRequest)
	collector.OnError(logError)

	return collector
}

func initProvinceCollector() {
	provinceCollector = collectorFromBase()

	provinceCollector.OnHTML("tr.provincetr", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		url = url[:strings.LastIndex(url, "/")+1]
		e.ForEach("a", func(i int, el *colly.HTMLElement) {
			name := GbkToUtf8(el.Text)
			href := el.Attr("href")
			code, _ := strconv.Atoi(strings.Split(href, ".")[0])
			code = code * 10000000
			area := &Area{
				ID:     code,
				Name:   name,
				Parent: 0,
			}
			InsertArea(area)
			// if code == 150000000 {
			cityCollector.Visit(url + href)
			// }
		})
	})
}

func initCityCollector() {
	cityCollector = collectorFromBase()

	cityCollector.OnHTML("tr.citytr", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		url = url[:strings.LastIndex(url, "/")+1]
		e.ForEach("td", func(i int, el *colly.HTMLElement) {
			if i == 1 {
				name := GbkToUtf8(el.ChildText("a"))
				href := el.ChildAttr("a", "href")
				code, _ := strconv.Atoi(strings.Split(strings.Split(href, ".")[0], "/")[1])
				parent := code / 100 * 10000000
				code = code * 100000
				area := &Area{
					ID:     code,
					Name:   name,
					Parent: parent,
				}
				InsertArea(area)
				countyCollector.Visit(url + href)
			}
		})
	})
}

func initCountyCollector() {
	countyCollector = collectorFromBase()

	countyCollector.OnHTML("tr.countytr, tr.towntr", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		url = url[:strings.LastIndex(url, "/")+1]
		text := e.ChildText("td")
		code, _ := strconv.Atoi(text[:9])
		name := GbkToUtf8(text[12:])
		var parent int
		if e.Attr("class") == "countytr" {
			parent = code / 100000 * 100000
		} else {
			parent = code / 1000 * 1000
		}
		area := &Area{
			ID:     code,
			Name:   name,
			Parent: parent,
		}
		InsertArea(area)
		if e.Attr("class") == "countytr" {
			e.ForEach("td", func(i int, el *colly.HTMLElement) {
				if i == 1 {
					href := el.ChildAttr("a", "href")
					if href != "" {
						countyCollector.Visit(url + href)
					}
				}
			})
		}
	})
}
