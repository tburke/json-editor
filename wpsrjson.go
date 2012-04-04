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

type Actors []Actor;

func actors(filename string) (v map[string]Actors) {
    obj, _ := os.Open(filename)
    dec := json.NewDecoder(obj)
    if err := dec.Decode(&v); err != nil {
        log.Println(err)
        return
    }
    return v
}

func (actors Actors) actor(id int64) Actor {
    for _, a := range actors {
        if a.Id == id {
            return a
        }
    }
  return Actor{"","",0,id,"person"}
}


func listActors(w io.Writer, actors Actors) {
    t, err := template.New("tib").Parse(`<h1>Actors</h1>
{{range .}}<a href="/edit/{{.Id}}">{{.DisplayName}}</a><br/>
{{end}}`)
    if err != nil {
        fmt.Printf("err %v",err)
        }
    err = t.Execute(w,actors)
    if err != nil {
        fmt.Printf("err %v",err)
        }
}

func editActor(w io.Writer, actors Actors, id int64) {
    t, _ := template.New("tib").Funcs(template.FuncMap{"eq": reflect.DeepEqual}).Parse(`<h1>edit</h1>
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
`)
    t.Execute(w,actors.actor(id))
}

func saveactor(id int64, v url.Values) {
    as := actors("objects.json")
    a := as["actors"].actor(id)
    a.DisplayName = v["displayname"][0]
    fmt.Fprintf(os.Stdout, "Actor is %v\n",a)
}

func edithandler(w http.ResponseWriter, r *http.Request) {
    // if GET, show edit. If POST update.
    ID, _ := strconv.ParseInt(r.URL.Path[len("/edit/"):], 10,64)
    if r.Method=="POST" {
        r.ParseForm()
        saveactor(ID, r.Form)
    }
    v := actors("objects.json")
    editActor(w, v["actors"], ID)
}

func listhandler(w http.ResponseWriter, r *http.Request) {
    listActors(w, actors("objects.json")["actors"])
}

func web() {
    http.HandleFunc("/", listhandler)
    http.HandleFunc("/edit/", edithandler)
    http.ListenAndServe(":8080", nil)
}

func main() {
    if len(os.Args) < 2 {
        listActors(os.Stdout, actors("objects.json")["actors"])
    } else if os.Args[1] == "web" {
        web()
    } else {
        ID, _ := strconv.ParseInt(os.Args[1],10,64)
        editActor(os.Stdout, actors("objects.json")["actors"],ID)
    }
}
    /*
    // dec := json.NewDecoder(os.Stdin)
    enc := json.NewEncoder(os.Stdout)
    for k, vv := range v {
        if k == "actors" {
            for _, u := range vv {
                fmt.Printf("%d\t%v\n",u.Id, u)
            }
        }
    }
    if err := enc.Encode(&v); err != nil {
        log.Println(err)
    }
    */

