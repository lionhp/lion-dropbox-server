package main

import (
	"fmt"
	"github.com/astaxie/beego/config"
	"golang.org/x/net/context"
	"golang.org/x/net/webdav"
	"log"
	"net/http"
	"os"
)

func main() {
	conf, err := config.NewConfig("ini", "./config.ini")
	if err != nil {
		log.Println("read config failed, err:", err)
		return
	}
	port := conf.String("server::port")
	uname := conf.String("user::name")
	upwd := conf.String("user::pwd")
	root := conf.String("root::path")

	fs := &webdav.Handler{
		FileSystem: webdav.Dir(root),
		LockSystem: webdav.NewMemLS(),
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		username, pwd, ok := r.BasicAuth()
		if r.Method != "MKCOL" && !ok {
			w.Header().Set("Content-Type", "text/html;charset=utf-8")
			w.Header().Set("WWW-Authenticate", `Basic realm=" Lion Cloud"`)
			w.WriteHeader(401)
			return
		}
		if r.Method != "MKCOL" && (username != uname || pwd != upwd) {
			http.Error(w, "Account/Password Error", 401)
			return
		}

		/*switch r.Method {
		case "PUT", "DELETE", "PROPPATCH", "MKCOL", "COPY", "MOVE":
			http.Error(w, "Your request method have no Permission", 403)
			return
		}*/
		if r.Method == "GET" && handleDirList(fs.FileSystem, w, r) {
			return
		}
		fs.ServeHTTP(w, r)
	})

	//http.ListenAndServe(":"+port, nil)
	http.ListenAndServeTLS(":"+port, "./server.crt", "./server.key", nil)
}

func handleDirList(fs webdav.FileSystem, w http.ResponseWriter, r *http.Request) bool {
	ctx := context.Background()

	f, err := fs.OpenFile(ctx, r.URL.Path, os.O_RDONLY, 0)

	if err != nil {
		return false
	}
	defer f.Close()

	if fi, _ := f.Stat(); fi != nil && !fi.IsDir() {
		return false
	}

	dirs, err := f.Readdir(-1)
	if err != nil {
		log.Print(w, "Read file error", 500)
		return false
	}

	w.Header().Set("Content-Type", "text/html;charset=utf-8")
	fmt.Fprintf(w, "<pre>\n")
	for _, d := range dirs {
		name := d.Name()
		if d.IsDir() {
			name += "/"
		}
		fmt.Fprintf(w, "<a href=\"%s\">%s</a>\n", name, name)
	}
	fmt.Fprintf(w, "</pre>\n")
	return true
}
