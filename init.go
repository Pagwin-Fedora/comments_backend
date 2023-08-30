package main

import (
    "html/template"
    "net/http"
    _ "database/sql"
    _ "github.com/lib/pq"
    "fmt"
    "regexp"
)

type Comment struct {
    Username string
    Content string
}

func main(){
    //connStr := "user=pqgotest dbname=pqgotest sslmode=verify-full"
    //db, err := sql.Open("postgres", connStr)
    //if err != nil {
    //    fmt.Errorf(" %w", err)
    //}
    http.HandleFunc("/",resolve_comments)
    http.ListenAndServe(":8080",nil)
}

func resolve_comments(w http.ResponseWriter,req *http.Request){
    //boilerplate
    url := fmt.Sprint(req.URL)
    mat, err := regexp.Match(".*favicon.ico",[]byte(url))
    //TODO: handle the error correctly
    if err != nil || mat{
        w.Write([]byte(""))
        return
    }
    w.Header().Add("Content-Type", "text/html")
    w.Write([]byte("<noscript>You need to enable javascript to see comments</noscript>"))


    //TODO: replace with grabbing comments from sql
    comments := []Comment{
	{
	    Username:"e",
	    Content:"<script>alert(1)</script>",
	},
    }
    
    //TODO: add in code to check for a login cookie, if one is present then great just give the static snippet for typing and submitting a comment(using htmx for convenience) if there's no login cookie then put a button telling the user to login if they want to comment

    //Does the work of rendering out the comments we got
    tmpl, err := template.New("Comment list template").Parse("{{range .}} <h4> {{.Username}} posted</h4>{{.Content}}{{end}}")
    if err != nil {
        fmt.Print(fmt.Errorf(" %w\n", err))
    }
    tmpl.Execute(w,comments)
}
