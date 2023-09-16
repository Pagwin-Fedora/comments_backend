package main

import (
	//"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	//"runtime/debug"
	"os"
	"regexp"
	"strconv"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	//_ "github.com/mattn/go-sqlite3"
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
	<div class="comment"><h4> {{.Username}}({{.Email}}) posted</h4><p>{{.Content}}</p></div>
    {{end}}
    {{define "list"}}
	{{range .}}
	    <div class="comment"><h4> {{.Username}}({{.Email}}) posted</h4><p>{{.Content}}</p></div>
	{{end}}
    {{end}}
    `)

func main(){

    connStr, err := gen_pq_str()
    if err != nil {
	fmt.Errorf("Bad Env var")
	os.Exit(1)
    }
    db, err := sqlx.Open("postgres", connStr)
    defer db.Close()
    if err != nil {
	fmt.Println(fmt.Errorf("error: %w", err))
        os.Exit(1)
    }
    // DB stuff
    _, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS comments(
        -- postgres seems to have SERIAL so we can use that instead of this being a primary key
        id SERIAL,
        blog_post TEXT NOT NULL,
        username TEXT NOT NULL,
        email TEXT NOT NULL,
        email_verified INTEGER NOT NULL,
        website TEXT,
        comment TEXT NOT NULL
    );
    `)
    if err != nil {
	fmt.Println(fmt.Errorf("error: %w", err))
        os.Exit(1)
    }
    _, err = db.Exec("CREATE TABLE IF NOT EXISTS banned_names(banned_name TEXT NOT NULL);")
    if err != nil {
	fmt.Println(fmt.Errorf("error: %w", err))
        os.Exit(1)
    }
    _, err = db.Exec("CREATE TABLE IF NOT EXISTS banned_domains(banned_domain TEXT NOT NULL);")
    if err != nil {
	fmt.Println(fmt.Errorf("error: %w", err))
        os.Exit(1)
    }

    //// Uncomment if testing to see if this works
    http.HandleFunc("/tmp", func(w http.ResponseWriter, req *http.Request){http.ServeFile(w,req,"index.html")})
    http.HandleFunc("/",resolve_comments(db))
    go http.ListenAndServe(":80",nil)
    fmt.Println("Listening on 80")
    select {}
}

func post_comment(db *sqlx.DB) func(w http.ResponseWriter, req *http.Request){
return func (w http.ResponseWriter, req *http.Request){
    //fmt.Println("post_comment: ",req.URL.Path)

    if req.Method == "GET" {
	return
    }

    comment := Comment{}
    json.NewDecoder(req.Body).Decode(&comment)

    //post_fragment := post_interface("e")
    cookie, err := req.Cookie("login")
    resolution := resolve_cookie(req.URL.Path, cookie, comment.Email, err)
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
    err := db.Select(&comments, "SELECT username, email, website, comment FROM comments WHERE blog_post=$1 ORDER BY id", path)
    if err != nil {
	return nil, err
    }
    ban_check, err := in_bans(db)
    if err != nil{
	return nil, err
    }
    comments = filter(comments, ban_check)
    return comments, err   
}
func in_bans(db *sqlx.DB)(func (comment Comment) bool, error){
    name_list :=  []string{}
    err := db.Select(&name_list,"SELECT banned_name FROM banned_names")
    if err != nil {
	return nil, err
    }
    domain_list := []string{}
    err = db.Select(&domain_list,"SELECT banned_domain FROM banned_domains")
    if err != nil {
	return nil, err
    }
    return func(c Comment) bool{
	comment_email_domain := strings.Split(c.Email,"@")
	for _, d := range domain_list {
	    if comment_email_domain[1] == d {
		return false
	    }
	}
	for _, n := range name_list {
	    if c.Email == n {
		return false
	    }
	}
	return true
	  return true 
    }, nil
}
type CookieResolution struct {
    UI string
    ValidatedEmail bool
    Err error
}
func resolve_cookie(path string, cookie *http.Cookie, email string, err error) CookieResolution{
    if err == http.ErrNoCookie || !is_validated(cookie, email){
	// TODO: send an email to the address
	return CookieResolution{
	    UI: post_interface(path, false),
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
	UI: post_interface(path, true),
	ValidatedEmail: true,
	Err:nil,
    }
}
func is_validated(cookie *http.Cookie, email string) bool{
    // TODO: make an sql table with cookies, emails and expiries
    return true
}

// TODO: swap out email_verified for an enum due to needing this method for resolve_comments
func post_interface(path string, email_verified bool) string{
    lead := ""
    if path[0] != '/' {
	lead = "/"
    }
    tmp :=  fmt.Sprintf(`<form hx-ext="json-enc" value="submit post"id="submission-form" hx-trigger="submit" hx-target="this" hx-post="%s%s" hx-swap="outerHTML">
	<label for="username">username</label>
	<input name="username" id="username" type="text" placeholder="username"/>
	<br/>
	<label for="email">email</label>
	<input name="email" id="email" type="email" placeholder="email"/>
	<br/>
	<label for="post">post</label>
	<input name="post" id="post" type="text" placeholder="post"/>
	<input name="submit" id="submit" value="Submit Comment" type="submit"/>
    </form>`, lead, path)
    return tmp
}

func  resolve_comments(db *sqlx.DB) func (w http.ResponseWriter,req *http.Request){
return func(w http.ResponseWriter,req *http.Request){
    //boilerplate
    if req.Method == "POST" {
	// this is hacky and the reason for why is because I didn't think through what I wanted the api to be at the start
	post_comment(db)(w,req)
	return
    }
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

    comments, err := query_comments(db, req.URL.Path)
    if err != nil {
        w.Write([]byte(fmt.Sprint(err)))
        w.WriteHeader(500)
        return
    }

    fragment := post_interface(req.URL.Path,false)
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

func gen_pq_str() (string, error){
    // setup the defaults
    URI := "localhost"
    Port := uint16(5432)
    User := "commentsdb"
    Password := "CHANGEME!!!"
    DBName := "commentsdb"
    //refer to docs if changing https://pkg.go.dev/github.com/lib/pq#hdr-Connection_String_Parameters
    SSL := "disable"
    
    tmp, set := os.LookupEnv("DB_URI")
    if set {
	URI = tmp
    }
    
    tmp, set = os.LookupEnv("DB_PORT")
    if set {
	tmp, err := strconv.ParseUint(tmp,10,16)
	if err != nil {
	    return "", err
	}
	Port = uint16(tmp)
    }

    tmp, set = os.LookupEnv("DB_USER")
    if set {
	User = tmp
    }
    tmp, set = os.LookupEnv("DB_PASSWORD")
    if set {
	Password = tmp
    }
    tmp, set = os.LookupEnv("DB_NAME")
    if set {
	DBName = tmp
    }
    tmp, set = os.LookupEnv("DB_SSL")
    if set {
	SSL = tmp
    }

    return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", URI, Port, User, Password, DBName, SSL), nil
}

// got this from https://stackoverflow.com/a/37563128
func filter(ss []Comment, test func(Comment) bool) (ret []Comment) {
    for _, s := range ss {
        if test(s) {
            ret = append(ret, s)
        }
    }
    return
}
