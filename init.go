package main

import (
    "database/sql"
    "fmt"
    "html/template"
    "net/http"
    "regexp"
    "encoding/json"
    _ "github.com/lib/pq"
    "golang.org/x/oauth2"
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
    // refer to https://github.com/golang/oauth2/blob/master/example_test.go
    oauth_cfg := &oauth2.Config{
		ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
		Scopes:       []string{"SCOPE1", "SCOPE2"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://provider.com/o/oauth2/auth",
			TokenURL: "https://provider.com/o/oauth2/token",
		},
	}
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
