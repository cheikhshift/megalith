package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/cheikhshift/db"
	"github.com/cheikhshift/gos/core"
	gosweb "github.com/cheikhshift/gos/web"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/fatih/color"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"html"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
)

var store = sessions.NewCookieStore([]byte("a very very very very secret key"))

var Prod = true

var TemplateFuncStore template.FuncMap
var templateCache = gosweb.NewTemplateCache()

func StoreNetfn() int {
	TemplateFuncStore = template.FuncMap{"a": gosweb.Netadd, "s": gosweb.Netsubs, "m": gosweb.Netmultiply, "d": gosweb.Netdivided, "js": gosweb.Netimportjs, "css": gosweb.Netimportcss, "sd": gosweb.NetsessionDelete, "sr": gosweb.NetsessionRemove, "sc": gosweb.NetsessionKey, "ss": gosweb.NetsessionSet, "sso": gosweb.NetsessionSetInt, "sgo": gosweb.NetsessionGetInt, "sg": gosweb.NetsessionGet, "form": gosweb.Formval, "eq": gosweb.Equalz, "neq": gosweb.Nequalz, "lte": gosweb.Netlt, "LoadWebAsset": NetLoadWebAsset, "Mega": NetMega, "AddServer": NetAddServer, "DServer": NetDServer, "UServer": NetUServer, "AddContact": NetAddContact, "GetLog": NetGetLog, "DContact": NetDContact, "UContact": NetUContact, "UMail": NetUMail, "UTw": NetUTw, "USetting": NetUSetting, "ProcessServer": NetProcessServer, "UpdateServer": NetUpdateServer, "RegisterServer": NetRegisterServer, "UpdateKubernetes": NetUpdateKubernetes, "AddPod": NetAddPod, "UpdatePod": NetUpdatePod, "GetPods": NetGetPods, "ang": Netang, "bang": Netbang, "cang": Netcang, "server": Netserver, "bserver": Netbserver, "cserver": Netcserver, "jquery": Netjquery, "bjquery": Netbjquery, "cjquery": Netcjquery}
	return 0
}

var FuncStored = StoreNetfn()

type dbflf db.O

func renderTemplate(w http.ResponseWriter, p *gosweb.Page) {
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
				renderTemplate(w, pag) ///your-500-page"

			}
		}
	}()

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
			renderTemplate(w, pag) // "/your-500-page"

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
func MakeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if attmpt := apiAttempt(w, r); !attmpt {
			fn(w, r)
		}
		context.Clear(r)

	}
}

func mResponse(v interface{}) string {
	data, _ := json.Marshal(&v)
	return string(data)
}
func apiAttempt(w http.ResponseWriter, r *http.Request) (callmet bool) {
	var response string
	response = ""
	var session *sessions.Session
	var er error
	if session, er = store.Get(r, "session-"); er != nil {
		session, _ = store.New(r, "session-")
	}

	if strings.Contains(r.URL.Path, "/") {

		if strings.Contains(r.URL.Path, ".map") || strings.Contains(r.URL.Path, "web/{{ server.Image }}") {
			return true
		}

	}
	if r.Method == "RESET" {
		return true
	} else if isURL := (r.URL.Path == "/update/server" && r.Method == strings.ToUpper("POST")); !callmet && isURL {

		decoder := json.NewDecoder(r.Body)
		var tmvv PayloadOfRequest
		err := decoder.Decode(&tmvv)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
			return true
		}
		_ = NetProcessServer(tmvv.req)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(OK)

		callmet = true
	} else if isURL := (r.URL.Path == "/mega" && r.Method == strings.ToUpper("POST")); !callmet && isURL {

		Cfg := &MegaConfig{}
		LoadConfig(&Config)
		w.Header().Set("Content-Type", "application/json")
		retjson := []byte(mResponse(Cfg))
		w.Write(retjson)
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
func SetField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		invalidTypeError := errors.New("Provided value type didn't match obj field type")
		return invalidTypeError
	}

	structFieldValue.Set(val)
	return nil
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
func Handler(w http.ResponseWriter, r *http.Request) {
	var p *gosweb.Page
	p, err := loadPage(r.URL.Path)
	var session *sessions.Session
	var er error
	if session, er = store.Get(r, "session-"); er != nil {
		session, _ = store.New(r, "session-")
	}

	if err != nil {
		log.Println(err.Error())

		w.WriteHeader(http.StatusNotFound)

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
			renderTemplate(w, pag) //"/your-500-page"
		}
		session = nil
		context.Clear(r)
		return
	}

	if !p.IsResource {
		w.Header().Set("Content-Type", "text/html")
		p.Session = session
		p.R = r
		renderTemplate(w, p) //fmt.Sprintf("web%s", r.URL.Path)
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

	if lPage, ok := WebCache.Get(title); ok {
		return &lPage, nil
	}

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

}

var Config *MegaConfig
var GL TrLock
var isInContainer bool

func init() {

}

//
func NetLoadWebAsset(args ...interface{}) string {

	data, err := Asset(fmt.Sprintf("web%s", args[0].(string)))
	if err != nil {
		return err.Error()
	}
	return string(data)

}

//
func NetMega() (result *MegaConfig) {

	ShouldLock()
	if WorkerAddressPort != DefaultAddress {
		LoadConfig(&Config)
	}
	ShouldUnlock()
	return Config

}

//
func NetAddServer() (result []Server) {

	ShouldLock()
	randint := rand.Intn(200) + 50 + len(Config.Servers)
	genimage := fmt.Sprintf("https://picsum.photos/%v/%v", randint, randint)
	ns := Server{ID: core.NewLen(20), Nickname: "New server", Image: genimage}
	Config.Servers = append(Config.Servers, ns)
	if DispatcherAddressPort != DefaultAddress {
		ioutil.WriteFile(fmt.Sprintf(urlformat, GenLogName(ns.ID), LockExt), OK, 0700)
	}
	ShouldUnlock()
	SaveConfig(&Config)
	return Config.Servers

}

//
func NetDServer(req Server) (result []Server) {

	result = []Server{}
	ShouldLock()
	for _, target := range Config.Servers {
		if target.ID != req.ID {
			result = append(result, target)
		}
	}
	DeleteLog(req.ID)
	Config.Servers = result
	ShouldUnlock()
	SaveConfig(&Config)
	return

}

//
func NetUServer(req Server) (result bool) {

	ShouldLock()
	for index, target := range Config.Servers {
		if target.ID == req.ID {
			Config.Servers[index] = req
		}
	}
	ShouldUnlock()
	SaveConfig(&Config)
	return true

}

//
func NetAddContact() (result []Contact) {

	ShouldLock()
	nc := Contact{ID: core.NewLen(20), Nickname: "New contact"}
	Config.Contacts = append(Config.Contacts, nc)
	ShouldUnlock()
	SaveConfig(&Config)
	return Config.Contacts

}

//
func NetGetLog(req Server) (result RequestLog) {

	ShouldLock()
	LoadLog(req.ID, &result)
	ShouldUnlock()
	return

}

//
func NetDContact(req Contact) (result []Contact) {

	result = []Contact{}
	ShouldLock()
	for _, target := range Config.Contacts {
		if target.ID != req.ID {
			result = append(result, target)
		}
	}

	Config.Contacts = result
	ShouldUnlock()
	SaveConfig(&Config)
	return

}

//
func NetUContact(req Contact) (result bool) {

	ShouldLock()
	for index, target := range Config.Contacts {
		if target.ID == req.ID {
			Config.Contacts[index] = req
		}
	}
	ShouldUnlock()
	SaveConfig(&Config)
	return true

}

//
func NetUMail(req MailSettings) (result bool) {

	ShouldLock()
	Config.Mail = req
	ShouldUnlock()
	SaveConfig(&Config)
	return true

}

//
func NetUTw(req TwilioInfo) (result bool) {

	ShouldLock()
	Config.SMS = req
	ShouldUnlock()
	SaveConfig(&Config)
	return true

}

//
func NetUSetting(req Settings) (result bool) {

	ShouldLock()
	Config.Misc = req
	Config.LastReset = time.Now().Unix()
	ShouldUnlock()
	SaveConfig(&Config)
	return true

}

//
func NetProcessServer(req string) (result bool) {

	LoadConfig(&Config)
	server, index := FindServer(req)
	Process(server, index)
	return true

}

//
func NetUpdateServer(req Server) (result bool) {

	ShouldLock()
	_, index := FindServer(req.ID)
	Config.Servers[index].Uptime = req.Uptime
	ShouldUnlock()
	SaveConfig(&Config)
	return true

}

//
func NetRegisterServer(req string) (result bool) {

	RegisterWorker(req)
	return true

}

//
func NetUpdateKubernetes(req k8sConfig) (result bool) {

	ShouldLock()
	Config.KubeSettings = req
	ShouldUnlock()
	SaveConfig(&Config)
	return true

}

//
func NetAddPod(req PodConfig) (watching []PodConfig) {

	ShouldLock()
	Config.KubeSettings.Monitoring = append(Config.KubeSettings.Monitoring, req)
	watching = Config.KubeSettings.Monitoring
	ShouldUnlock()
	SaveConfig(&Config)
	return

}

//
func NetUpdatePod(req PodConfig) (result bool) {

	ShouldLock()
	for index, target := range Config.KubeSettings.Monitoring {
		if target.Name == req.Name {
			Config.KubeSettings.Monitoring[index] = req
		}
	}
	ShouldUnlock()
	SaveConfig(&Config)
	return true

}

//
func NetGetPods() (result []Pod) {

	list, _ := RequestPods()
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
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   true,
		Domain:   "",
	}

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
