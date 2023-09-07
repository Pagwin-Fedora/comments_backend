package main

import (
	//"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	//"os"
	"regexp"
	//_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/jmoiron/sqlx"
)

type Comment struct {
    Username string `json:"username" db:"username"`
    Email string `json:"email" db:"email"`
    Website sql.NullString `json:"website" db:"website"`
    Content string `json:"post" db:"comment"`
}

// there may be a way of using the single template in the list one but I'm not gonna bother figuring out how
var tmpl, err = template.New("Comment template").Parse(`
    {{define "single"}}
	<h4> {{.Username}}({{.Email}}) posted</h4>{{.Content}}
    {{end}}
    {{define "list"}}
	{{range .}}
	    <h4> {{.Username}}({{.Email}}) posted</h4>{{.Content}}
	{{end}}
    {{end}}
    `)

func main(){
    //connStr := "user=pqgotest dbname=pqgotest sslmode=verify-full"
    //db, err := sql.Open("postgres", connStr)

    // while testing using sqlite
    // https://github.com/mattn/go-sqlite3#connection-string
    connStr := "file:test.db"
    db, err := sqlx.Open("sqlite3", connStr)
    if err != nil {
	fmt.Println(fmt.Errorf("error: %w", err))
        return
    }

    http.HandleFunc("/tmp", func(w http.ResponseWriter, req *http.Request){http.ServeFile(w,req,"index.html")})
    http.HandleFunc("/",resolve_comments(db))
    http.HandleFunc("/post/", post_comment(db))
    go http.ListenAndServe(":8080",nil)
    fmt.Println("Listening on 8080")
    select {}
}

func post_comment(db *sqlx.DB) func(w http.ResponseWriter, req *http.Request){
return func (w http.ResponseWriter, req *http.Request){
    if req.Method == "GET" {
	return
    }

    comment := Comment{}
    json.NewDecoder(req.Body).Decode(&comment)

    //post_fragment := post_interface("e")
    cookie, err := req.Cookie("login")
    resolution := resolve_cookie(cookie,comment.Email, err)
    if resolution.Err != nil {
	w.Write([]byte(fmt.Sprint("Error:",resolution.Err)))
        w.WriteHeader(500)
        return
    }

    w.Write([]byte(resolution.UI))

    if comment.Content == "" {
	return
    }
    
    err = insert_comment(db, req.URL.Path, comment, false)
    if err != nil {
	w.Write([]byte(fmt.Sprint("insertion error: ", err)))
	w.WriteHeader(500)
        return
    }

    err = tmpl.ExecuteTemplate(w,"single",comment)
    if err != nil {
	w.Write([]byte(fmt.Sprint("Template fuckup: ", err)))
        w.WriteHeader(500)
        return
    }
    
}
}

func insert_comment(db *sqlx.DB, path string, comment Comment, email_verified bool) error {
    var tmp int
    if email_verified{
	tmp = 1
    }else {
	tmp = 0
    }
    _, err := db.Exec("INSERT INTO comments(blog_post, username, email, email_verified, comment) VALUES ($1,$2,$3,$4,$5)", path, comment.Username, comment.Email, tmp, comment.Content)
    return err
}

func query_comments(db *sqlx.DB, path string) ([]Comment, error){
    comments := []Comment{}
    err := db.Select(&comments, "SELECT username, email, website, comment FROM comments WHERE blog_post=$1", path)

    return comments, err   
}
type CookieResolution struct {
    UI string
    ValidatedEmail bool
    Err error
}
func resolve_cookie(cookie *http.Cookie, email string, err error) CookieResolution{
    if err == http.ErrNoCookie || !is_validated(cookie, email){
	// TODO: send an email to the address
	return CookieResolution{
	    UI: post_interface(false),
	    ValidatedEmail: false,
	    Err:nil,
	}
    }
    if err != nil{
	return CookieResolution{
	    UI: "",
	    ValidatedEmail: false,
	    Err:err,
	}
    }
    return CookieResolution{
	UI: post_interface(true),
	ValidatedEmail: true,
	Err:nil,
    }
}
func is_validated(cookie *http.Cookie, email string) bool{
    // TODO: make an sql table with cookies, emails and expiries
    return true
}

// TODO: swap out email_verified for an enum due to needing this method for resolve_comments
func post_interface(email_verified bool) string{
    return `<form hx-ext="json-enc" value="submit post"id="submission-form" hx-trigger="submit" hx-target="this" hx-post="/post/" hx-swap="outerHTML">
	<label for="username">username</label>
	<input name="username" id="username" type="text" placeholder="username"/>
	<label for="email">email</label>
	<input name="email" id="email" type="email" placeholder="email"/>
	<label for="post">post</label>
	<input name="post" id="post" type="text" placeholder="post"/>
	<label for="submit">submit</label>
	<input name="submit" id="submit" type="submit"/>
    </form>`
}

func  resolve_comments(db *sqlx.DB) func (w http.ResponseWriter,req *http.Request){
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

    comments, err := query_comments(db, "")
    if err != nil {
        w.Write([]byte(fmt.Sprint(err)))
        w.WriteHeader(500)
        return
    }

    fragment := post_interface(false)
    w.Write([]byte(fragment))

    //Does the work of rendering out the comments we got
    if err != nil {
	w.Write([]byte(fmt.Sprint("Template fuckup: ", err)))
        w.WriteHeader(500)
        return
    }
    tmpl.ExecuteTemplate(w,"list",comments)
}
}
