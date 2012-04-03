package main

import (
    "encoding/json"
    "net/http"
    "log"
    "os"
    "io"
    "strconv"
    "fmt"
    "html/template"
)

type Actor struct {
    DisplayName string `json:"displayName"`
    Url string `json:"url,omitempty"`
    Rid int64 `json:"rid,omitempty"`
    Id int64 `json:"id"`
    ObjectType string `json:"objectType"`
}

func actors(filename string) (v map[string][]Actor) {
    obj, _ := os.Open(filename)
    dec := json.NewDecoder(obj)
    if err := dec.Decode(&v); err != nil {
        log.Println(err)
        return
    }
    return v
}

func listActors(w io.Writer, actors []Actor) {
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

func editActor(w io.Writer, actors []Actor, id int64) {
    t, _ := template.New("tib").Parse(`<h1>edit</h1>
<a href="/">List</a><br/>
<form action="" method="POST">
<h2>{{.Id}}</h2>
displayName: <input name="DisplayName" value="{{.DisplayName}}"  size="50" /><br/>
url: <input name="Url" value="{{.Url}}" size="50" /><br/>
rid: <input name="Rid" value="{{.Rid}}" /><br/>
<select>
<option value="service">service</option>
<option value="person">person</option>
</select><br/>
 {{.ObjectType}}
<input type="submit" value="submit"/>
</form>
`)
    for _, a := range actors {
        if a.Id == id {
            t.Execute(w,a)
        }
    }
}

func edithandler(w http.ResponseWriter, r *http.Request) {
    ID, _ := strconv.ParseInt(r.URL.Path[len("/edit/"):], 10,64)
    v := actors("objects.json")
    editActor(w, v["actors"], ID)
}

func handler(w http.ResponseWriter, r *http.Request) {
    v := actors("objects.json")
    listActors(w, v["actors"])
}

func web() {
    http.HandleFunc("/", handler)
    http.HandleFunc("/edit/", edithandler)
    http.ListenAndServe(":8080", nil)
}

func main() {
    if len(os.Args) < 2 {
        v := actors("objects.json")
        listActors(os.Stdout, v["actors"])
    } else if os.Args[1] == "web" {
        web()
    } else {
        v := actors("objects.json")
        ID, _ := strconv.ParseInt(os.Args[1],10,64)
        editActor(os.Stdout, v["actors"],ID)
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

