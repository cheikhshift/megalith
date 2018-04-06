package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cheikhshift/db"
	"github.com/cheikhshift/gos/core"
	gosweb "github.com/cheikhshift/gos/web"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/fatih/color"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/opentracing/opentracing-go"
	"html"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"sourcegraph.com/sourcegraph/appdash"
	appdashot "sourcegraph.com/sourcegraph/appdash/opentracing"
	"sourcegraph.com/sourcegraph/appdash/traceapp"
	"strings"
	"sync"
	"time"
)

var store = sessions.NewCookieStore([]byte("a very very very very secret key"))

var Prod = false

var TemplateFuncStore template.FuncMap
var templateCache = gosweb.NewTemplateCache()

func StoreNetfn() int {
	TemplateFuncStore = template.FuncMap{"a": gosweb.Netadd, "s": gosweb.Netsubs, "m": gosweb.Netmultiply, "d": gosweb.Netdivided, "js": gosweb.Netimportjs, "css": gosweb.Netimportcss, "sd": gosweb.NetsessionDelete, "sr": gosweb.NetsessionRemove, "sc": gosweb.NetsessionKey, "ss": gosweb.NetsessionSet, "sso": gosweb.NetsessionSetInt, "sgo": gosweb.NetsessionGetInt, "sg": gosweb.NetsessionGet, "form": gosweb.Formval, "eq": gosweb.Equalz, "neq": gosweb.Nequalz, "lte": gosweb.Netlt, "LoadWebAsset": NetLoadWebAsset, "Mega": NetMega, "AddServer": NetAddServer, "DServer": NetDServer, "UServer": NetUServer, "AddContact": NetAddContact, "GetLog": NetGetLog, "DContact": NetDContact, "UContact": NetUContact, "UMail": NetUMail, "UTw": NetUTw, "USetting": NetUSetting, "ProcessServer": NetProcessServer, "UpdateServer": NetUpdateServer, "RegisterServer": NetRegisterServer, "UpdateKubernetes": NetUpdateKubernetes, "AddPod": NetAddPod, "UpdatePod": NetUpdatePod, "GetPods": NetGetPods, "ang": Netang, "bang": Netbang, "cang": Netcang, "server": Netserver, "bserver": Netbserver, "cserver": Netcserver, "jquery": Netjquery, "bjquery": Netbjquery, "cjquery": Netcjquery}
	return 0
}

var FuncStored = StoreNetfn()

type dbflf db.O

func renderTemplate(w http.ResponseWriter, p *gosweb.Page, span opentracing.Span) {
	defer func() {
		if n := recover(); n != nil {
			color.Red(fmt.Sprintf("Error loading template in path : web%s.tmpl reason : %s", p.R.URL.Path, n))

			DebugTemplate(w, p.R, fmt.Sprintf("web%s", p.R.URL.Path))
			w.WriteHeader(http.StatusInternalServerError)

			pag, err := loadPage("/your-500-page")

			if err != nil {
				log.Println(err.Error())
				return
			}

			if pag.IsResource {
				w.Write(pag.Body)
			} else {
				pag.R = p.R
				pag.Session = p.Session
				renderTemplate(w, pag, span) ///your-500-page"

			}
		}
	}()

	var sp opentracing.Span
	opName := fmt.Sprintf("Building template %s%s", p.R.URL.Path, ".tmpl")

	if true {
		carrier := opentracing.HTTPHeadersCarrier(p.R.Header)
		wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
		if err != nil {
			sp = opentracing.StartSpan(opName)
		} else {
			sp = opentracing.StartSpan(opName, opentracing.ChildOf(wireContext))
		}
	}
	defer sp.Finish()

	// TemplateFuncStore

	if _, ok := templateCache.Get(p.R.URL.Path); !ok || !Prod {
		var tmpstr = string(p.Body)
		var localtemplate = template.New(p.R.URL.Path)

		localtemplate.Funcs(TemplateFuncStore)
		localtemplate.Parse(tmpstr)
		templateCache.Put(p.R.URL.Path, localtemplate)
	}

	outp := new(bytes.Buffer)
	err := templateCache.JGet(p.R.URL.Path).Execute(outp, p)
	if err != nil {
		log.Println(err.Error())
		DebugTemplate(w, p.R, fmt.Sprintf("web%s", p.R.URL.Path))
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "text/html")
		pag, err := loadPage("/your-500-page")

		if err != nil {
			log.Println(err.Error())
			return
		}
		pag.R = p.R
		pag.Session = p.Session

		if pag.IsResource {
			w.Write(pag.Body)
		} else {
			renderTemplate(w, pag, span) // "/your-500-page"

		}
		return
	}

	// p.Session.Save(p.R, w)

	var outps = outp.String()
	var outpescaped = html.UnescapeString(outps)
	outp = nil
	fmt.Fprintf(w, outpescaped)

}

// Access you .gxml's end tags with
// this http.HandlerFunc.
// Use MakeHandler(http.HandlerFunc) to serve your web
// directory from memory.
func MakeHandler(fn func(http.ResponseWriter, *http.Request, opentracing.Span)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		span := opentracing.StartSpan(fmt.Sprintf("%s %s", r.Method, r.URL.Path))
		defer span.Finish()
		carrier := opentracing.HTTPHeadersCarrier(r.Header)
		if err := span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier); err != nil {
			log.Fatalf("Could not inject span context into header: %v", err)
		}

		if attmpt := apiAttempt(w, r, span); !attmpt {
			fn(w, r, span)
		}
		context.Clear(r)

	}
}

func mResponse(v interface{}) string {
	data, _ := json.Marshal(&v)
	return string(data)
}
func apiAttempt(w http.ResponseWriter, r *http.Request, span opentracing.Span) (callmet bool) {
	var response string
	response = ""
	var session *sessions.Session
	var er error
	if session, er = store.Get(r, "session-"); er != nil {
		session, _ = store.New(r, "session-")
	}

	if strings.Contains(r.URL.Path, "/") {

		lastLine := ""
		var sp opentracing.Span
		opName := fmt.Sprintf(" []%s %s", r.Method, r.URL.Path)

		if true {
			carrier := opentracing.HTTPHeadersCarrier(r.Header)
			wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
			if err != nil {
				sp = opentracing.StartSpan(opName)
			} else {
				sp = opentracing.StartSpan(opName, opentracing.ChildOf(wireContext))
			}
		}
		defer sp.Finish()
		defer func() {
			if n := recover(); n != nil {
				log.Println("Web request (/) failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml :", strings.TrimSpace(lastLine))
				log.Println("Reason : ", n)
				//wheredefault
				span.SetTag("error", true)
				span.LogEvent(fmt.Sprintf("%s request at %s, reason : %s ", r.Method, r.URL.Path, n))

				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "text/html")
				pag, err := loadPage("/your-500-page")

				if err != nil {
					log.Println(err.Error())
					callmet = true
					return
				}
				pag.R = r
				pag.Session = session
				if pag.IsResource {
					w.Write(pag.Body)
				} else {
					// renderTemplate(w, pag, span)

				}

				callmet = true
			}
		}()
		lastLine = `if strings.Contains(r.URL.Path, ".map") || strings.Contains(r.URL.Path, "web/{{ server.Image }}") {`
		if strings.Contains(r.URL.Path, ".map") || strings.Contains(r.URL.Path, "web/{{ server.Image }}") {
			lastLine = `return true`
			return true
			lastLine = `}`
		}

	}
	if r.Method == "RESET" {
		return true
	} else if isURL := (r.URL.Path == "/update/server" && r.Method == strings.ToUpper("POST")); !callmet && isURL {

		lastLine := ""
		var sp opentracing.Span
		opName := fmt.Sprintf(" []%s %s", r.Method, r.URL.Path)

		if true {
			carrier := opentracing.HTTPHeadersCarrier(r.Header)
			wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
			if err != nil {
				sp = opentracing.StartSpan(opName)
			} else {
				sp = opentracing.StartSpan(opName, opentracing.ChildOf(wireContext))
			}
		}
		defer sp.Finish()
		defer func() {
			if n := recover(); n != nil {
				log.Println("Web request (/update/server) failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml :", strings.TrimSpace(lastLine))
				log.Println("Reason : ", n)
				//wheredefault
				span.SetTag("error", true)
				span.LogEvent(fmt.Sprintf("%s request at %s, reason : %s ", r.Method, r.URL.Path, n))
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "text/html")
				pag, err := loadPage("/your-500-page")

				if err != nil {
					log.Println(err.Error())
					callmet = true
					return
				}
				pag.R = r
				pag.Session = session
				if pag.IsResource {
					w.Write(pag.Body)
				} else {
					renderTemplate(w, pag, span) //"s"

				}
				callmet = true

			}
		}()

		lastLine = `decoder := json.NewDecoder(r.Body)`
		decoder := json.NewDecoder(r.Body)
		lastLine = `var tmvv PayloadOfRequest`
		var tmvv PayloadOfRequest
		lastLine = `err := decoder.Decode(&tmvv)`
		err := decoder.Decode(&tmvv)
		lastLine = `if err != nil {`
		if err != nil {
			lastLine = `w.WriteHeader(http.StatusInternalServerError)`
			w.WriteHeader(http.StatusInternalServerError)
			lastLine = `w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))`
			w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
			lastLine = `return true`
			return true
			lastLine = `}`
		}
		lastLine = `_ = NetProcessServer(tmvv.req)`
		_ = NetProcessServer(tmvv.req)
		lastLine = `w.Header().Set("Content-Type", "text/plain")`
		w.Header().Set("Content-Type", "text/plain")
		lastLine = `w.Write(OK)`
		w.Write(OK)
		callmet = true
	} else if isURL := (r.URL.Path == "/mega" && r.Method == strings.ToUpper("POST")); !callmet && isURL {

		lastLine := ""
		var sp opentracing.Span
		opName := fmt.Sprintf(" []%s %s", r.Method, r.URL.Path)

		if true {
			carrier := opentracing.HTTPHeadersCarrier(r.Header)
			wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
			if err != nil {
				sp = opentracing.StartSpan(opName)
			} else {
				sp = opentracing.StartSpan(opName, opentracing.ChildOf(wireContext))
			}
		}
		defer sp.Finish()
		defer func() {
			if n := recover(); n != nil {
				log.Println("Web request (/mega) failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml :", strings.TrimSpace(lastLine))
				log.Println("Reason : ", n)
				//wheredefault
				span.SetTag("error", true)
				span.LogEvent(fmt.Sprintf("%s request at %s, reason : %s ", r.Method, r.URL.Path, n))
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "text/html")
				pag, err := loadPage("/your-500-page")

				if err != nil {
					log.Println(err.Error())
					callmet = true
					return
				}
				pag.R = r
				pag.Session = session
				if pag.IsResource {
					w.Write(pag.Body)
				} else {
					renderTemplate(w, pag, span) //"s"

				}
				callmet = true

			}
		}()

		lastLine = `Cfg := &MegaConfig{}`
		Cfg := &MegaConfig{}
		lastLine = `LoadConfig(&Config)`
		LoadConfig(&Config)
		lastLine = `w.Header().Set("Content-Type", "application/json")`
		w.Header().Set("Content-Type", "application/json")
		lastLine = `retjson := []byte(mResponse(Cfg))`
		retjson := []byte(mResponse(Cfg))
		lastLine = `w.Write(retjson)`
		w.Write(retjson)
		lastLine = `retjson = nil`
		retjson = nil
		callmet = true
	}

	if callmet {
		session.Save(r, w)
		session = nil
		if response != "" {
			//Unmarshal json
			//w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(response))
		}
		return
	}
	session = nil
	return
}

func DebugTemplate(w http.ResponseWriter, r *http.Request, tmpl string) {
	lastline := 0
	linestring := ""
	defer func() {
		if n := recover(); n != nil {
			log.Println()
			// log.Println(n)
			log.Println("Error on line :", lastline+1, ":"+strings.TrimSpace(linestring))
			//http.Redirect(w,r,"/your-500-page",307)
		}
	}()

	p, err := loadPage(r.URL.Path)
	filename := tmpl + ".tmpl"
	body, err := Asset(filename)
	session, er := store.Get(r, "session-")

	if er != nil {
		session, er = store.New(r, "session-")
	}
	p.Session = session
	p.R = r
	if err != nil {
		log.Print(err)

	} else {

		lines := strings.Split(string(body), "\n")
		// log.Println( lines )
		linebuffer := ""
		waitend := false
		open := 0
		for i, line := range lines {

			processd := false

			if strings.Contains(line, "{{with") || strings.Contains(line, "{{ with") || strings.Contains(line, "with}}") || strings.Contains(line, "with }}") || strings.Contains(line, "{{range") || strings.Contains(line, "{{ range") || strings.Contains(line, "range }}") || strings.Contains(line, "range}}") || strings.Contains(line, "{{if") || strings.Contains(line, "{{ if") || strings.Contains(line, "if }}") || strings.Contains(line, "if}}") || strings.Contains(line, "{{block") || strings.Contains(line, "{{ block") || strings.Contains(line, "block }}") || strings.Contains(line, "block}}") {
				linebuffer += line
				waitend = true

				endstr := ""
				processd = true
				if !(strings.Contains(line, "{{end") || strings.Contains(line, "{{ end") || strings.Contains(line, "end}}") || strings.Contains(line, "end }}")) {

					open++

				}
				for i := 0; i < open; i++ {
					endstr += "\n{{end}}"
				}
				//exec
				outp := new(bytes.Buffer)
				t := template.New("PageWrapper")
				t = t.Funcs(TemplateFuncStore)
				t, _ = t.Parse(string(body))
				lastline = i
				linestring = line
				erro := t.Execute(outp, p)
				if erro != nil {
					log.Println("Error on line :", i+1, line, erro.Error())
				}
			}

			if waitend && !processd && !(strings.Contains(line, "{{end") || strings.Contains(line, "{{ end")) {
				linebuffer += line

				endstr := ""
				for i := 0; i < open; i++ {
					endstr += "\n{{end}}"
				}
				//exec
				outp := new(bytes.Buffer)
				t := template.New("PageWrapper")
				t = t.Funcs(TemplateFuncStore)
				t, _ = t.Parse(string(body))
				lastline = i
				linestring = line
				erro := t.Execute(outp, p)
				if erro != nil {
					log.Println("Error on line :", i+1, line, erro.Error())
				}

			}

			if !waitend && !processd {
				outp := new(bytes.Buffer)
				t := template.New("PageWrapper")
				t = t.Funcs(TemplateFuncStore)
				t, _ = t.Parse(string(body))
				lastline = i
				linestring = line
				erro := t.Execute(outp, p)
				if erro != nil {
					log.Println("Error on line :", i+1, line, erro.Error())
				}
			}

			if !processd && (strings.Contains(line, "{{end") || strings.Contains(line, "{{ end")) {
				open--

				if open == 0 {
					waitend = false

				}
			}
		}

	}

}

func DebugTemplatePath(tmpl string, intrf interface{}) {
	lastline := 0
	linestring := ""
	defer func() {
		if n := recover(); n != nil {

			log.Println("Error on line :", lastline+1, ":"+strings.TrimSpace(linestring))
			log.Println(n)
			//http.Redirect(w,r,"/your-500-page",307)
		}
	}()

	filename := tmpl
	body, err := Asset(filename)

	if err != nil {
		log.Print(err)

	} else {

		lines := strings.Split(string(body), "\n")
		// log.Println( lines )
		linebuffer := ""
		waitend := false
		open := 0
		for i, line := range lines {

			processd := false

			if strings.Contains(line, "{{with") || strings.Contains(line, "{{ with") || strings.Contains(line, "with}}") || strings.Contains(line, "with }}") || strings.Contains(line, "{{range") || strings.Contains(line, "{{ range") || strings.Contains(line, "range }}") || strings.Contains(line, "range}}") || strings.Contains(line, "{{if") || strings.Contains(line, "{{ if") || strings.Contains(line, "if }}") || strings.Contains(line, "if}}") || strings.Contains(line, "{{block") || strings.Contains(line, "{{ block") || strings.Contains(line, "block }}") || strings.Contains(line, "block}}") {
				linebuffer += line
				waitend = true

				endstr := ""
				if !(strings.Contains(line, "{{end") || strings.Contains(line, "{{ end") || strings.Contains(line, "end}}") || strings.Contains(line, "end }}")) {

					open++

				}

				for i := 0; i < open; i++ {
					endstr += "\n{{end}}"
				}
				//exec

				processd = true
				outp := new(bytes.Buffer)
				t := template.New("PageWrapper")
				t = t.Funcs(TemplateFuncStore)
				t, _ = t.Parse(string([]byte(fmt.Sprintf("%s%s", linebuffer, endstr))))
				lastline = i
				linestring = line
				erro := t.Execute(outp, intrf)
				if erro != nil {
					log.Println("Error on line :", i+1, line, erro.Error())
				}
			}

			if waitend && !processd && !(strings.Contains(line, "{{end") || strings.Contains(line, "{{ end") || strings.Contains(line, "end}}") || strings.Contains(line, "end }}")) {
				linebuffer += line

				endstr := ""
				for i := 0; i < open; i++ {
					endstr += "\n{{end}}"
				}
				//exec
				outp := new(bytes.Buffer)
				t := template.New("PageWrapper")
				t = t.Funcs(TemplateFuncStore)
				t, _ = t.Parse(string([]byte(fmt.Sprintf("%s%s", linebuffer, endstr))))
				lastline = i
				linestring = line
				erro := t.Execute(outp, intrf)
				if erro != nil {
					log.Println("Error on line :", i+1, line, erro.Error())
				}

			}

			if !waitend && !processd {
				outp := new(bytes.Buffer)
				t := template.New("PageWrapper")
				t = t.Funcs(TemplateFuncStore)
				t, _ = t.Parse(string([]byte(fmt.Sprintf("%s%s", linebuffer))))
				lastline = i
				linestring = line
				erro := t.Execute(outp, intrf)
				if erro != nil {
					log.Println("Error on line :", i+1, line, erro.Error())
				}
			}

			if !processd && (strings.Contains(line, "{{end") || strings.Contains(line, "{{ end") || strings.Contains(line, "end}}") || strings.Contains(line, "end }}")) {
				open--

				if open == 0 {
					waitend = false

				}
			}
		}

	}

}
func Handler(w http.ResponseWriter, r *http.Request, span opentracing.Span) {
	var p *gosweb.Page
	p, err := loadPage(r.URL.Path)
	var session *sessions.Session
	var er error
	if session, er = store.Get(r, "session-"); er != nil {
		session, _ = store.New(r, "session-")
	}

	var sp opentracing.Span
	opName := fmt.Sprintf(fmt.Sprintf("Web:/%s", r.URL.Path))

	if true {
		carrier := opentracing.HTTPHeadersCarrier(r.Header)
		wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
		if err != nil {
			sp = opentracing.StartSpan(opName)
		} else {
			sp = opentracing.StartSpan(opName, opentracing.ChildOf(wireContext))
		}
	}
	defer sp.Finish()

	if err != nil {
		log.Println(err.Error())

		w.WriteHeader(http.StatusNotFound)
		span.SetTag("error", true)
		span.LogEvent(fmt.Sprintf("%s request at %s, reason : %s ", r.Method, r.URL.Path, err))
		pag, err := loadPage("/your-404-page")

		if err != nil {
			log.Println(err.Error())
			//context.Clear(r)
			return
		}
		pag.R = r
		pag.Session = session
		if p != nil {
			p.Session = nil
			p.Body = nil
			p.R = nil
			p = nil
		}

		if pag.IsResource {
			w.Write(pag.Body)
		} else {
			renderTemplate(w, pag, span) //"/your-500-page"
		}
		session = nil
		context.Clear(r)
		return
	}

	if !p.IsResource {
		w.Header().Set("Content-Type", "text/html")
		p.Session = session
		p.R = r
		renderTemplate(w, p, span) //fmt.Sprintf("web%s", r.URL.Path)
		session.Save(r, w)
		// log.Println(w)
	} else {
		w.Header().Set("Cache-Control", "public")
		if strings.Contains(r.URL.Path, ".css") {
			w.Header().Add("Content-Type", "text/css")
		} else if strings.Contains(r.URL.Path, ".js") {
			w.Header().Add("Content-Type", "application/javascript")
		} else {
			w.Header().Add("Content-Type", http.DetectContentType(p.Body))
		}

		w.Write(p.Body)
	}

	p.Session = nil
	p.Body = nil
	p.R = nil
	p = nil
	session = nil
	context.Clear(r)
	return
}

var WebCache = gosweb.NewCache()

func loadPage(title string) (*gosweb.Page, error) {

	var nPage = gosweb.Page{}
	if roottitle := (title == "/"); roottitle {
		webbase := "web/"
		fname := fmt.Sprintf("%s%s", webbase, "index.html")
		body, err := Asset(fname)
		if err != nil {
			fname = fmt.Sprintf("%s%s", webbase, "index.tmpl")
			body, err = Asset(fname)
			if err != nil {
				return nil, err
			}
			nPage.Body = body
			WebCache.Put(title, nPage)
			body = nil
			return &nPage, nil
		}
		nPage.Body = body
		nPage.IsResource = true
		WebCache.Put(title, nPage)
		body = nil
		return &nPage, nil

	}

	filename := fmt.Sprintf("web%s.tmpl", title)

	if body, err := Asset(filename); err != nil {
		filename = fmt.Sprintf("web%s.html", title)

		if body, err = Asset(filename); err != nil {
			filename = fmt.Sprintf("web%s", title)

			if body, err = Asset(filename); err != nil {
				return nil, err
			} else {
				if strings.Contains(title, ".tmpl") {
					return nil, nil
				}
				nPage.Body = body
				nPage.IsResource = true
				WebCache.Put(title, nPage)
				body = nil
				return &nPage, nil
			}
		} else {
			nPage.Body = body
			nPage.IsResource = true
			WebCache.Put(title, nPage)
			body = nil
			return &nPage, nil
		}
	} else {
		nPage.Body = body
		WebCache.Put(title, nPage)
		body = nil
		return &nPage, nil
	}

	//wheredefault

}

var Config *MegaConfig
var GL TrLock
var isInContainer bool

func init() {

}

//
func NetLoadWebAsset(args ...interface{}) string {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `data,err := Asset( fmt.Sprintf("web%s", args[0].(string) ) )`
	data, err := Asset(fmt.Sprintf("web%s", args[0].(string)))
	lastLine = `if err != nil {`
	if err != nil {
		lastLine = `return err.Error()`
		return err.Error()
		lastLine = `}`
	}
	lastLine = `return string(data)`
	return string(data)
}

//
func NetMega() (result *MegaConfig) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `if WorkerAddressPort != DefaultAddress {`
	if WorkerAddressPort != DefaultAddress {
		lastLine = `LoadConfig(&Config)`
		LoadConfig(&Config)
		lastLine = `}`
	}
	lastLine = `ClearHistory()`
	ClearHistory()
	lastLine = `ShouldLock()`
	ShouldLock()
	lastLine = `defer ShouldUnlock()`
	defer ShouldUnlock()
	lastLine = `return Config`
	return Config
}

//
func NetAddServer() (result []Server) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `ShouldLock()`
	ShouldLock()
	lastLine = `randint := rand.Intn(200) + 50 + len(Config.Servers)`
	randint := rand.Intn(200) + 50 + len(Config.Servers)
	lastLine = `genimage := fmt.Sprintf("https://picsum.photos/%v/%v",randint, randint)`
	genimage := fmt.Sprintf("https://picsum.photos/%v/%v", randint, randint)
	lastLine = `ns := Server{ID : core.NewLen(20), Nickname:"New server",Image : genimage}`
	ns := Server{ID: core.NewLen(20), Nickname: "New server", Image: genimage}
	lastLine = `Config.Servers = append(Config.Servers, ns)`
	Config.Servers = append(Config.Servers, ns)
	lastLine = `if DispatcherAddressPort != DefaultAddress {`
	if DispatcherAddressPort != DefaultAddress {
		lastLine = `ioutil.WriteFile(fmt.Sprintf(urlformat, GenLogName(ns.ID), LockExt), OK ,0700)`
		ioutil.WriteFile(fmt.Sprintf(urlformat, GenLogName(ns.ID), LockExt), OK, 0700)
		lastLine = `}`
	}
	lastLine = `ShouldUnlock()`
	ShouldUnlock()
	lastLine = `SaveConfig(&Config);`
	SaveConfig(&Config)
	lastLine = `return Config.Servers`
	return Config.Servers
}

//
func NetDServer(req Server) (result []Server) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `result = []Server{}`
	result = []Server{}
	lastLine = `ShouldLock()`
	ShouldLock()
	lastLine = `for _,target := range Config.Servers {`
	for _, target := range Config.Servers {
		lastLine = `if target.ID != req.ID {`
		if target.ID != req.ID {
			lastLine = `result = append(result, target)`
			result = append(result, target)
			lastLine = `}`
		}
		lastLine = `}`
	}
	lastLine = `DeleteLog(req.ID)`
	DeleteLog(req.ID)
	lastLine = `Config.Servers = result`
	Config.Servers = result
	lastLine = `ShouldUnlock()`
	ShouldUnlock()
	lastLine = `SaveConfig(&Config);`
	SaveConfig(&Config)
	lastLine = `return`
	return
}

//
func NetUServer(req Server) (result bool) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `ShouldLock()`
	ShouldLock()
	lastLine = `for index,target := range Config.Servers {`
	for index, target := range Config.Servers {
		lastLine = `if target.ID == req.ID {`
		if target.ID == req.ID {
			lastLine = `Config.Servers[index] = req`
			Config.Servers[index] = req
			lastLine = `}`
		}
		lastLine = `}`
	}
	lastLine = `ShouldUnlock()`
	ShouldUnlock()
	lastLine = `SaveConfig(&Config);`
	SaveConfig(&Config)
	lastLine = `return true`
	return true
}

//
func NetAddContact() (result []Contact) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `ShouldLock()`
	ShouldLock()
	lastLine = `nc := Contact{ID : core.NewLen(20), Nickname:"New contact"}`
	nc := Contact{ID: core.NewLen(20), Nickname: "New contact"}
	lastLine = `Config.Contacts = append(Config.Contacts, nc)`
	Config.Contacts = append(Config.Contacts, nc)
	lastLine = `ShouldUnlock()`
	ShouldUnlock()
	lastLine = `SaveConfig(&Config);`
	SaveConfig(&Config)
	lastLine = `return Config.Contacts`
	return Config.Contacts
}

//
func NetGetLog(req Server) (result RequestLog) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `ShouldLock()`
	ShouldLock()
	lastLine = `LoadLog(req.ID, &result)`
	LoadLog(req.ID, &result)
	lastLine = `ShouldUnlock()`
	ShouldUnlock()
	lastLine = `return`
	return
}

//
func NetDContact(req Contact) (result []Contact) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `result = []Contact{}`
	result = []Contact{}
	lastLine = `ShouldLock()`
	ShouldLock()
	lastLine = `for _,target := range Config.Contacts {`
	for _, target := range Config.Contacts {
		lastLine = `if target.ID != req.ID {`
		if target.ID != req.ID {
			lastLine = `result = append(result, target)`
			result = append(result, target)
			lastLine = `}`
		}
		lastLine = `}`
	}
	lastLine = `Config.Contacts = result`
	Config.Contacts = result
	lastLine = `ShouldUnlock()`
	ShouldUnlock()
	lastLine = `SaveConfig(&Config);`
	SaveConfig(&Config)
	lastLine = `return`
	return
}

//
func NetUContact(req Contact) (result bool) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `ShouldLock()`
	ShouldLock()
	lastLine = `for index,target := range Config.Contacts {`
	for index, target := range Config.Contacts {
		lastLine = `if target.ID == req.ID {`
		if target.ID == req.ID {
			lastLine = `Config.Contacts[index] = req`
			Config.Contacts[index] = req
			lastLine = `}`
		}
		lastLine = `}`
	}
	lastLine = `ShouldUnlock()`
	ShouldUnlock()
	lastLine = `SaveConfig(&Config);`
	SaveConfig(&Config)
	lastLine = `return true`
	return true
}

//
func NetUMail(req MailSettings) (result bool) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `ShouldLock()`
	ShouldLock()
	lastLine = `Config.Mail = req`
	Config.Mail = req
	lastLine = `ShouldUnlock()`
	ShouldUnlock()
	lastLine = `SaveConfig(&Config);`
	SaveConfig(&Config)
	lastLine = `return true`
	return true
}

//
func NetUTw(req TwilioInfo) (result bool) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `ShouldLock()`
	ShouldLock()
	lastLine = `Config.SMS = req`
	Config.SMS = req
	lastLine = `ShouldUnlock()`
	ShouldUnlock()
	lastLine = `SaveConfig(&Config);`
	SaveConfig(&Config)
	lastLine = `return true`
	return true
}

//
func NetUSetting(req Settings) (result bool) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `ShouldLock()`
	ShouldLock()
	lastLine = `Config.Misc = req`
	Config.Misc = req
	lastLine = `Config.LastReset = time.Now().Unix()`
	Config.LastReset = time.Now().Unix()
	lastLine = `ShouldUnlock()`
	ShouldUnlock()
	lastLine = `SaveConfig(&Config);`
	SaveConfig(&Config)
	lastLine = `return true`
	return true
}

//
func NetProcessServer(req string) (result bool) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `LoadConfig(&Config)`
	LoadConfig(&Config)
	lastLine = `server,index := FindServer(req)`
	server, index := FindServer(req)
	lastLine = `Process(server,index)`
	Process(server, index)
	lastLine = `return true`
	return true
}

//
func NetUpdateServer(req Server) (result bool) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `ShouldLock()`
	ShouldLock()
	lastLine = `_,index := FindServer(req.ID)`
	_, index := FindServer(req.ID)
	lastLine = `Config.Servers[index].Uptime = req.Uptime`
	Config.Servers[index].Uptime = req.Uptime
	lastLine = `ShouldUnlock()`
	ShouldUnlock()
	lastLine = `SaveConfig(&Config)`
	SaveConfig(&Config)
	lastLine = `return true`
	return true
}

//
func NetRegisterServer(req string) (result bool) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `RegisterWorker(req)`
	RegisterWorker(req)
	lastLine = `return true`
	return true
}

//
func NetUpdateKubernetes(req k8sConfig) (result bool) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `ShouldLock()`
	ShouldLock()
	lastLine = `Config.KubeSettings = req`
	Config.KubeSettings = req
	lastLine = `ShouldUnlock()`
	ShouldUnlock()
	lastLine = `SaveConfig(&Config);`
	SaveConfig(&Config)
	lastLine = `return true`
	return true
}

//
func NetAddPod(req PodConfig) (watching []PodConfig) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `ShouldLock()`
	ShouldLock()
	lastLine = `Config.KubeSettings.Monitoring = append(Config.KubeSettings.Monitoring, req)`
	Config.KubeSettings.Monitoring = append(Config.KubeSettings.Monitoring, req)
	lastLine = `watching = Config.KubeSettings.Monitoring`
	watching = Config.KubeSettings.Monitoring
	lastLine = `ShouldUnlock()`
	ShouldUnlock()
	lastLine = `SaveConfig(&Config);`
	SaveConfig(&Config)
	lastLine = `return`
	return
}

//
func NetUpdatePod(req PodConfig) (result bool) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `ShouldLock()`
	ShouldLock()
	lastLine = `for index,target := range Config.KubeSettings.Monitoring {`
	for index, target := range Config.KubeSettings.Monitoring {
		lastLine = `if target.Name == req.Name {`
		if target.Name == req.Name {
			lastLine = `Config.KubeSettings.Monitoring[index] = req`
			Config.KubeSettings.Monitoring[index] = req
			lastLine = `}`
		}
		lastLine = `}`
	}
	lastLine = `ShouldUnlock()`
	ShouldUnlock()
	lastLine = `SaveConfig(&Config);`
	SaveConfig(&Config)
	lastLine = `return true`
	return true
}

//
func NetGetPods() (result []Pod) {

	lastLine := ""

	defer func() {
		if n := recover(); n != nil {
			log.Println("Pipeline failed at line :", gosweb.GetLine(".//gos.gxml", lastLine), "Of file:.//gos.gxml:", strings.TrimSpace(lastLine))
			log.Println("Reason : ", n)

		}
	}()
	lastLine = `ClearHistory()`
	ClearHistory()
	lastLine = `list,_ := RequestPods()`
	list, _ := RequestPods()
	lastLine = `return list.Items`
	return list.Items
}

func templateFNang(localid string, d interface{}) {
	if n := recover(); n != nil {
		color.Red(fmt.Sprintf("Error loading template in path (momentum/ang) : %s", localid))
		// log.Println(n)
		DebugTemplatePath(localid, d)
	}
}

var templateIDang = "tmpl/momentum/ang.tmpl"

func Netang(args ...interface{}) string {

	localid := templateIDang
	var d *gosweb.NoStruct
	defer templateFNang(localid, d)
	if len(args) > 0 {
		jso := args[0].(string)
		var jsonBlob = []byte(jso)
		err := json.Unmarshal(jsonBlob, d)
		if err != nil {
			return err.Error()
		}
	} else {
		d = &gosweb.NoStruct{}
	}

	output := new(bytes.Buffer)

	if _, ok := templateCache.Get(localid); !ok || !Prod {

		body, er := Asset(localid)
		if er != nil {
			return ""
		}
		var localtemplate = template.New("ang")
		localtemplate.Funcs(TemplateFuncStore)
		var tmpstr = string(body)
		localtemplate.Parse(tmpstr)
		body = nil
		templateCache.Put(localid, localtemplate)
	}

	erro := templateCache.JGet(localid).Execute(output, d)
	if erro != nil {
		color.Red(fmt.Sprintf("Error processing template %s", localid))
		DebugTemplatePath(localid, d)
	}
	var outps = output.String()
	var outpescaped = html.UnescapeString(outps)
	d = nil
	output.Reset()
	output = nil
	args = nil
	return outpescaped

}
func bang(d gosweb.NoStruct) string {
	return Netbang(d)
}

//
func Netbang(d gosweb.NoStruct) string {
	localid := templateIDang
	defer templateFNang(localid, d)
	output := new(bytes.Buffer)

	if _, ok := templateCache.Get(localid); !ok || !Prod {

		body, er := Asset(localid)
		if er != nil {
			return ""
		}
		var localtemplate = template.New("ang")
		localtemplate.Funcs(TemplateFuncStore)
		var tmpstr = string(body)
		localtemplate.Parse(tmpstr)
		body = nil
		templateCache.Put(localid, localtemplate)
	}

	erro := templateCache.JGet(localid).Execute(output, d)
	if erro != nil {
		log.Println(erro)
	}
	var outps = output.String()
	var outpescaped = html.UnescapeString(outps)
	d = gosweb.NoStruct{}
	output.Reset()
	output = nil
	return outpescaped
}
func Netcang(args ...interface{}) (d gosweb.NoStruct) {
	if len(args) > 0 {
		var jsonBlob = []byte(args[0].(string))
		err := json.Unmarshal(jsonBlob, &d)
		if err != nil {
			log.Println("error:", err)
			return
		}
	} else {
		d = gosweb.NoStruct{}
	}
	return
}

func cang(args ...interface{}) (d gosweb.NoStruct) {
	if len(args) > 0 {
		d = Netcang(args[0])
	} else {
		d = Netcang()
	}
	return
}

func templateFNserver(localid string, d interface{}) {
	if n := recover(); n != nil {
		color.Red(fmt.Sprintf("Error loading template in path (momentum/server) : %s", localid))
		// log.Println(n)
		DebugTemplatePath(localid, d)
	}
}

var templateIDserver = "tmpl/momentum/server.tmpl"

func Netserver(args ...interface{}) string {

	localid := templateIDserver
	var d *gosweb.NoStruct
	defer templateFNserver(localid, d)
	if len(args) > 0 {
		jso := args[0].(string)
		var jsonBlob = []byte(jso)
		err := json.Unmarshal(jsonBlob, d)
		if err != nil {
			return err.Error()
		}
	} else {
		d = &gosweb.NoStruct{}
	}

	output := new(bytes.Buffer)

	if _, ok := templateCache.Get(localid); !ok || !Prod {

		body, er := Asset(localid)
		if er != nil {
			return ""
		}
		var localtemplate = template.New("server")
		localtemplate.Funcs(TemplateFuncStore)
		var tmpstr = string(body)
		localtemplate.Parse(tmpstr)
		body = nil
		templateCache.Put(localid, localtemplate)
	}

	erro := templateCache.JGet(localid).Execute(output, d)
	if erro != nil {
		color.Red(fmt.Sprintf("Error processing template %s", localid))
		DebugTemplatePath(localid, d)
	}
	var outps = output.String()
	var outpescaped = html.UnescapeString(outps)
	d = nil
	output.Reset()
	output = nil
	args = nil
	return outpescaped

}
func bserver(d gosweb.NoStruct) string {
	return Netbserver(d)
}

//
func Netbserver(d gosweb.NoStruct) string {
	localid := templateIDserver
	defer templateFNserver(localid, d)
	output := new(bytes.Buffer)

	if _, ok := templateCache.Get(localid); !ok || !Prod {

		body, er := Asset(localid)
		if er != nil {
			return ""
		}
		var localtemplate = template.New("server")
		localtemplate.Funcs(TemplateFuncStore)
		var tmpstr = string(body)
		localtemplate.Parse(tmpstr)
		body = nil
		templateCache.Put(localid, localtemplate)
	}

	erro := templateCache.JGet(localid).Execute(output, d)
	if erro != nil {
		log.Println(erro)
	}
	var outps = output.String()
	var outpescaped = html.UnescapeString(outps)
	d = gosweb.NoStruct{}
	output.Reset()
	output = nil
	return outpescaped
}
func Netcserver(args ...interface{}) (d gosweb.NoStruct) {
	if len(args) > 0 {
		var jsonBlob = []byte(args[0].(string))
		err := json.Unmarshal(jsonBlob, &d)
		if err != nil {
			log.Println("error:", err)
			return
		}
	} else {
		d = gosweb.NoStruct{}
	}
	return
}

func cserver(args ...interface{}) (d gosweb.NoStruct) {
	if len(args) > 0 {
		d = Netcserver(args[0])
	} else {
		d = Netcserver()
	}
	return
}

func templateFNjquery(localid string, d interface{}) {
	if n := recover(); n != nil {
		color.Red(fmt.Sprintf("Error loading template in path (momentum/jquery) : %s", localid))
		// log.Println(n)
		DebugTemplatePath(localid, d)
	}
}

var templateIDjquery = "tmpl/momentum/jquery.tmpl"

func Netjquery(args ...interface{}) string {

	localid := templateIDjquery
	var d *gosweb.NoStruct
	defer templateFNjquery(localid, d)
	if len(args) > 0 {
		jso := args[0].(string)
		var jsonBlob = []byte(jso)
		err := json.Unmarshal(jsonBlob, d)
		if err != nil {
			return err.Error()
		}
	} else {
		d = &gosweb.NoStruct{}
	}

	output := new(bytes.Buffer)

	if _, ok := templateCache.Get(localid); !ok || !Prod {

		body, er := Asset(localid)
		if er != nil {
			return ""
		}
		var localtemplate = template.New("jquery")
		localtemplate.Funcs(TemplateFuncStore)
		var tmpstr = string(body)
		localtemplate.Parse(tmpstr)
		body = nil
		templateCache.Put(localid, localtemplate)
	}

	erro := templateCache.JGet(localid).Execute(output, d)
	if erro != nil {
		color.Red(fmt.Sprintf("Error processing template %s", localid))
		DebugTemplatePath(localid, d)
	}
	var outps = output.String()
	var outpescaped = html.UnescapeString(outps)
	d = nil
	output.Reset()
	output = nil
	args = nil
	return outpescaped

}
func bjquery(d gosweb.NoStruct) string {
	return Netbjquery(d)
}

//
func Netbjquery(d gosweb.NoStruct) string {
	localid := templateIDjquery
	defer templateFNjquery(localid, d)
	output := new(bytes.Buffer)

	if _, ok := templateCache.Get(localid); !ok || !Prod {

		body, er := Asset(localid)
		if er != nil {
			return ""
		}
		var localtemplate = template.New("jquery")
		localtemplate.Funcs(TemplateFuncStore)
		var tmpstr = string(body)
		localtemplate.Parse(tmpstr)
		body = nil
		templateCache.Put(localid, localtemplate)
	}

	erro := templateCache.JGet(localid).Execute(output, d)
	if erro != nil {
		log.Println(erro)
	}
	var outps = output.String()
	var outpescaped = html.UnescapeString(outps)
	d = gosweb.NoStruct{}
	output.Reset()
	output = nil
	return outpescaped
}
func Netcjquery(args ...interface{}) (d gosweb.NoStruct) {
	if len(args) > 0 {
		var jsonBlob = []byte(args[0].(string))
		err := json.Unmarshal(jsonBlob, &d)
		if err != nil {
			log.Println("error:", err)
			return
		}
	} else {
		d = gosweb.NoStruct{}
	}
	return
}

func cjquery(args ...interface{}) (d gosweb.NoStruct) {
	if len(args) > 0 {
		d = Netcjquery(args[0])
	} else {
		d = Netcjquery()
	}
	return
}

func dummy_timer() {
	dg := time.Second * 5
	log.Println(dg)
}
func main() {
	fmt.Fprintf(os.Stdout, "%v\n", os.Getpid())

	nobrowser := flag.Bool("nobrowser", false, "Launch without openning browser")
	worker := flag.Bool("worker", false, "Launch megalith instance as worker")
	dispaddr := flag.String("dispatcher", DefaultAddress, "Host name of dispatcher instance. Add port number as needed. ie hostname:9000")
	workaddr := flag.String("hostname", DefaultAddress, "Host name of worker instance. Add port number as needed. ie hostname:9000")
	portNumber := flag.String("port", DefaultPort, "The port number megalith will to listen on")
	fws := flag.String("workspace", "megaWorkSpace", "Set instance directory")
	container := flag.Bool("container", false, "Get Dispatcher and hostname addresses from env. variables.")

	flag.Parse()
	WorkerMode = *worker

	if *container == false {
		DispatcherAddressPort = *dispaddr
		WorkerAddressPort = *workaddr
		megaWorkspace = *fws
		if *portNumber != DefaultPort {
			os.Setenv(PORT, *portNumber)
		}
	} else {
		DispatcherAddressPort = os.ExpandEnv("$DISPATCHER_ADDR")
		WorkerAddressPort = os.ExpandEnv("$WORKER_ADDR")
	}

	isInContainer = *container
	ChdirHome()

	GL = TrLock{Lock: new(sync.RWMutex)}

	if !WorkerMode {
		InitConfigLoad()
		if !*nobrowser {
			LaunchBrowser()
		}
		if WorkerAddressPort != DefaultAddress {
			RegisterWorker(WorkerAddressPort)
		}
		ticker := time.NewTicker(Checkinterval)
		go MegaTimer(ticker)
	} else {
		SelfAnnounce(DispatcherAddressPort)
		Config = &MegaConfig{}
	}

	//psss go code here : func main()
	store := appdash.NewMemoryStore()

	// Listen on any available TCP port locally.
	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		log.Fatal(err)
	}
	collectorPort := l.Addr().(*net.TCPAddr).Port

	// Start an Appdash collection server that will listen for spans and
	// annotations and add them to the local collector (stored in-memory).
	cs := appdash.NewServer(l, appdash.NewLocalCollector(store))
	go cs.Start()

	// Print the URL at which the web UI will be running.
	appdashPort := 8700
	appdashURLStr := fmt.Sprintf("http://localhost:%d", appdashPort)
	appdashURL, err := url.Parse(appdashURLStr)
	if err != nil {
		log.Fatalf("Error parsing %s: %s", appdashURLStr, err)
	}
	color.Red("✅ Important!")
	log.Println("To see your traces, go to ", appdashURL)

	// Start the web UI in a separate goroutine.
	tapp, err := traceapp.New(nil, appdashURL)
	if err != nil {
		log.Fatal(err)
	}
	tapp.Store = store
	tapp.Queryer = store
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", appdashPort), tapp))
	}()

	tracer := appdashot.NewTracer(appdash.NewRemoteCollector(fmt.Sprintf(":%d", collectorPort)))
	opentracing.InitGlobalTracer(tracer)

	port := ":9001"
	if envport := os.ExpandEnv("$PORT"); envport != "" {
		port = fmt.Sprintf(":%s", envport)
	}
	log.Printf("Listenning on Port %v\n", port)
	http.HandleFunc("/", MakeHandler(Handler))

	http.HandleFunc("/momentum/templates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.FormValue("name") == "reset" || r.Method == "OPTIONS" {
			return
		} else if r.FormValue("name") == "ang" {
			w.Header().Set("Content-Type", "text/html")
			tmplRendered := Netang(r.FormValue("payload"))
			w.Write([]byte(tmplRendered))
		} else if r.FormValue("name") == "server" {
			w.Header().Set("Content-Type", "text/html")
			tmplRendered := Netserver(r.FormValue("payload"))
			w.Write([]byte(tmplRendered))
		} else if r.FormValue("name") == "jquery" {
			w.Header().Set("Content-Type", "text/html")
			tmplRendered := Netjquery(r.FormValue("payload"))
			w.Write([]byte(tmplRendered))

		}
	})

	http.HandleFunc("/funcfactory.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		w.Write([]byte(`function ang(dataOfInterface, cb){ jsrequestmomentum("/momentum/templates", {name: "ang", payload: JSON.stringify(dataOfInterface)},"POST",  cb) }
function server(dataOfInterface, cb){ jsrequestmomentum("/momentum/templates", {name: "server", payload: JSON.stringify(dataOfInterface)},"POST",  cb) }
function jquery(dataOfInterface, cb){ jsrequestmomentum("/momentum/templates", {name: "jquery", payload: JSON.stringify(dataOfInterface)},"POST",  cb) }
function Mega(  cb){
	var t = {}
	
	jsrequestmomentum("/momentum/funcs?name=Mega", t, "POSTJSON", cb)
}
function AddServer(  cb){
	var t = {}
	
	jsrequestmomentum("/momentum/funcs?name=AddServer", t, "POSTJSON", cb)
}
function DServer(Req , cb){
	var t = {}
	
	t.Req = Req
	jsrequestmomentum("/momentum/funcs?name=DServer", t, "POSTJSON", cb)
}
function UServer(Req , cb){
	var t = {}
	
	t.Req = Req
	jsrequestmomentum("/momentum/funcs?name=UServer", t, "POSTJSON", cb)
}
function AddContact(  cb){
	var t = {}
	
	jsrequestmomentum("/momentum/funcs?name=AddContact", t, "POSTJSON", cb)
}
function GetLog(Req , cb){
	var t = {}
	
	t.Req = Req
	jsrequestmomentum("/momentum/funcs?name=GetLog", t, "POSTJSON", cb)
}
function DContact(Req , cb){
	var t = {}
	
	t.Req = Req
	jsrequestmomentum("/momentum/funcs?name=DContact", t, "POSTJSON", cb)
}
function UContact(Req , cb){
	var t = {}
	
	t.Req = Req
	jsrequestmomentum("/momentum/funcs?name=UContact", t, "POSTJSON", cb)
}
function UMail(Req , cb){
	var t = {}
	
	t.Req = Req
	jsrequestmomentum("/momentum/funcs?name=UMail", t, "POSTJSON", cb)
}
function UTw(Req , cb){
	var t = {}
	
	t.Req = Req
	jsrequestmomentum("/momentum/funcs?name=UTw", t, "POSTJSON", cb)
}
function USetting(Req , cb){
	var t = {}
	
	t.Req = Req
	jsrequestmomentum("/momentum/funcs?name=USetting", t, "POSTJSON", cb)
}
function ProcessServer(Req , cb){
	var t = {}
	
	t.Req = Req
	jsrequestmomentum("/momentum/funcs?name=ProcessServer", t, "POSTJSON", cb)
}
function UpdateServer(Req , cb){
	var t = {}
	
	t.Req = Req
	jsrequestmomentum("/momentum/funcs?name=UpdateServer", t, "POSTJSON", cb)
}
function RegisterServer(Req , cb){
	var t = {}
	
	t.Req = Req
	jsrequestmomentum("/momentum/funcs?name=RegisterServer", t, "POSTJSON", cb)
}
function UpdateKubernetes(Req , cb){
	var t = {}
	
	t.Req = Req
	jsrequestmomentum("/momentum/funcs?name=UpdateKubernetes", t, "POSTJSON", cb)
}
function AddPod(Req , cb){
	var t = {}
	
	t.Req = Req
	jsrequestmomentum("/momentum/funcs?name=AddPod", t, "POSTJSON", cb)
}
function UpdatePod(Req , cb){
	var t = {}
	
	t.Req = Req
	jsrequestmomentum("/momentum/funcs?name=UpdatePod", t, "POSTJSON", cb)
}
function GetPods(  cb){
	var t = {}
	
	jsrequestmomentum("/momentum/funcs?name=GetPods", t, "POSTJSON", cb)
}
`))
	})

	http.HandleFunc("/momentum/funcs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.FormValue("name") == "reset" || r.Method == "OPTIONS" {
			return
		} else if r.FormValue("name") == "Mega" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadMega struct {
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadMega
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respresult0 := NetMega()

			resp["result"] = respresult0
			w.Write([]byte(mResponse(resp)))
		} else if r.FormValue("name") == "AddServer" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadAddServer struct {
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadAddServer
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respresult0 := NetAddServer()

			resp["result"] = respresult0
			w.Write([]byte(mResponse(resp)))
		} else if r.FormValue("name") == "DServer" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadDServer struct {
				Req Server
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadDServer
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respresult0 := NetDServer(tmvv.Req)

			resp["result"] = respresult0
			w.Write([]byte(mResponse(resp)))
		} else if r.FormValue("name") == "UServer" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadUServer struct {
				Req Server
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadUServer
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respresult0 := NetUServer(tmvv.Req)

			resp["result"] = respresult0
			w.Write([]byte(mResponse(resp)))
		} else if r.FormValue("name") == "AddContact" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadAddContact struct {
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadAddContact
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respresult0 := NetAddContact()

			resp["result"] = respresult0
			w.Write([]byte(mResponse(resp)))
		} else if r.FormValue("name") == "GetLog" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadGetLog struct {
				Req Server
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadGetLog
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respresult0 := NetGetLog(tmvv.Req)

			resp["result"] = respresult0
			w.Write([]byte(mResponse(resp)))
		} else if r.FormValue("name") == "DContact" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadDContact struct {
				Req Contact
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadDContact
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respresult0 := NetDContact(tmvv.Req)

			resp["result"] = respresult0
			w.Write([]byte(mResponse(resp)))
		} else if r.FormValue("name") == "UContact" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadUContact struct {
				Req Contact
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadUContact
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respresult0 := NetUContact(tmvv.Req)

			resp["result"] = respresult0
			w.Write([]byte(mResponse(resp)))
		} else if r.FormValue("name") == "UMail" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadUMail struct {
				Req MailSettings
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadUMail
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respresult0 := NetUMail(tmvv.Req)

			resp["result"] = respresult0
			w.Write([]byte(mResponse(resp)))
		} else if r.FormValue("name") == "UTw" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadUTw struct {
				Req TwilioInfo
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadUTw
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respresult0 := NetUTw(tmvv.Req)

			resp["result"] = respresult0
			w.Write([]byte(mResponse(resp)))
		} else if r.FormValue("name") == "USetting" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadUSetting struct {
				Req Settings
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadUSetting
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respresult0 := NetUSetting(tmvv.Req)

			resp["result"] = respresult0
			w.Write([]byte(mResponse(resp)))
		} else if r.FormValue("name") == "ProcessServer" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadProcessServer struct {
				Req string
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadProcessServer
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respresult0 := NetProcessServer(tmvv.Req)

			resp["result"] = respresult0
			w.Write([]byte(mResponse(resp)))
		} else if r.FormValue("name") == "UpdateServer" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadUpdateServer struct {
				Req Server
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadUpdateServer
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respresult0 := NetUpdateServer(tmvv.Req)

			resp["result"] = respresult0
			w.Write([]byte(mResponse(resp)))
		} else if r.FormValue("name") == "RegisterServer" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadRegisterServer struct {
				Req string
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadRegisterServer
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respresult0 := NetRegisterServer(tmvv.Req)

			resp["result"] = respresult0
			w.Write([]byte(mResponse(resp)))
		} else if r.FormValue("name") == "UpdateKubernetes" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadUpdateKubernetes struct {
				Req k8sConfig
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadUpdateKubernetes
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respresult0 := NetUpdateKubernetes(tmvv.Req)

			resp["result"] = respresult0
			w.Write([]byte(mResponse(resp)))
		} else if r.FormValue("name") == "AddPod" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadAddPod struct {
				Req PodConfig
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadAddPod
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respwatching0 := NetAddPod(tmvv.Req)

			resp["watching"] = respwatching0
			w.Write([]byte(mResponse(resp)))
		} else if r.FormValue("name") == "UpdatePod" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadUpdatePod struct {
				Req PodConfig
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadUpdatePod
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respresult0 := NetUpdatePod(tmvv.Req)

			resp["result"] = respresult0
			w.Write([]byte(mResponse(resp)))
		} else if r.FormValue("name") == "GetPods" {
			w.Header().Set("Content-Type", "application/json")
			type PayloadGetPods struct {
			}
			decoder := json.NewDecoder(r.Body)
			var tmvv PayloadGetPods
			err := decoder.Decode(&tmvv)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
				return
			}
			resp := db.O{}
			respresult0 := NetGetPods()

			resp["result"] = respresult0
			w.Write([]byte(mResponse(resp)))

		}
	})
	//+++extendgxmlmain+++
	http.Handle("/dist/", http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: "web"}))

	errgos := http.ListenAndServe(port, nil)
	if errgos != nil {
		log.Fatal(errgos)
	}

}

//+++extendgxmlroot+++
