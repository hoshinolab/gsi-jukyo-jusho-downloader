package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var outdir string

func main() {
	// output directory from command line arg
	var (
		out        = flag.String("outdir", "./", "Output destination")
		noDownload = flag.Bool("nodownload", false, "When you already have zip files, designate this flag. Default is false.")
		noUnzip    = flag.Bool("nounzip", false, "When you do NOT want to extract zip files, designate this flag. Default is false.")
	)
	flag.Parse()

	fmt.Println("Output destination directory: " + *out)
	outdir = *out
	if *noDownload {
		fmt.Println("NO DOWNLOAD MODE")
	} else {
		gsiDownloadURL := "https://saigai.gsi.go.jp/jusho/download/"

		doc, err := goquery.NewDocument(gsiDownloadURL)
		if err != nil {
			fmt.Print("Connection Error")
		}

		doc.Find("a").Each(func(_ int, s *goquery.Selection) {
			url, _ := s.Attr("href")
			text := s.Text()
			if strings.Contains(url, "pref/") {
				getPref(gsiDownloadURL+"/"+url, text)
				time.Sleep(2000 * time.Millisecond)
			}
		})
	}
}

// get list of zip files from prefecture page
func getPref(url string, prefName string) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		fmt.Print("Connection Error")
	}
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		url, _ := s.Attr("href")
		text := s.Text()
		if strings.Contains(url, ".zip") {
			fn := getZIPFileName(url)
			num := strings.Replace(fn, ".zip", "", -1)
			fu := getZIPFileURL(fn)
			fmt.Println(prefName + text + fu)
			downloadAndWait(num+"_"+prefName+text+".zip", fu)
		}
	})
}

// get a zip file name from path like "../data/03482.zip"
func getZIPFileName(path string) string {
	s := strings.Split(path, "/")
	return s[2]
}

// get a zip file URL from file name like "03482.zip"
func getZIPFileURL(filename string) string {
	return "https://saigai.gsi.go.jp/jusho/download/data/" + filename
}

func downloadAndWait(saveFileName string, url string) error {
	res, err := http.Get(url)

	//error
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// create file for the download target
	out, err := os.Create(outdir + "/" + saveFileName)
	if err != nil {
		return err
	}
	defer out.Close()

	// write response body to the file
	_, err = io.Copy(out, res.Body)

	fmt.Println("Done. Waiting")
	time.Sleep(2000 * time.Millisecond)
	return err
}
