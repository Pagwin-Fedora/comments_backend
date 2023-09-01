package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"regexp"
	_ "github.com/lib/pq"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type Comment struct {
    Username string `json:"username"`
    Content string `json:"post"`
}

// there may be a way of using the single template in the list one but I'm not gonna bother figuring out how
var tmpl, err = template.New("Comment template").Parse(`
    {{define "single"}}
	<h4> {{.Username}} posted</h4>{{.Content}}
    {{end}}
    {{define "list"}}
	{{range .}}
	    <h4> {{.Username}} posted</h4>{{.Content}}
	{{end}}
    {{end}}
    `)

func main(){
    connStr := "user=pqgotest dbname=pqgotest sslmode=verify-full"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
	fmt.Println(fmt.Errorf("error: %w", err))
        return
    }
    // refer to https://github.com/golang/oauth2/blob/master/example_test.go
    google_oauth_cfg := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{},
		Endpoint: google.Endpoint,
	}
    github_oauth_cfg := &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Scopes:       []string{},
		Endpoint: github.Endpoint,
	}
    if err != nil {
        fmt.Println(fmt.Errorf(" %w", err))
    }
    http.HandleFunc("/tmp", func(w http.ResponseWriter, req *http.Request){http.ServeFile(w,req,"index.html")})
    http.HandleFunc("/",resolve_comments(db))
    http.HandleFunc("/post/", post_comment(db))
    http.HandleFunc("/oauth",oauth_callback)
    go http.ListenAndServe(":8080",nil)
    fmt.Println("Listening on 8080")
    select {}
}

func oauth_callback(w http.ResponseWriter, req *http.Request){
    ctx := context.Background()
    db := ctx.Value("db").(*sql.DB)
    ref := req.Referer()
    mat, err := regexp.Match(`google\.com`,[]byte(ref))
    if err != nil {
        w.Write([]byte("regex oops"))
        w.WriteHeader(500)
        return
    }
    // set which table
    var table_prefix string
    if mat{
	//google case 
	table_prefix = ""
    } else {
	//github case 
	table_prefix = ""
    }
    var username string
    err = db.QueryRow(fmt.Sprintf("SELECT username FROM %s_oauth WHERE token=?", table_prefix), idk_figure_out_where_oauth_token_at).Scan(&username)
    if err == sql.ErrNoRows {
	// TODO: figure out how tf we get the username that the user wants
	err = db.Exec("")
    }
    w.Header().Add("Set-Cookie","login=")
}

func post_comment(db *sql.DB) func(w http.ResponseWriter, req *http.Request){
return func (w http.ResponseWriter, req *http.Request){
    if req.Method == "GET" {
	return
    }

    //post_fragment := post_interface("e")
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

    if comment.Content == "" {
	return
    }
    
    if err != nil {
        w.Write([]byte("Template fuckup"))
        w.WriteHeader(500)
        return
    }
    


    //
    tmpl.ExecuteTemplate(w,"single",comment)
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
    return `<form hx-ext="json-enc" value="submit post"id="submission-form" hx-trigger="submit" hx-target="this" hx-post="/post/" hx-swap="outerHTML">
	<label for="username">username</label>
	<input name="username" id="username" type="text" placeholder="username"/>
	<label for="post">post</label>
	<input name="post" id="post" type="text" placeholder="post"/>
	<label for="submit">submit</label>
	<input name="submit" id="submit" type="submit"/>
    </form>`
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

    comments := query_comments(db, "")
    
    //cookie, err := req.Cookie("login")
    //fragment, err := resolve_cookie(cookie,err)
    //if err != nil {
    //    w.Write([]byte("error handling cookie"))
    //    w.WriteHeader(500)
    //    return
    //}
    fragment := post_interface("e")
    w.Write([]byte(fragment))

    //Does the work of rendering out the comments we got
    if err != nil {
        w.Write([]byte("Template fuckup"))
        w.WriteHeader(500)
        return
    }
    tmpl.ExecuteTemplate(w,"list",comments)
}
}
