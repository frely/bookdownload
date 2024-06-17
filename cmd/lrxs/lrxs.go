package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var (
	bookIndexList     [][]string
	home              string = "https://www.lrxs.org"
	url               string
	bookName          string
	author            string
	latestChapterName string
	fileName          string
	firstRun          int = 0
)

func main() {
	fmt.Println("请输入书籍地址：")
	fmt.Scanln(&url)
	fmt.Println("开始搜索章节，请等待。。。")
	getIndex(url)
	getChapterInfo(bookIndexList)
	fmt.Println("下载已完成。")
}

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

	ct := res.Header.Get("Content-Type")
	bodyReader, err := charset.NewReader(res.Body, ct)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(bodyReader)
	if err != nil {
		log.Fatal(err)
	}

	// 查找书名
	bookName = doc.Find("h1").Text()

	// 查找作者
	author = string(doc.Find("#info p").First().Text())[17:]

	// 查找章节信息
	doc.Find("#list dd").Each(func(i int, s *goquery.Selection) {
		chapterName := s.Find("a").Text()
		chapterUrl, exists := s.Find("a").Attr("href")
		chapterUrl = home + chapterUrl
		latestChapterName = chapterName
		if exists {
			var list []string
			list = append(list, chapterName, chapterUrl)
			bookIndexList = append(bookIndexList, list)
		}
	})
}

func getChapterInfo(bookIndexList [][]string) {
	for _, index := range bookIndexList {
		chapterName := index[0]
		chapterUrl := index[1]

		// 写入文件
		fileName = fmt.Sprintf("《%s》作者:%s[更新至%s].txt", bookName, author, latestChapterName)
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
		time.Sleep(2 * time.Second)
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

	ct := res.Header.Get("Content-Type")
	bodyReader, err := charset.NewReader(res.Body, ct)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(bodyReader)
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
	if _, err := io.WriteString(f, doc.Find("#content").Text()+"\n"); err != nil {
		log.Fatal(err)
	}
}

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}
