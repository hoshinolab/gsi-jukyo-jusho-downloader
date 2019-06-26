package main

import (
	"archive/zip"
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

var outdir string

func main() {
	// output directory from command line arg
	var (
		out         = flag.String("outdir", "./", "Output destination")
		noDownload  = flag.Bool("nodownload", false, "When you already have zip files, designate this flag. Default is false.")
		noUnzip     = flag.Bool("nounzip", false, "When you do NOT want to extract zip files, designate this flag. Default is false.")
		delTmpFiles = flag.Bool("del", false, "When you want to delete temporary data (zip/csv), designate this flag. Default is false.")
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

	if *noUnzip {
		fmt.Println("NO UNZIP MODE")
	} else {
		// unzip
		files, _ := ioutil.ReadDir(outdir)
		for _, file := range files {
			if strings.Contains(file.Name(), ".zip") {
				zp := path.Join(outdir, file.Name())
				extractCSV(zp)
			}
		}
		fmt.Println("========================")

		//create concat CSV
		ut := time.Now().Unix()
		uts := strconv.FormatInt(ut, 10)
		ccf := uts + "_jukyo-jusho-concat.csv"
		ccp := path.Join(outdir, ccf)
		concatCSV, err := os.OpenFile(ccp, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer concatCSV.Close()

		// list up concat unzipped csv
		files, _ = ioutil.ReadDir(outdir)
		for _, file := range files {
			if strings.Contains(file.Name(), ".csv") && !strings.Contains(file.Name(), "concat") {
				srcbyte := []byte(file.Name())
				rx := regexp.MustCompile(".+_(.+)_(.+).csv")
				match := rx.FindSubmatch(srcbyte)
				prefName := string(match[1])
				cityName := string(match[2])

				//add csv data to concat CSV
				cp := path.Join(outdir, file.Name())
				cf, err := os.Open(cp)
				fmt.Println("open: " + cp)
				if err != nil {
					log.Fatal(err)
					return
				}

				s := bufio.NewScanner(transform.NewReader(cf, japanese.ShiftJIS.NewDecoder()))
				for s.Scan() {
					sl := strings.SplitN(s.Text(), ",", 2)
					writestr := sl[0] + "," + prefName + "," + cityName + "," + sl[1]
					fmt.Fprintln(concatCSV, writestr)
				}
				fmt.Println("close: " + cp)
			}
		}
	}
	if *delTmpFiles {
		fmt.Println("DELETE TMP FILES")
		files, _ := ioutil.ReadDir(outdir)
		for _, file := range files {
			if strings.Contains(file.Name(), ".csv") && !strings.Contains(file.Name(), "concat") || strings.Contains(file.Name(), ".zip") {
				rp := path.Join(outdir, file.Name())
				if err := os.RemoveAll(rp); err != nil {
					fmt.Println(err)
				}
			}
		}
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
		cityName := s.Text()
		if strings.Contains(url, ".zip") {
			fn := getZIPFileName(url)
			num := strings.Replace(fn, ".zip", "", -1)
			fu := getZIPFileURL(fn)
			fmt.Println(prefName + cityName + fu)
			downloadAndWait(num+"_"+prefName+"_"+cityName+".zip", fu)
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

func extractCSV(src string) error {
	srcbyte := []byte(src)
	rx := regexp.MustCompile("\\d+_(.+)_(.+).zip")
	match := rx.FindSubmatch(srcbyte)
	prefName := string(match[1])
	cityName := string(match[2])

	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		fn := f.FileInfo().Name()
		if strings.Contains(fn, ".csv") {
			fmt.Println(fn)
			rf, err := f.Open()
			if err != nil {
				return err
			}
			defer rf.Close()

			buf := make([]byte, f.UncompressedSize)
			_, err = io.ReadFull(rf, buf)
			if err != nil {
				return err
			}

			savefn := strings.Replace(fn, ".csv", "_"+prefName+"_"+cityName+".csv", -1)
			path := filepath.Join(outdir, savefn)
			err = ioutil.WriteFile(path, buf, f.Mode())
			if err != nil {
				return err
			}

		}
	}
	return nil
}
