package server

import (
	"database/sql"
	"fmt"
	"os"

	"errors"

	"encoding/json"
	"log"
	"net/http"

	"github.com/MSC-XDU/playground/share"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	_ "github.com/mattes/migrate/driver/mysql"
	"github.com/mattes/migrate/migrate"
)

var dbUser, dbPassword, dbAddr, dbName string

var db *sql.DB

func init() {
	dbUser = os.Getenv("SHARE_DB_USER")
	dbPassword = os.Getenv("SHARE_DB_PASSWORD")
	dbAddr = os.Getenv("DB_ADDR")
	dbName = os.Getenv("DB_NAME")

	if dbUser == "" || dbPassword == "" || dbAddr == "" || dbName == "" {
		panic("没有配置数据库账号密码")
	}

	dbURL := fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPassword, dbAddr, dbName)

	var err error
	db, err = sql.Open("mysql", dbURL)
	if err != nil {
		panic(err)
	}

	log.Println("开始进行数据库迁移")
	errs, ok := migrate.UpSync("mysql://"+dbURL, "./server/migrations")
	if !ok {
		log.Fatalf("%+v", errs)
	}
	log.Println("数据库迁移完成")
}

var (
	ErrMissParam = errors.New("miss param")
	ErrBadType   = errors.New("code type invalid")
)

type LangType uint8

const (
	TypeGo LangType = iota
	TypePy
	TypeC
	TypeErr
)

func ItoLangType(t string) (LangType, error) {
	var lt LangType
	switch t {
	case "go", "golang", "Go", "Golang":
		lt = TypeGo
	case "python", "Py", "Python", "py":
		lt = TypePy
	case "c", "C":
		lt = TypeC
	default:
		return TypeErr, ErrBadType
	}

	return lt, nil
}

func (t LangType) String() string {
	switch t {
	case TypeGo:
		return "Go"
	case TypePy:
		return "Python"
	case TypeC:
		return "C"
	default:
		return ""
	}
}

func saveCode(url, code string, t LangType) error {
	id, err := share.IDtoI(url)
	if err != nil {
		log.Printf("分享保存代码ID错误: %s", err.Error())
		return err
	}

	_, err = db.Exec("INSERT INTO code_share (id, code, type, url) VALUES (?, ?, ?, ?)", id, code, t, url)
	if err != nil {
		log.Printf("分享保存代码数据库错误: %s", err.Error())
		return err
	}

	return nil
}

// 生成一个新的 ID，将用户代码存入数据库并返回 ID
func SaveCode(code string, t LangType) (string, error) {
	url, err := share.NewID()
	if err != nil {
		return "", err
	}

	return url, saveCode(url, code, t)
}

// 根据用户的分享 ID 获取保存的代码和代码语言类型
func GetCode(url string) (code string, t LangType, err error) {
	id, err := share.IDtoI(url)
	if err != nil {
		log.Printf("分享获取代码ID错误: %s", err.Error())
		return "", 0, err
	}

	result := db.QueryRow("SELECT code, type FROM code_share WHERE id = ?", id)
	err = result.Scan(&code, &t)
	if err != nil {
		log.Printf("分享获取代码数据库错误: %s", err.Error())
	}
	return
}

func SaveCodeHandle(w http.ResponseWriter, req *http.Request) {
	var code, t string
	var err error

	if c := req.MultipartForm.Value["code"]; len(c) > 0 {
		code = c[0]
	} else {
		err = ErrMissParam
	}

	if c := req.MultipartForm.Value["type"]; len(c) > 0 {
		t = c[0]
	} else {
		err = ErrMissParam
	}

	langType, err := ItoLangType(t)
	if len(code) == 0 {
		err = ErrMissParam
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	url, err := SaveCode(code, langType)
	if err != nil {
		log.Printf("保存错误: %s", err.Error())
		http.Error(w, "服务器错误，稍后尝试", http.StatusInternalServerError)
		return
	}

	type resp struct {
		URL string `json:"url"`
	}

	r, _ := json.Marshal(resp{URL: url})

	w.Header().Set("Content-Type", "application/json")
	w.Write(r)
}

func GetCodeHandle(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	url := vars["id"]
	code, t, err := GetCode(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	editorTmpl.Execute(w, editorData{Code: code, Mode: t.String(), ModeSelect: false})
}
