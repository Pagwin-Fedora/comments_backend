package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"encoding/json"
	_ "github.com/lib/pq"
)

type Comment struct {
    Username string `json:"username"`
    Content string `json:"post"`
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
    if req.Method == "GET" {
	return
    }

    cookie, err := req.Cookie("login")
    post_fragment, err := resolve_cookie(cookie,err)
    if err != nil {
        w.Write([]byte("Server error"))
        w.WriteHeader(500)
        return
    }

    w.Write([]byte(post_fragment))

    comment := Comment{}
    json.NewDecoder(req.Body).Decode(&comment)
    
    
    tmpl, err := template.New("Single Comment template").Parse("<h4> {{.Username}} posted</h4>{{.Content}}")
    if err != nil {
        w.Write([]byte("Template fuckup"))
        w.WriteHeader(500)
        return
    }
    


    //
    tmpl.Execute(w,comment)
    w.Write([]byte(post_fragment))
}
}

//TODO: change to query sql db instead of giving a static comment list
func query_comments(db *sql.DB, path string) []Comment{
    return []Comment{
	{
	    Username:"e",
	    Content:"<script>alert(1)</script>",
	},
    }
}

func resolve_cookie(cookie *http.Cookie, err error) (string, error){
    if err == http.ErrNoCookie{
	return login_button(), nil
    }
    if err == nil {
	if cookie.Valid() != nil {
	    return login_button(), nil
	}
    } else {
	return "", err
    }
    return post_interface(cookie.Value), nil
}

func post_interface(username string) string{
    return "interface to post comment goes here"
}

func login_button() string{
    return "login button goes here"
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
        w.Write([]byte("Template fuckup"))
        w.WriteHeader(500)
        return
    }
    tmpl.Execute(w,comments)
}
}
