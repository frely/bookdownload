package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

const indexUrl string = "https://www.ciweimao.com/book/100171528"

var (
	browser     string
	bookName    string
	author      string
	chapterList [][]string
)

func main() {
	sysEnv()
	getIndex()
	fmt.Println(chapterList)

}

func sysEnv() {
	sys := runtime.GOOS
	if sys == "windows" {
		browser = "C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe"
	} else {
		fmt.Println("当前系统不是 Windows, 暂未支持.")
		os.Exit(0)
	}
}

func getIndex() {
	u := launcher.New().Bin(browser).MustLaunch()
	page := rod.New().ControlURL(u).MustConnect().MustPage(indexUrl)
	defer page.MustClose()
	page.Timeout(5 * time.Second)

	// 查找书籍信息：书名，作者
	metaTags, err := page.Elements("meta")
	if err != nil {
		log.Fatalln("no find meta")
	}
	for _, meta := range metaTags {
		property := meta.MustAttribute("property")
		content := meta.MustAttribute("content")
		if property != nil && *property == "og:novel:book_name" {
			bookName = *content
		}
		if property != nil && *property == "og:novel:author" {
			author = *content
		}
	}
	chapterPage, err := page.Element(".btn-more")
	if err != nil {
		log.Fatalln("not find chapterPage")
	}
	chapterPageUrl := chapterPage.MustAttribute("href")

	// 查找章节列表
	index := rod.New().ControlURL(u).MustConnect().MustPage(*chapterPageUrl)
	defer index.MustClose()
	index.Timeout(5 * time.Second)
	bookChapterBox, _ := index.MustElement(".book-chapter").Elements("a")
	for _, v := range bookChapterBox {
		chapterName, err := v.Attribute("href")
		if err != nil {
			continue
		}
		chapterUrl, err := v.Text()
		if err != nil || chapterUrl == "" {
			continue
		}
		tmpList := make([]string, 0)
		tmpList = append(tmpList, *chapterName, chapterUrl)
		chapterList = append(chapterList, tmpList)
	}
}
