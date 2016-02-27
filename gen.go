// Generate SDK from Examle Doc site

// +build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"
)

var (
	docIndex = flag.String("doc", "./api/index.html", "Original doc from")
	apis     []*API
)

type Elem struct {
	Name     string
	Type     string
	Doc      string
	Required bool
}

const apiRequestTpl = `

type {{.Name}} struct {
	{{range $elem := .Parameters}}	
		{{ $elem.Name }} {{ $elem | convertElemType}} //{{$elem.Doc}}
	{{end}}
}

func (r *{{.Name}}) GenURL() (v *url.Values) {
	setAuth(v)
	{{range $elem := .Parameters}}
		{{if eq $elem.Type "Integer" }}
			v.Set("{{ $elem.Name }}", intToString(r.{{$elem.Name}}))
		{{else}}
			v.Set("{{ $elem.Name }}", r.{{$elem.Name}})
		{{end}}
	{{end}}
	v.Set("Action", "{{.Name}}")
	return v
}

func (r *{{.Name}}) Do() (response *{{.Name}}Response, err error) {
	v := r.GenURL()
	resp, err := http.PostForm(Endpoint, *v)
	if err != nil {
		return
	}
	dec := json.NewDecoder(resp.Body)
	dec.Decode(response)
	return
}

type {{ .Name}}Response struct {
	{{range $elem := .ResponseElements}}	
		{{ $elem.Name }} {{ $elem | convertElemType}} //{{$elem.Doc}}
	{{end}}
}

{{if .ExtStructs}}
// Extra struct

{{range $key, $elem := .ExtStructs}}
type {{ $key }}	struct {
	{{range $i := $elem}}	
		{{ $i.Name }} {{ $i | convertElemType}} //{{$i.Doc}}
	{{end}}
}

{{end}}
{{end}}
`

/*
TODO
Default Value
Validation
more type
*/

type API struct {
	Name             string
	URL              string
	Doc              string
	Parameters       []*Elem
	ResponseElements []*Elem
	ExtStructs       map[string][]*Elem
}

func extractAPI(i int, sel *goquery.Selection) {
	api := &API{}
	doc := sel.Children()
	api.Doc, _ = doc.Html()
	href := doc.Next().Children()
	api.URL, _ = href.Attr("href")
	api.Name, _ = href.Html()
	apis = append(apis, api)
}

func genRequest(api *API, doc *goquery.Document) {
	// Request start
	requestNodes := doc.Find("#request tbody tr")
	requestParams := make([]*Elem, 0)
	requestNodes.Each(func(i int, sel *goquery.Selection) {
		p := &Elem{}
		sel = sel.Children()

		for i := 0; i < 4; i++ {
			s, err := sel.Html()
			if err != nil {
				log.Fatal(err)
			}
			switch i {
			case 0:
				p.Name = s
			case 1:
				p.Type = s
			case 2:
				p.Doc = s
			case 3:
				if s == "Yes" {
					p.Required = true
				} else {
					p.Required = false
				}
			}
			sel = sel.Next()
		}
		log.Print("Request:", p)
		requestParams = append(requestParams, p)
	})
	api.Parameters = requestParams

}

func genResponse(api *API, doc *goquery.Document) bool {
	// Response
	responseNodes := doc.Find("#response tbody tr")
	responseElements := make([]*Elem, 0)
	hasExt := false
	responseNodes.Each(func(i int, sel *goquery.Selection) {
		p := &Elem{}
		sel = sel.Children()

		for i := 0; i < 3; i++ {
			s, err := sel.Html()
			if err != nil {
				log.Fatal(err)
			}
			switch i {
			case 0:
				p.Name = s
			case 1:
				if s == "Array" {
					hasExt = true
				}
				p.Type = s
			case 2:
				p.Doc = s
			}
			sel = sel.Next()
		}
		log.Print("Request:", p)
		responseElements = append(responseElements, p)
	})
	api.ResponseElements = responseElements
	return hasExt
}

func genExtra(api *API, doc *goquery.Document) {
	// Extra
	extra := doc.Find(".extra")
	api.ExtStructs = make(map[string][]*Elem)
	extra.Each(func(i int, sel *goquery.Selection) {
		log.Print(i, sel)
		name, ok := sel.Attr("id")
		if !ok {
			return
		}
		Elements := make([]*Elem, 0)
		sel.Find("tbody tr").Each(func(i int, sel *goquery.Selection) {
			p := &Elem{}
			sel = sel.Children()

			for i := 0; i < 3; i++ {
				s, err := sel.Html()
				if err != nil {
					log.Fatal(err)
				}
				switch i {
				case 0:
					p.Name = s
				case 1:
					p.Type = s
				case 2:
					p.Doc = s
				}
				sel = sel.Next()
			}
			log.Print("Extra:", p)
			Elements = append(Elements, p)
		})
		api.ExtStructs[name] = Elements
	})
}

func genAPI(api *API) (err error) {
	log.Print("Generating ", api.Name)
	buf := bytes.NewBuffer([]byte{})
	buf.WriteString(header)

	file, err := os.Open(path.Join(path.Dir(*docIndex), strings.Replace(api.URL, "/api", "", 1)))
	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		log.Fatal(err)
	}
	genRequest(api, doc)
	extra := genResponse(api, doc)
	if extra {
		genExtra(api, doc)
	}

	t := template.New("request")
	maps := make(template.FuncMap)
	maps["convertElemType"] = convertElemType
	t, _ = t.Funcs(maps).Parse(apiRequestTpl)
	tmp := bytes.NewBuffer([]byte{})
	log.Print(t.Execute(tmp, api))
	buf.ReadFrom(tmp)

	// Format
	formated, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile(fmt.Sprintf("%s_gen.go", strings.ToLower(api.Name)), formated, 0644)
	return
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("doc gen: ")
	flag.Parse()

	file, err := os.Open(*docIndex)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	doc, _ := goquery.NewDocumentFromReader(file)
	doc.Find("#api tbody tr").Each(extractAPI)
	log.Printf("Total %d api", len(apis))
	for _, a := range apis {
		genAPI(a)
	}

}

const header = `
// DO NOT EDIT
// This file is automatically generated by gen.go 
// go run gen.go 
// 

package xc
import (
	"net/url"
	"encoding/json"
	"net/http"
)
`

func convertElemType(p *Elem) (dst string) {

	switch p.Type {
	case "String":
		dst = "string"
	case "Integer":
		dst = "int"
	case "Array":
		dst = fmt.Sprintf("[]*%s", p.Name) // Struct will form that
	}
	return
}
