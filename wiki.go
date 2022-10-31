package main

import (
	"errors"
	//"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
)

type Page struct {
	Title string
	//[]byte=a byte slice
	Body []byte
}

//template.Must()在通过非nil时死机，否则返回一个*Template
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
//regexp.MustCompile()解析和编译正则表达式，如果表达式编译失败会死机，而Complie会返回第二个参数
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
	//p1 := &Page{
	//	Title: "TestPage",
	//	Body: []byte("This is a sample Page"),
	//}
	//p1.save()
	//p2, _ := loadPage("TestPage")
	//fmt.Println(string(p2.Body))
}

//保存文件
func (p *Page) save() error {
	//save方法作为*Page指针类型p的接收器
	filename := p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

//读取文件
func loadPage(title string) (*Page, error) {
	//如果需要处理错误，函数返回值类型也要加上error
	filename := title + ".txt"
	body, err := os.ReadFile(filename) //返回[]byte和error,如果不需要处理错误，用_去掉error
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	//title, err := getTitle(w, r)
	//title := r.URL.Path[len("/view/"):] //从r.URL.Path中获取页面标题
	// 路径用len("/view/")重新切片，因为页面标题需要去掉"/view/"
	p, err := loadPage(title) //调用函数，不需要处理错误，用_忽略
	if err != nil {
		//访问到不存在的页面，重定向到编辑页
		http.Redirect(w, r, "/edit/"+title, http.StatusNotFound)
		return
	}
	//fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	//title, err := getTitle(w, r)
	//if err != nil{
	//	return
	//}
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	//t, _ := template.ParseFiles("edit.html")
	////template.ParseFiles方法读取edit.html并返回一个*template.Template.
	//t.Execute(w, p)
	////执行模板并把被生成的HTML写到http.ResponseWriter
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	//title, err := getTitle(w, r)
	//if err != nil {
	//	return
	//}
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		//保存时出错返回错误状态码
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	//如果标题合法，FindStringSubmatch()返回一个nil
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("无效的页面标题！")
	}
	//title是第二个子表达式
	return m[2], nil
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	//装饰器，传入一个函数对象，这个函数对象的参数和Handler的参数一样，函数返回值类型是 http.HandlerFunc
	return func(w http.ResponseWriter, r *http.Request) {
		//返回一个函数，这个返回的函数就叫闭包
		//如果title合法，ResponseWriter, Request, 和 title会作为参数传递给fn并调用
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		//title是第二个表达式
		//fn就是调用的函数，这里就是viewHandler,editHandler和saveHandler
		fn(w, r, m[2])
	}
}
