package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

var (
	browser     string
	url         string
	bookName    string
	author      string
	chapterList [][]string
	fileName    string
	firstRun    int = 0
	sys         string
)

func main() {
	sysEnv()
	getIndex()
	writeFile()
}

func sysEnv() {
	sys = runtime.GOOS
	if sys == "windows" {
		browser = "C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe"
	} else {
		fmt.Println("当前系统不是 Windows, 暂未支持.")
		os.Exit(0)
	}
}

func getIndex() {
	fmt.Println("请输入书籍地址：")
	fmt.Scanln(&url)
	url = strings.TrimSpace(url)
	fmt.Println("开始搜索章节，请等待。。。")

	u := launcher.New().Bin(browser).MustLaunch()
	page := rod.New().ControlURL(u).MustConnect().MustPage(url)
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
		chapterUrl, err := v.Attribute("href")
		if err != nil {
			continue
		}
		chapterName, err := v.Text()
		if err != nil || chapterName == "" {
			continue
		}
		tmpList := make([]string, 0)
		tmpList = append(tmpList, chapterName, *chapterUrl)
		chapterList = append(chapterList, tmpList)
	}
}

func writeFile() {
	for _, index := range chapterList {
		chapterName := index[0]
		chapterUrl := index[1]

		// 写入文件
		fileName = fmt.Sprintf("《%s》作者：%s.txt", bookName, author)
		if firstRun == 0 {
			os.Remove(fileName)
			firstRun = 1
		}
		f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			log.Fatal("write file err\n", err)
		}
		defer f.Close()
		if _, err := io.WriteString(f, "\n"+chapterName+"\n\n"); err != nil {
			log.Fatal("write chapter err\n", err)
		}
		getChapterData(chapterUrl)
		fmt.Println("已下载：", chapterName)
	}
	fmt.Println("下载已完成。")
}

func getChapterData(url string) {
	u := launcher.New().Bin(browser).MustLaunch()
	page := rod.New().ControlURL(u).MustConnect().MustPage(url)
	defer page.MustClose()
	page.Timeout(5 * time.Second)

	ps, err := page.MustElement("#J_BookRead").Elements("p")
	if err != nil {
		log.Fatal("getChapterData err\n", err)
	}
	for _, p := range ps {
		i, _ := p.HTML()
		findStr := regexp.MustCompile(`<p class="chapter">(.*?)<span>`).FindStringSubmatch(i)
		if findStr != nil {
			f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				log.Fatal("write file err\n", err)
			}
			defer f.Close()
			if _, err := io.WriteString(f, strings.TrimSpace(findStr[1])+"\n"); err != nil {
				log.Fatal("write chapter err\n", err)
			}
		}
	}
}
