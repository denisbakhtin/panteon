package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/claudiu/gocron"
	"github.com/denisbakhtin/panteon/config"
	"github.com/denisbakhtin/panteon/controllers"
	"github.com/denisbakhtin/panteon/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/utrack/gin-csrf"
)

func main() {
	gob.Register(controllers.CartType{})

	mode := flag.String("mode", "debug", "Application mode: debug, release, test")
	flag.Parse()

	gin.SetMode(*mode)

	initLogger()
	config.LoadConfig()
	models.SetDB(config.GetConnectionString())
	models.AutoMigrate()

	//Periodic tasks
	gocron.Every(1).Day().Do(controllers.CreateXMLSitemap)
	gocron.Start()

	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	router := gin.Default()
	router.StaticFS("/public", http.Dir(config.PublicPath())) //better use nginx to serve assets (Cache-Control, Etag, fast gzip, etc)
	controllers.LoadTemplates(router)

	//setup sessions
	conf := config.GetConfig()
	store := memstore.NewStore([]byte(conf.SessionSecret))
	store.Options(sessions.Options{HttpOnly: true, MaxAge: 7 * 86400}) //Also set Secure: true if using SSL, you should though
	router.Use(sessions.Sessions("gin-session", store))
	router.Use(controllers.ContextData())

	//setup csrf protection
	router.Use(csrf.Middleware(csrf.Options{
		Secret: conf.SessionSecret,
		ErrorFunc: func(c *gin.Context) {
			logrus.Error("CSRF token mismatch")
			controllers.ShowErrorPage(c, 400, fmt.Errorf("CSRF token mismatch"))
			c.Abort()
		},
	}))

	router.GET("/", controllers.HomeGet)
	router.NoRoute(controllers.NotFound)
	router.NoMethod(controllers.MethodNotAllowed)

	if config.GetConfig().SignupEnabled {
		router.GET("/signup", controllers.SignUpGet)
		router.POST("/signup", controllers.SignUpPost)
	}
	router.GET("/signin", controllers.SignInGet)
	router.POST("/signin", controllers.SignInPost)
	router.GET("/signout", controllers.SignoutGet)

	router.GET("/pages/:idslug", controllers.PageGet)
	router.GET("/c/:idslug", controllers.CategoryGet)
	router.GET("/p/:idslug", controllers.ProductGet)
	router.GET("/rss", controllers.RssGet)

	router.GET("/cart", controllers.CartGet)
	router.POST("/cart/add/:id", controllers.CartAdd)
	router.POST("/cart/delete/:id", controllers.CartDelete)

	router.GET("/new_order", controllers.OrderNew)
	router.POST("/new_order", controllers.OrderCreate)
	router.GET("/confirm_order/:id", controllers.OrderConfirm)

	router.GET("/search", controllers.SearchGet)

	router.POST("/orderconsult", controllers.OrderConsultPost)

	authorized := router.Group("/admin")
	authorized.Use(controllers.AuthRequired())
	{
		authorized.GET("/", controllers.AdminGet)

		authorized.POST("/upload", controllers.UploadPost) //image upload

		authorized.GET("/users", controllers.UserIndex)
		authorized.GET("/new_user", controllers.UserNew)
		authorized.POST("/new_user", controllers.UserCreate)
		authorized.GET("/users/:id/edit", controllers.UserEdit)
		authorized.POST("/users/:id/edit", controllers.UserUpdate)
		authorized.POST("/users/:id/delete", controllers.UserDelete)

		authorized.GET("/pages", controllers.PageIndex)
		authorized.GET("/new_page", controllers.PageNew)
		authorized.POST("/new_page", controllers.PageCreate)
		authorized.GET("/pages/:id/edit", controllers.PageEdit)
		authorized.POST("/pages/:id/edit", controllers.PageUpdate)
		authorized.POST("/pages/:id/delete", controllers.PageDelete)

		authorized.GET("/menus", controllers.MenuIndex)
		authorized.GET("/new_menu", controllers.MenuNew)
		authorized.POST("/new_menu", controllers.MenuCreate)
		authorized.GET("/menus/:id/edit", controllers.MenuEdit)
		authorized.POST("/menus/:id/edit", controllers.MenuUpdate)
		authorized.POST("/menus/:id/delete", controllers.MenuDelete)

		authorized.GET("/categories", controllers.CategoryIndex)
		authorized.GET("/new_category", controllers.CategoryNew)
		authorized.POST("/new_category", controllers.CategoryCreate)
		authorized.GET("/categories/:id/edit", controllers.CategoryEdit)
		authorized.POST("/categories/:id/edit", controllers.CategoryUpdate)
		authorized.POST("/categories/:id/delete", controllers.CategoryDelete)

		authorized.GET("/products", controllers.ProductIndex)
		authorized.GET("/new_product", controllers.ProductNew)
		authorized.POST("/new_product", controllers.ProductCreate)
		authorized.GET("/products/:id/edit", controllers.ProductEdit)
		authorized.POST("/products/:id/edit", controllers.ProductUpdate)
		authorized.POST("/products/:id/delete", controllers.ProductDelete)

		authorized.POST("/new_image", controllers.ImageCreate)
		authorized.POST("/images/:id/delete", controllers.ImageDelete)

		authorized.GET("/settings", controllers.SettingIndex)
		authorized.GET("/new_setting", controllers.SettingNew)
		authorized.POST("/new_setting", controllers.SettingCreate)
		authorized.GET("/settings/:id/edit", controllers.SettingEdit)
		authorized.POST("/settings/:id/edit", controllers.SettingUpdate)
		authorized.POST("/settings/:id/delete", controllers.SettingDelete)

		authorized.GET("/advantages", controllers.AdvantageIndex)
		authorized.GET("/new_advantage", controllers.AdvantageNew)
		authorized.POST("/new_advantage", controllers.AdvantageCreate)
		authorized.GET("/advantages/:id/edit", controllers.AdvantageEdit)
		authorized.POST("/advantages/:id/edit", controllers.AdvantageUpdate)
		authorized.POST("/advantages/:id/delete", controllers.AdvantageDelete)

		authorized.GET("/orders", controllers.OrderIndex)
		authorized.GET("/orders/:id", controllers.OrderGet)
		authorized.POST("/orders/:id/delete", controllers.OrderDelete)
	}

	// Listen and server on 0.0.0.0:8081
	router.Run(":8081")
}

//initLogger initializes logrus logger with some defaults
func initLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	if gin.Mode() == gin.DebugMode {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
