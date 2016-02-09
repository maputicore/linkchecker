package main

import (
	"encoding/csv"
	"golang.org/x/text/encoding/japanese"
    "golang.org/x/text/transform"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"regexp"
	charset "github.com/mozillazg/go-charset"
	"os"
	"io"
	"log"
	_ "strings"
)

var regxHttp, _ = regexp.Compile("^http(s)?://")
var regxNoIndex, _ = regexp.Compile("<meta[^>]*noindex.*?meta>")
var regxNoFollow, _ = regexp.Compile("nofollow")
var regxLastSlash, _ = regexp.Compile("/$")
var targetUrl = []string{"http://www.baitoru.com", "http://www.hatarako.net"}
var regxDomain, _ = regexp.Compile("(http(s)?://[^/]+)")

type LinkResult []string
// - url
// - link
// - noindex
// - nofollow

func main() {
	file, err := os.Open("./urls.csv")
	FailOnError(err)
    defer file.Close()
    reader := csv.NewReader(transform.NewReader(file, japanese.ShiftJIS.NewDecoder()))

    // fmt.Println(err)
    // fmt.Println(linkResult)

    file2, err := os.Create("./result.csv")
    FailOnError(err)
    defer file2.Close()
    writer := csv.NewWriter(transform.NewWriter(file2, japanese.ShiftJIS.NewEncoder()))

    writer.Write([]string{"URL", "URLを開く", "ドメイン", "リンクURL", "nofollow", "noindex"})

    for {
        record, err := reader.Read() // 1行読み出す
        if err == io.EOF {
            break
        } else {
            FailOnError(err)
        }

       	url := FormatUrl(record[0])
        linkResult, err := Check(url, targetUrl)

        if err != nil {
        	writer.Write([]string{"[!NG!]", url, ": Can't get the page"})
        	fmt.Println("[!NG!]", url)
    	} else {
    		for _, result := range linkResult {
 				writer.Write(result)
 				fmt.Println("[!OK!]", url)
    		}
    	}
    }
    writer.Flush()
	// fmt.Println(err)
	return
}

func FailOnError(err error) {
    if err != nil {
        log.Fatal("Error:", err)
    }
}


func FormatUrl(url string) string {
	matchHttp := regxHttp.FindStringIndex(url)

	if len(matchHttp) == 0 {
		url = "http://" + url
	}
	return url
}

func Check(url string, targets []string) ([]LinkResult, error){
	doc, err := goquery.NewDocument(url)


	if err != nil {
		return nil, err
	}

	isNoIndex := CheckNoIndex(doc)
	noIndex := ""
	if isNoIndex {
		noIndex = "◯"
	}

	result := make([]LinkResult, 0)
	domain := getDomain(url)

	for _, target := range targets {
		target = stripLastSlash(target)
		doc.Find("a[href*='"+target+"']").Each(func(_ int, link *goquery.Selection) {
			href, _ := link.Attr("href")
			isNoFollow := CheckNoFollow(link)
			noFollow := ""
			if isNoFollow {
				noFollow = "◯"
			}
			result = append(result, LinkResult {url, "", domain, href, noFollow, noIndex})
		})
	}
	return result, nil
}


// func get(url string, link string) (domain string, noindex bool, nofollow bool, err error) ([]LinkResult){
	

// 	if err != nil {
// 		return "",false,false,err
// 	}
// 	// char := GetCharset(doc)

// 	//TODO Convert correct charset
// 	isNoFollow := CheckNoFollow(doc, link)
// 	isNoIndex := CheckNoIndex(doc)
// 	return "", isNoIndex, isNoFollow, nil
// }


func getDomain(url string) string {
	return regxDomain.FindString(url)
}

func GetCharset(doc *goquery.Document) string {
	//Checking meta[charset]
	content, exists := doc.Find("meta[charset]").Attr("charset")

	//Checking meta[chasert]
	if !exists {
		content, exists = doc.Find("meta[http-equiv]").Attr("content")
	} else {
		content = "text/html; charset=" + content
	}

	if !exists {
		return ""
	}
	return charset.Parse(nil, []byte(content))
}

func CheckNoFollow(link *goquery.Selection) bool {
	rel, isExists := link.Attr("rel")
	if !isExists {
		return false
	}
	return regxNoFollow.MatchString(rel)
}

func CheckNoIndex(doc *goquery.Document) bool {
	return doc.Find("meta[content*='noindex']").Length() > 0
	
}

func stripLastSlash(str string) string {
	if str[len(str)-1:len(str)] == "/" {
		str = string(str[:len(str)-1])
	}
    return str
}