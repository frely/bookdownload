package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	bookIndexList     [][]string
	home              string = "https://www.changyexiaoshuo.com"
	url               string
	bookName          string
	author            string
	latestChapterName string
	fileName          string
	firstRun          int = 0
)

func getIndex(url string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// 查找书名
	bookName = doc.Find("h1").Text()

	// 查找作者
	author = doc.Find(".fix a").First().Text()

	// 查找章节信息
	doc.Find(".section-box").Last().Find("li").Each(func(i int, s *goquery.Selection) {
		chapterName := s.Find("a").Text()
		chapterUrl, exists := s.Find("a").Attr("href")
		latestChapterName = chapterName
		if exists {
			var list []string
			list = append(list, chapterName, chapterUrl)
			bookIndexList = append(bookIndexList, list)
		}
	})

	// 查找下一页
	doc.Find(".right").Each(func(i int, s *goquery.Selection) {
		nextUrl, exists := s.Find("a").Attr("href")
		if exists {
			time.Sleep(2 * time.Second)
			getIndex(nextUrl)
		}
	})
}

func getChapterInfo(bookIndexList [][]string) {
	for _, index := range bookIndexList {
		chapterName := index[0]
		chapterUrl := index[1]

		// 写入文件
		fileName = fmt.Sprintf("《%s》作者:%s[更新至%s]", bookName, author, latestChapterName)
		if firstRun == 0 {
			os.Remove(fileName)
		}
		f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		firstRun = 1

		// 写入章节名
		if _, err := io.WriteString(f, "\n\n"+chapterName+"\n\n"); err != nil {
			log.Fatal(err)
		}

		getChapterData(chapterName, chapterUrl)
		fmt.Println("已下载：", chapterName)
	}
}

func getChapterData(chapterName string, chapterUrl string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Get(chapterUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// 写入文件
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// 写入章节内容
	doc.Find(".content p").Each(func(i int, s *goquery.Selection) {
		if _, err := io.WriteString(f, s.Text()+"\n"); err != nil {
			log.Fatal(err)
		}
	})

	// 查找下一页
	nextUrl, exists := doc.Find("#next_url").Attr("href")
	if exists {
		if string(nextUrl[len(nextUrl)-7]) == "_" {
			time.Sleep(2 * time.Second)
			getChapterData(chapterName, home+nextUrl)
		}
	}
}

func main() {
	fmt.Println("请输入书籍地址：")
	fmt.Scanln(&url)
	fmt.Println("开始搜索章节，请等待。。。")
	getIndex(url)
	getChapterInfo(bookIndexList)
	fmt.Println("下载已完成。")
}
