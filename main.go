package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/blackss2/utility/convert"
	"github.com/blackss2/xlsx"
	"go/parser"
	"go/token"
	"time"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	list, err := getTotalImportList("./")
	if err != nil {
		panic(err)
	}

	excel := xlsx.NewFile()
	sheetLicense, _ := excel.AddSheet("License")
	sheetNotFound, _ := excel.AddSheet("Not Found")
	goPath := os.Getenv("GOPATH")
	for _, v := range list {
		src := v[len(goPath)+5:]
		url := fmt.Sprintf("https://libraries.io/go/%s", strings.Replace(src, `\`, "%2F", -1))
		doc, err := goquery.NewDocument(url)
		if err != nil {
			println(err)
			continue
		}

		src = strings.Replace(src, `\`, `/`, -1)
		license := strings.TrimSpace(doc.Find(".col-md-8").Eq(1).Find("p").Eq(4).Text())
		idx := strings.Index(license, "License: ")
		if idx >= 0 {
			r := sheetLicense.AddRow()
			r.AddCell().SetString(src)
			r.AddCell().SetString(license[idx+9:])
		} else {
			r := sheetNotFound.AddRow()
			r.AddCell().SetString(src)
		}
	}
	excel.Save(fmt.Sprintf("./License_%s.xlsx", strings.Replace(strings.Replace(strings.Replace(convert.String(time.Now())[:19], ":", "", -1), "-", "", -1), " ", "", -1)))
}

func getTotalImportList(path string) ([]string, error) {
	importList, err := getImportList(path, nil)
	if err != nil {
		return nil, err
	}
	pathHash := make(map[string]bool)
	for len(importList) > 0 {
		subList := make([]string, 0)
		for _, v := range importList {
			list, err := getImportList(v, pathHash)
			if err != nil {
				return nil, err
			}
			if len(list) > 0 {
				subList = append(subList, list...)
			}
		}
		importList = subList
	}
	totalImportList := make([]string, 0)
	for key, _ := range pathHash {
		totalImportList = append(totalImportList, key)
	}
	convert.String(123)
	return totalImportList, nil
}

func getImportList(path string, pathHash map[string]bool) ([]string, error) {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	importList := make([]string, 0)
	isVisited := false
	if pathHash != nil {
		if _, has := pathHash[path]; has {
			isVisited = true
		}
	}

	if !isVisited {
		if _, err := os.Stat(path); err == nil {
			filepath.Walk(path, func(subpath string, finfo os.FileInfo, err error) error {
				if !finfo.IsDir() {
					ext := filepath.Ext(subpath)
					if ext == ".go" {
						fset := token.NewFileSet()
						f, err := parser.ParseFile(fset, subpath, nil, 0)
						if err != nil {
							panic(err)
						}

						goPath := os.Getenv("GOPATH")
						for _, v := range f.Imports {
							Len := len(v.Path.Value)
							src := fmt.Sprintf("%s/src/%s", goPath, v.Path.Value[1:Len-1])
							importList = append(importList, src)
						}
					}
				}
				return nil
			})

			if pathHash != nil {
				pathHash[path] = true
			}
		}
	}
	return importList, nil
}
