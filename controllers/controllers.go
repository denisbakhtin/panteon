package controllers

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/denisbakhtin/panteon/config"
	"github.com/denisbakhtin/panteon/models"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

const userIDKey = "UserID"

var tmpl *template.Template

//Option represents select option entry
type Option struct {
	Value string
	Text  string
}

//DefaultH returns common to all pages template data
func DefaultH(c *gin.Context) gin.H {
	return gin.H{
		"Title":           "", //page title:w
		"Context":         c,
		"Csrf":            csrf.GetToken(c),
		"MetaKeywords":    "",
		"MetaDescription": "",
	}
}

//LoadTemplates loads templates from views directory to gin engine
func LoadTemplates(router *gin.Engine) {
	router.SetFuncMap(getFuncMap())
	router.LoadHTMLGlob("views/**/*")
}

func getFuncMap() template.FuncMap {
	return template.FuncMap{
		"isActiveLink":         isActiveLink,
		"stringInSlice":        stringInSlice,
		"formatDateTime":       formatDateTime,
		"now":                  now,
		"activeUserEmail":      activeUserEmail,
		"activeUserName":       activeUserName,
		"activeUserID":         activeUserID,
		"isUserAuthenticated":  isUserAuthenticated,
		"signUpEnabled":        SignUpEnabled,
		"noescape":             noescape,
		"topLevelMenuItems":    topLevelMenuItems,
		"refEqUint":            refEqUint,
		"selectableCategories": selectableCategories,
		"topLevelCategories":   topLevelCategories,
		"userRoles":            userRoles,
		"userRole":             userRole,
		"recommended":          recommended,
		"getSetting":           getSetting,
		"ourWorks":             ourWorks,
		"ourAdvantages":        ourAdvantages,
		"isNotBlank":           isNotBlank,
		"tel":                  tel,
		"productTitles":        productTitles,
		"cssVersion":           cssVersion,
		"jsVersion":            jsVersion,
		"domain":               domain,
		"isAdmin":              isAdmin,
		"isManager":            isManager,
		"isMember":             isMember,
		"safeURL":              safeURL,
	}
}

//atouint converts string to uint, returns 0 if error
func atouint(s string) uint {
	i, _ := strconv.ParseUint(s, 10, 32)
	return uint(i)
}

//atouint64 converts string to uint64, returns 0 if error
func atouint64(s string) uint64 {
	i, _ := strconv.ParseUint(s, 10, 64)
	return i
}

//isActiveLink checks uri against currently active (uri, or nil) and returns "active" if they are equal
func isActiveLink(c *gin.Context, uri string) string {
	if c != nil && c.Request.RequestURI == uri {
		return "active"
	}
	return ""
}

//formatDateTime prints timestamp in human format
func formatDateTime(t time.Time) string {
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

//stringInSlice returns true if value is in list slice
func stringInSlice(value string, list []string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

//now returns current timestamp
func now() time.Time {
	return time.Now()
}

//activeUserEmail returns currently authenticated user email
func activeUserEmail(c *gin.Context) string {
	user := activeUser(c)
	if user != nil {
		return user.Email
	}
	return ""
}

//activeUserName returns currently authenticated user name
func activeUserName(c *gin.Context) string {
	user := activeUser(c)
	if user != nil {
		return fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}
	return ""
}

//activeUserID returns currently authenticated user ID
func activeUserID(c *gin.Context) uint64 {
	user := activeUser(c)
	if user != nil {
		return user.ID
	}
	return 0
}

//activeUserRole returns currently authenticated user role
func activeUser(c *gin.Context) *models.User {
	if c != nil {
		u, _ := c.Get("User")
		if user, ok := u.(*models.User); ok {
			return user
		}
	}
	return nil
}

//isUserAuthenticated returns true is user is authenticated
func isUserAuthenticated(c *gin.Context) bool {
	user := activeUser(c)
	return user != nil
}

func isAdmin(c *gin.Context) bool {
	user := activeUser(c)
	return user != nil && user.IsAdmin()
}

func isManager(c *gin.Context) bool {
	user := activeUser(c)
	return user != nil && user.IsManager()
}

func isMember(c *gin.Context) bool {
	user := activeUser(c)
	return user != nil && user.IsMember()
}

//SignUpEnabled returns true if sign up is enabled by config
func SignUpEnabled() bool {
	return config.GetConfig().SignupEnabled
}

//noescape unescapes html content
func noescape(content string) template.HTML {
	return template.HTML(content)
}

//top level menu items
func topLevelMenuItems() []models.MenuItem {
	db := models.GetDB()
	var menus []models.MenuItem
	db.Preload("Children").Order("ord asc").Find(&menus, "parent_id is null")
	return menus
}

//refEqUint checks if *uint64 equals uint64
func refEqUint(ref *uint64, val uint64) bool {
	if ref == nil {
		return false
	}
	return *ref == val
}

//selectableCategories returns a slice of categories available for selection by products
func selectableCategories() []models.Category {
	db := models.GetDB()
	var categories []models.Category
	db.Where("NOT EXISTS(select 1 from categories as c where c.parent_id = categories.id)").Order("title").Find(&categories)
	return categories
}

func topLevelCategories() []models.Category {
	db := models.GetDB()
	var categories []models.Category
	db.Preload("Children").Order("ord asc").Find(&categories, "parent_id is null")
	return categories
}

func userRoles() []Option {
	return []Option{Option{Value: models.MEMBER, Text: "Покупатель"}, Option{Value: models.MANAGER, Text: "Менеджер"}, Option{Value: models.ADMIN, Text: "Администратор"}}
}

func userRole(role string) string {
	switch role {
	case models.MEMBER:
		return "Покупатель"
	case models.MANAGER:
		return "Менеджер"
	case models.ADMIN:
		return "Администратор"
	default:
		return "Неизвестно"
	}
}

func getSetting(code string) string {
	db := models.GetDB()
	setting := models.Setting{}
	db.Where("code = ?", code).First(&setting)
	return setting.Value
}

func recommended() []models.Product {
	db := models.GetDB()
	var products []models.Product
	db.Where("recommended = true and published = true").Find(&products)
	return products
}

func ourWorks() []models.Product {
	db := models.GetDB()
	var products []models.Product
	db.Where("published = true and category_id = ?", getSetting("our_works")).Find(&products)
	return products
}

func ourAdvantages() []models.Advantage {
	db := models.GetDB()
	var advantages []models.Advantage
	db.Order("id").Find(&advantages)
	return advantages
}

func isNotBlank(content string) bool {
	return len(content) > 0 && content != "<p>&nbsp;</p>"
}

func tel(content string) string {
	reg, err := regexp.Compile("[^\\+0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	processedString := reg.ReplaceAllString(content, "")
	return processedString
}

func productTitles() []string {
	db := models.GetDB()
	var titles []string
	db.Model(&models.Product{}).Where("published = true").Order("title").Pluck("title", &titles)
	return titles
}

func cssVersion() string {
	return fileVersion(path.Join(config.GetConfig().Public, "assets", "main.css"))
}

func jsVersion() string {
	return fileVersion(path.Join(config.GetConfig().Public, "assets", "application.js"))
}

func fileVersion(path string) string {
	file, err := os.Stat(path)
	if err != nil {
		return timeToString(time.Now())
	}
	modified := file.ModTime()
	return timeToString(modified)
}

func timeToString(t time.Time) string {
	return fmt.Sprintf("%04d%02d%02d-%02d%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
}

func domain() string {
	return config.GetConfig().Domain
}

//panelEntryURL returns an entry point for authenticated users
func panelEntryURL(user models.User) string {
	url := "/"
	switch user.Role {
	case models.ADMIN:
		url = "/admin/orders"
	case models.MANAGER:
		url = "/manager/orders"
	default:
		url = "/"
	}
	return url
}

func safeURL(url string) template.URL {
	return template.URL(url)
}
