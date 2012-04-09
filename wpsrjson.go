package main

import (
    "encoding/json"
    "net/http"
    "net/url"
    "log"
    "os"
    "io"
    "strconv"
    "fmt"
    "html/template"
    "reflect"
)

type Actor struct {
    DisplayName string `json:"displayName"`
    Url string `json:"url,omitempty"`
    Rid int64 `json:"rid,omitempty"`
    Id int64 `json:"id"`
    ObjectType string `json:"objectType"`
}

type Actors []*Actor;

func actors(filename string) (v map[string]Actors) {
    obj, _ := os.Open(filename)
    dec := json.NewDecoder(obj)
    if err := dec.Decode(&v); err != nil {
        log.Println(err)
        return
    }
    return v
}

func (actors Actors) actor(id int64) *Actor {
    for _, a := range actors {
        if a.Id == id {
            return a
        }
    }
    return nil
}


func listActors(w io.Writer, actors Actors) {
    t, _ := template.New("tib").Parse(`<h1>Actors</h1>
{{range .}}<a href="/edit/{{.Id}}">{{.DisplayName}}</a><br/>
{{end}}`)
    t.Execute(w,actors)
}

func editActor(w io.Writer, actors Actors, id int64) {
    t, _ := template.New("tib").Funcs(template.FuncMap{"eq": reflect.DeepEqual}).Parse(editTemplate)
    a := actors.actor(id)
    if a == nil {
        js := loadJson(fmt.Sprintf("https://graph.facebook.com/%d",id))
        ent := js.(map[string]interface{})
        a = &Actor{ent["name"].(string),ent["link"].(string),0,id,"person"}
    }
    t.Execute(w,a)
}

func saveactor(id int64, v url.Values) {
    as := actors("object2.json")
    a := as["actors"].actor(id)
    if a == nil {
      a = &Actor{"","",0,id,"person"}
      as["actors"] = append(as["actors"], a)
    }
    RID, _ := strconv.ParseInt(v["rid"][0],10,64)
    a.Rid = RID
    a.DisplayName = v["displayname"][0]
    a.Url = v["url"][0]
    a.ObjectType = v["objecttype"][0]
    fh, _ := os.Create("object2.json")
    json.NewEncoder(fh).Encode(&as)
}

func edithandler(w http.ResponseWriter, r *http.Request) {
    // if GET, show edit. If POST update.
    ID, _ := strconv.ParseInt(r.URL.Path[len("/edit/"):], 10,64)
    if r.Method=="POST" {
        r.ParseForm()
        saveactor(ID, r.Form)
    }
    v := actors("object2.json")
    editActor(w, v["actors"], ID)
}

func listhandler(w http.ResponseWriter, r *http.Request) {
    listActors(w, actors("object2.json")["actors"])
}

func loadJson(url string) (v interface{}) {
    resp, _ := http.Get(url)
    defer resp.Body.Close()
    dec := json.NewDecoder(resp.Body)
    if err := dec.Decode(&v); err != nil {
        log.Println(err)
        return
    }
    return v
}

func web() {
    http.HandleFunc("/", listhandler)
    http.HandleFunc("/edit/", edithandler)
    http.ListenAndServe(":8080", nil)
}

func main() {
    if len(os.Args) < 2 {
        listActors(os.Stdout, actors("object2.json")["actors"])
    } else if os.Args[1] == "web" {
        web()
    } else {
        ID, _ := strconv.ParseInt(os.Args[1],10,64)
        editActor(os.Stdout, actors("object2.json")["actors"],ID)
    }
}

const editTemplate = `<h1>edit</h1>
<a href="/">List</a><br/>
<form action="" method="POST">
<h2>{{.Id}}</h2>
displayName: <input name="displayname" value="{{.DisplayName}}" size="50" /><br/>
url: <input name="url" value="{{.Url}}" size="50" /><br/>
rid: <input name="rid" value="{{.Rid}}" /><br/>
<select name="objecttype">
<option {{if eq .ObjectType "service"}}selected="1"{{end}}>service</option>
<option {{if eq .ObjectType "person"}}selected="1"{{end}}>person</option>
</select><br/>
 {{.ObjectType}}
<input type="submit" value="submit" />
</form>
`

