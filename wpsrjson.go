package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
)

type Actor struct {
	DisplayName string `json:"displayName"`
	Url         string `json:"url,omitempty"`
	Rid         int64  `json:"rid,omitempty"`
	Id          int64  `json:"id"`
	ObjectType  string `json:"objectType"`
}

type Actors map[string][]*Actor

const actorsFile = "objects.json"
const actorsUrl = "http://cdn.wapolabs.com/trove/authors/objects.json"

func (actors *Actors) load() *Actors {
	var obj io.ReadCloser
	obj, ferr := os.Open(actorsFile)
	if ferr != nil {
		loadJson(actorsUrl, actors)
	} else if err := json.NewDecoder(obj).Decode(actors); err != nil {
		fmt.Println(err)
	}
	return actors
}

func (actors Actors) actor(id int64) *Actor {
	for _, a := range actors["actors"] {
		if a.Id == id {
			return a
		}
	}
	return nil
}

func (actors Actors) save(w io.Writer) {
	json.NewEncoder(w).Encode(&actors)
}

func (actors Actors) list(w io.Writer) {
	rootTemplate.Execute(w, actors["actors"])
}

var fbmatch = regexp.MustCompile(`[^\/]+$`)
var dig = regexp.MustCompile(`^\d+$`)
func (actors Actors) edit(w io.Writer, sid string) {
	var a *Actor
	match := fbmatch.FindString(sid)
	if len(match) < 1 {
		return
	} else if dig.MatchString(sid) {
		id, _ := strconv.ParseInt(match, 10, 64)
		a = actors.actor(id)
	}
	var ent map[string]interface{}
	if a == nil && loadJson(fmt.Sprintf("http://graph.facebook.com/%s", match), &ent) == nil {
		id, _ := strconv.ParseInt(ent["id"].(string), 10, 64)
		a = &Actor{ent["name"].(string), ent["link"].(string), 0, id, "person"}
	}
	editTemplate.Execute(w, a)
}

func (actor *Actor) update(v url.Values) {
	actor.Rid, _ = strconv.ParseInt(v["rid"][0], 10, 64)
	actor.DisplayName = v["displayname"][0]
	actor.Url = v["url"][0]
	actor.ObjectType = v["objecttype"][0]
}

func loadJson(url string, v interface{}) error {
	resp, urlerr := http.Get(url)
	if urlerr != nil {
		return urlerr
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

func edithandler(w http.ResponseWriter, r *http.Request) {
	// if GET, show edit. If POST update.
	r.ParseForm()
	var id string
	if len(r.URL.Path) > len(editPath) {
		id = r.URL.Path[len(editPath):]
	} else {
		id = r.Form["id"][0]
	}
	var actors Actors
	actors.load()
	if r.Method == "POST" && dig.MatchString(id) {
		ID, _ := strconv.ParseInt(id, 10, 64)
		a := actors.actor(ID)
		if a == nil {
			a = &Actor{"", "", 0, ID, "person"}
			actors["actors"] = append(actors["actors"], a)
		}
		a.update(r.Form)
		fh, _ := os.Create(actorsFile)
		actors.save(fh)
	}
	actors.edit(w, id)
}

func listhandler(w http.ResponseWriter, r *http.Request) {
	var actors Actors
	actors.load().list(w)
}

func web(port string) {
	http.HandleFunc("/", listhandler)
	http.HandleFunc(editPath, edithandler)
	fmt.Printf("Listening on http://localhost%s\n^C to end\n",port)
	http.ListenAndServe(port, nil)
}

func main() {
	var port string
	flag.StringVar(&port, "http", ":8080", "http port")
	flag.Parse()
	var actors Actors
	if flag.NArg() < 1 {
		web(port)
	} else if flag.Arg(0) == "list" {
		actors.load().list(os.Stdout)
	} else {
		actors.load().edit(os.Stdout, flag.Arg(0))
	}
}

const editPath = "/edit/"

var rootTemplate = template.Must(template.New("root").Parse(`
<h1>Actors</h1>
<form action="/edit/" method="POST">
New Actor: <input name="id" >
</form>
<ul>
{{range .}}<li><a href="/edit/{{.Id}}">{{.DisplayName}}</a></li>
{{end}}</ul>
`))

var editTemplate = template.Must(template.New("edit").Funcs(
	template.FuncMap{"eq": reflect.DeepEqual}).Parse(`
<h1>edit</h1>
<a href="/">List</a><br/>
<form action="" method="POST">
<h2>{{.Id}}</h2>
<input name="id" type="hidden" value="{{.Id}}" />
displayName: <input name="displayname" value="{{.DisplayName}}" size="50" /><br/>
url: <input name="url" value="{{.Url}}" size="50" /><br/>
rid: <input name="rid" value="{{.Rid}}" /><br/>
<select name="objecttype">
<option {{if eq .ObjectType "service"}}selected="1"{{end}}>service</option>
<option {{if eq .ObjectType "person"}}selected="1"{{end}}>person</option>
</select><br/>
 {{.ObjectType}}
<input type="submit" value="save" />
</form>
`))
