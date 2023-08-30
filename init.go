package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	_ "github.com/lib/pq"
)

type Comment struct {
    Username string
    Content string
}

func main(){
    connStr := "user=pqgotest dbname=pqgotest sslmode=verify-full"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        fmt.Println(fmt.Errorf(" %w", err))
    }
    http.HandleFunc("/tmp", func(w http.ResponseWriter, req *http.Request){http.ServeFile(w,req,"index.html")})
    http.HandleFunc("/",resolve_comments(db))
    http.HandleFunc("/post/", post_comment(db))
    go http.ListenAndServe(":8080",nil)
    fmt.Println("Listening on 8080")
    select {}
}

func post_comment(db *sql.DB) func(w http.ResponseWriter, req *http.Request){
return func (w http.ResponseWriter, req *http.Request){
    cookie, err := req.Cookie("login")
    err = req.ParseForm()
    if err != nil {
	fmt.Println(err)
    }
    pf := req.Form
    username := pf.Get("username")
    post := pf.Get("post")
    fmt.Println("username: ", username)
    fmt.Println("post:", post)
    if err == http.ErrNoCookie {
	w.Write([]byte("I don't know how you did this but please stop"))
    }
    post_fragment, err :=resolve_cookie(cookie,err)
    if err != nil {
        w.Write([]byte(""))
        w.WriteHeader(500)
        return
    }
    // TODO: handle form post or whatever htmx sends and turn that into a comment Object which we can apply a template to and stick into the db also stick comment into db based on uri /article-name/post
    w.Write([]byte(post_fragment))
}
}

//TODO: change to query sql db instead of giving a static comment list
func query_comments(db *sql.DB) []Comment{
    return []Comment{
	{
	    Username:"e",
	    Content:"<script>alert(1)</script>",
	},
    }
}

func resolve_cookie(cookie *http.Cookie, err error) (string, error){
    if err == http.ErrNoCookie{
	return "login button goes here", nil
    }
    if err != nil{
	return "", err
    }
    return "interface to post comment goes here", nil
}

func  resolve_comments(db *sql.DB) func (w http.ResponseWriter,req *http.Request){
return func(w http.ResponseWriter,req *http.Request){
    //boilerplate
    url := fmt.Sprint(req.URL)
    mat, err := regexp.Match(".*favicon.ico",[]byte(url))
    if err != nil{
        w.Write([]byte("error with the regex"))
	w.WriteHeader(500)
        return
    }
    if mat{
	w.WriteHeader(404)
	w.Header().Set("Status","404")
        return
    }
    w.Header().Add("Content-Type", "text/html")


    comments := query_comments(db)
    
    cookie, err := req.Cookie("login")
    fragment, err := resolve_cookie(cookie,err)
    if err != nil {
        w.Write([]byte("error handling cookie"))
        w.WriteHeader(500)
        return
    }
    w.Write([]byte(fragment))

    //Does the work of rendering out the comments we got
    tmpl, err := template.New("Comment list template").Parse("{{range .}} <h4> {{.Username}} posted</h4>{{.Content}}{{end}}")
    if err != nil {
        fmt.Print(fmt.Errorf(" %w\n", err))
    }
    tmpl.Execute(w,comments)
}
}
