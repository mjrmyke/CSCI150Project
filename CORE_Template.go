package main

import (
	"bytes"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"

	"regexp"

	"github.com/Esseh/retrievable"
	humanize "github.com/dustin/go-humanize" // russross markdown parser
	"golang.org/x/net/context"
	"google.golang.org/appengine"
)

// Global Template file.
var tpl *template.Template

func init() {
	// Tie functions into template here with ... "functionName":theFunction,
	funcMap := template.FuncMap{
		"getAvatarURL":  getAvatarURL,
		"getUser":       GetUserFromID,
		"humanize":      humanize.Time,
		"humanizeSize":  humanize.Bytes,
		"monthfromtime": monthfromtime,
		"yearfromtime":  yearfromtime,
		"dayfromtime":   dayfromtime,
		"findsvg":       FindSVG,
		"findtemplate":  FindTemplate,
		"inc":           Inc,
		"addCtx":        addCtx,
		"getDate":       getDate,
		// "isOwner":       isOwner,
		"parse": parse,
	} // Load up all templates.
	tpl = template.New("").Funcs(funcMap)
	tpl = template.Must(tpl.ParseGlob("templates/*"))

}

func ServeTemplateWithParams(res http.ResponseWriter, templateName string, params interface{}) error {
	return tpl.ExecuteTemplate(res, templateName, &params)
}

// Header Data,
// Present in most template executions. (Unless it's an internal it should be assumed to be used.)
type HeaderData struct {
	ShowLogin    bool
	ShowRegister bool
	Ctx          context.Context
	User         *User
	CurrentPath  string
}

// Constructs the header.
// As the header gets more complex(such as capturing the current path)
// the need for such a helper function increases.
func MakeHeader(res http.ResponseWriter, req *http.Request, login, register bool) *HeaderData {
	u, _ := GetUserFromSession(req)
	oldCookie, err := GetCookieValue(req, "session")
	if err == nil {
		MakeCookie(res, "session", oldCookie)
	}
	redirectURL := req.URL.Path[1:]
	if redirectURL == "login" || redirectURL == "register" || redirectURL == "elevatedlogin" {
		redirectURL = req.URL.Query().Get("redirect")
	}
	return &HeaderData{
		login, register, appengine.NewContext(req), u, redirectURL,
	}
}

/// Parses markdown to produce HTML.
func parse(inp string) string {
	// inp = strings.Replace(inp, `- [ ]`, `- <input type="checkbox">`, -1)         // Set Unchecked Checkbox
	// inp = strings.Replace(inp, `- [x]`, `- <input type="checkbox" checked>`, -1) // Set Checked Checkbox
	data := []byte(inp)                                    // Convert to Byte
	regex, _ := regexp.Compile("[sS][cC][rR][iI][pP][tT]") // Escape Script Tag
	data = regex.ReplaceAll(data, []byte("&#115;&#99;&#114;&#105;&#112;&#116;"))
	regex2, _ := regexp.Compile("[iI][fF][rR][aA][mM][eE]") // Escape Iframe Tag
	data = regex2.ReplaceAll(data, []byte("&#105;&#102;&#114;&#97;&#109;&#101;"))
	// data = blackfriday.MarkdownCommon(data) // Get Common Markdown from russross's parser
	return string(data)
}

func Inc(inp string) string {
	i, _ := strconv.ParseInt(inp, 10, 64)
	return strconv.FormatInt(i+1, 10)
}

//Finds corresponding SVG template
func FindSVG(name string) (ret template.HTML, err error) {
	buf := bytes.NewBuffer([]byte{})
	err = tpl.ExecuteTemplate(buf, ("svg-" + name), nil)
	ret = template.HTML(buf.String())
	return
}

func FindTemplate(name string) (ret template.HTML, err error) {
	buf := bytes.NewBuffer([]byte{})
	err = tpl.ExecuteTemplate(buf, (name), nil)
	ret = template.HTML(buf.String())
	return
}

type contextData struct {
	Ctx  context.Context
	Data interface{}
}

func addCtx(ctx context.Context, data interface{}) *contextData {
	return &contextData{
		Ctx:  ctx,
		Data: data,
	}
}

func getDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// Gets the Avatar URL
func getAvatarURL(userID retrievable.IntID) string {
	return "https://storage.googleapis.com/" + gcsBucket + "/" + getAvatarPath(int64(userID))
}

//gets the Year from a submitted time.Time
func yearfromtime(t time.Time) int {
	return t.Year()
}

//gets the Month from a submitted time.Time
func monthfromtime(t time.Time) time.Month {
	return t.Month()
}

//gets the Day from a submitted time.Time
func dayfromtime(t time.Time) int {
	return t.Day()
}

// Gets all the notes made my the user.
func GetAllNotes(ctx context.Context, userID int64) ([]Course, error) {
	q := datastore.NewQuery("Note").Filter("OwnerID =", userID)
	res := []Note{}
	keys, err := q.GetAll(ctx, &res)
	if err != nil {
		return nil, err
	}
	for i, key := range keys {
		res[i].StoreKey(key)
	}
	return res, nil
}
