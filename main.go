package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"

	uuid "github.com/satori/go.uuid"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	UserName string
	Password []byte
	First    string
	Last     string
}

type paint_detail struct {
	Filename    string
	Username    string
	Category    string
	Description string
}

var tpl *template.Template
var dbUsers = map[string]user{}      // user ID, user
var dbSessions = map[string]string{} // session ID, user ID
var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://bond:password@localhost/art_gallery?sslmode=disable")
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	tpl = template.Must(template.ParseGlob("login/*"))
	bs, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	dbUsers["test@test.com"] = user{"test@test.com", bs, "James", "Bond"}
}

func main() {
	// http.HandleFunc("/", index)
	//http.Handle("/login", http.StripPrefix("/login", http.FileServer(http.Dir("/login"))))
	http.Handle("/", http.FileServer(http.Dir("img/")))
	// http.HandleFunc("/bar", bar)
	http.HandleFunc("/category/", display)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/nav", nav)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/about",about)
	http.HandleFunc("/exhibition",exhibition)
	http.HandleFunc("/profile", profile)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.ListenAndServe(":8000", nil)

}

func display(w http.ResponseWriter, req *http.Request) {
	res := strings.Split(req.URL.Path, "/")
	category := res[len(res)-1]
	cur := make([]paint_detail, 0)

	rows, err := db.Query("SELECT * FROM paintings where category= $1", category)

	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		tmp := paint_detail{}
		err = rows.Scan(&tmp.Username, &tmp.Category, &tmp.Description, &tmp.Filename)
		cur = append(cur, tmp)
	}

	for i := 0; i < len(cur); i++ {
		cur[i].Filename = strings.TrimSpace(cur[i].Filename)
		cur[i].Filename = "/" + cur[i].Filename
	}

	tpl.ExecuteTemplate(w, "category.html", cur)
}

func upload(w http.ResponseWriter, req *http.Request) {

	src, hdr, err := req.FormFile("myFile")
	if err != nil {
		panic(err)
	}
	defer src.Close()

	fileName := hdr.Filename
	dst, err := os.Create("img/" + fileName)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer dst.Close()

	io.Copy(dst, src)

	// req.ParseMultipartForm(10 << 20)
	// file, handler, err := req.FormFile("myFile")
	// if err != nil {
	// 	fmt.Println("Error Retrieving the File")
	// 	fmt.Println(err)
	// 	return
	// }
	// defer file.Close()
	// fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	// fmt.Printf("File Size: %+v\n", handler.Size)
	// fmt.Printf("MIME Header: %+v\n", handler.Header)

	// // Create a temporary file within our temp-images directory that follows
	// // a particular naming pattern
	// tempFile, err := ioutil.TempFile("img", "handler.Filename")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// defer tempFile.Close()

	// // byte array
	// // read all of the contents of our uploaded file into a
	// fileBytes, err := ioutil.ReadAll(file)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// updating category and description
	cat := req.FormValue("cat")
	des := req.FormValue("des")
	// get the username using cookies
	un := getUser(w, req)
	fmt.Println(cat)
	fmt.Println(des)
	fmt.Println(un)
	_, err = db.Exec("INSERT INTO paintings(user_id,category,description,filename) VALUES ($1,$2,$3,$4)", un, cat, des, fileName)
	if err != nil {
		fmt.Printf("gandu\n")
	}
	// write this byte array to our temporary file
	// tempFile.Write(fileBytes)
	http.Redirect(w, req, "/profile", http.StatusSeeOther)
	// return that we have successfully uploaded our file!
	// fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func profile(w http.ResponseWriter, req *http.Request) {
	cur := make([]paint_detail, 0)
	un := getUser(w, req)
	fmt.Println(" here" + un)
	rows, err := db.Query("SELECT * FROM paintings where user_id= $1", un)

	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		tmp := paint_detail{}
		err = rows.Scan(&tmp.Username, &tmp.Category, &tmp.Description, &tmp.Filename)
		cur = append(cur, tmp)
	}

	for i := 0; i < len(cur); i++ {
		cur[i].Filename = strings.TrimSpace(cur[i].Filename)
		cur[i].Filename = "/" + cur[i].Filename
	}

	tpl.ExecuteTemplate(w, "profile.html", cur)
}

func nav(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL)
	tpl.ExecuteTemplate(w, "nav.html", nil)
}

func about(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL)
	tpl.ExecuteTemplate(w, "about.html", nil)
}

func exhibition(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL)
	tpl.ExecuteTemplate(w, "exhibition.html", nil)
}

func signup(w http.ResponseWriter, req *http.Request) {
	if alreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	// var u user

	// fmt.Println("here")
	// process form submission
	if req.Method == http.MethodPost {
		// get form values
		un := req.FormValue("username")
		p := req.FormValue("password")
		f := req.FormValue("firstname")
		l := req.FormValue("lastname")
		e := req.FormValue("email_id")
		// username taken?

		if un == "" || p == "" || f == "" || l == "" || e == "" {
			http.Error(w, "Fill all the fields", http.StatusForbidden)
			return
		}
		// if _, ok := dbUsers[un]; ok {
		// 	http.Error(w, "Username already taken", http.StatusForbidden)
		// 	return
		// }
		row := db.QueryRow("SELECT * FROM user_details where user_id= $1", un)
		err := row.Scan()

		if err != sql.ErrNoRows {
			http.Error(w, "Username already taken", http.StatusForbidden)
			return
		}

		_, err = db.Exec("INSERT INTO user_details(user_id,fname,lname,email_id,password) VALUES ($1,$2,$3,$4,$5)", un, f, l, e, p)

		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}
		// create session
		sID, _ := uuid.NewV4()
		c := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
		dbSessions[c.Value] = un
		// store user in dbUsers
		// bs, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
		// if err != nil {
		// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
		// 	return
		// }

		//u = user{un, bs, f, l}
		//dbUsers[un] = u
		// redirect
		http.Redirect(w, req, "/nav", http.StatusSeeOther) // change it to somewhere w
		return
	}
	// fmt.Println("gandu\n")
	tpl.ExecuteTemplate(w, "signup.html", nil)
}

func login(w http.ResponseWriter, req *http.Request) {
	if alreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	if req.Method == http.MethodGet {

	}
	// process form submission
	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		p := string(req.FormValue("password"))
		var tmp string
		row := db.QueryRow("SELECT password FROM user_details where user_id=$1", un)
		err := row.Scan(&tmp)
		if err == sql.ErrNoRows {
			http.Error(w, "User not registered ", http.StatusForbidden)
			return
		}
		if err != nil {
			fmt.Println("asdf")
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}
		xx := strings.TrimSpace(tmp)
		fmt.Println(xx)
		fmt.Println(p)
		if xx != p {
			fmt.Printf("%vadsf\n", p)
			fmt.Printf("%vasdf\n", tmp)
			fmt.Println("hadn ")
			http.Redirect(w, req, "/login", http.StatusSeeOther)
			// http.Error(w, "Username and password doesnot match", http.StatusForbidden)
			return
		}
		// u, ok := dbUsers[un]
		// if !ok {
		// 	http.Error(w, "Username and/or password do not match", http.StatusForbidden)
		// 	return
		// }
		// err := bcrypt.CompareHashAndPassword(u.Password, []byte(p))
		// if err != nil {
		// 	http.Error(w, "Username and/or password do not match", http.StatusForbidden)
		// 	return
		// }
		sID, _ := uuid.NewV4()
		c := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
		dbSessions[c.Value] = un
		http.Redirect(w, req, "/nav", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "index.html", nil)
}

func logout(w http.ResponseWriter, req *http.Request) {
	fmt.Println("gandasdfsadf")
	c, err := req.Cookie("session")
	if err != nil {
		fmt.Println("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		panic(err)
	}
	delete(dbSessions, c.Value)
	c = &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, c)
	http.Redirect(w, req, "/nav", http.StatusSeeOther)
}

func getUser(w http.ResponseWriter, req *http.Request) string {
	// get cookie
	c, err := req.Cookie("session")
	if err != nil {
		panic(err)
	}
	un := dbSessions[c.Value]
	return un
}

func alreadyLoggedIn(req *http.Request) bool {
	c, err := req.Cookie("session")
	if err != nil {
		return false
	}
	un := dbSessions[c.Value]
	_, ok := dbUsers[un]
	return ok
}

// func index(w http.ResponseWriter, req *http.Request) {
// 	u := getUser(w, req)
// 	tpl.ExecuteTemplate(w, "index.html", u)
// }

// func bar(w http.ResponseWriter, req *http.Request) {
// 	u := getUser(w, req)
// 	if !alreadyLoggedIn(req) {
// 		http.Redirect(w, req, "/", http.StatusSeeOther)
// 		return
// 	}
// 	tpl.ExecuteTemplate(w, "bar.html", u)
// }
