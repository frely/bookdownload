package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

const indexUrl string = "https://www.ciweimao.com/book/100171528"

var (
	browser  string
	bookName string
	author   string
)

func main() {
	sysEnv()

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

	// 查找书籍信息
	metaTags, _ := page.Elements("meta")
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
}
