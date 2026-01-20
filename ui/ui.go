package ui

import (
	"dhtc/config"
	"dhtc/db"
	"dhtc/notifier"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/leekchan/gtf"
)

type Controller struct {
	Database      db.Repository
	Configuration *config.Configuration
	Hub           *Hub
	Notifier      *notifier.Manager
}

func loadTemplates() multitemplate.Render {
	renderer := multitemplate.New()

	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	viewDirectory, _ := templates.ReadDir("templates/view")
	includeDirectory, _ := templates.ReadDir("templates/include")
	var includeFiles []string
	for _, includeFile := range includeDirectory {
		includeFiles = append(includeFiles, "templates/include/"+includeFile.Name())
	}

	for _, viewPath := range viewDirectory {
		if viewPath.IsDir() {
			continue
		}
		viewFileName := viewPath.Name()
		viewName := strings.TrimSuffix(viewFileName, ".html")
		tpl := template.Must(gtf.New(viewName).Funcs(funcMap).ParseFS(templates, append(includeFiles, "templates/view/"+viewFileName)...))
		renderer.Add(viewName, tpl)
	}

	return renderer
}

func (c *Controller) getCommonH(ctx *gin.Context) gin.H {
	return gin.H{
		"path":   ctx.FullPath(),
		"config": c.Configuration,
	}
}

func RunWebServer(configuration *config.Configuration, database db.Repository, hub *Hub, nManager *notifier.Manager) {
	// gin.SetMode(gin.ReleaseMode)

	srv := gin.Default()
	srv.HTMLRender = loadTemplates()
	_ = srv.SetTrustedProxies(nil)

	store := cookie.NewStore([]byte(configuration.SessionSecret))
	srv.Use(sessions.Sessions("dhtc-session", store))

	if configuration.AuthUser != "" && configuration.AuthPass != "" {
		srv.Use(gin.BasicAuth(gin.Accounts{
			configuration.AuthUser: configuration.AuthPass,
		}))
	}

	uiCtrl := Controller{
		Database:      database,
		Configuration: configuration,
		Hub:           hub,
		Notifier:      nManager,
	}

	srv.GET("", uiCtrl.Dashboard)
	srv.GET("/dashboard", uiCtrl.Dashboard)
	srv.GET("/search", uiCtrl.SearchGet)
	srv.POST("/search", uiCtrl.SearchPost)
	srv.GET("/discover", uiCtrl.DiscoverGet)
	srv.POST("/discover", uiCtrl.DiscoverPost)
	srv.GET("/watches", uiCtrl.WatchGet)
	srv.POST("/watches", uiCtrl.WatchPost)
	srv.GET("/blacklist", uiCtrl.BlacklistGet)
	srv.POST("/blacklist", uiCtrl.BlacklistPost)
	srv.GET("/settings", uiCtrl.SettingsGet)
	srv.POST("/settings", uiCtrl.SettingsPost)
	srv.GET("/trawl", uiCtrl.Trawl)
	srv.GET("/ws/trawl", func(c *gin.Context) {
		uiCtrl.HandleWebSocket(c.Writer, c.Request)
	})

	srv.GET("/download/transmission", uiCtrl.SendToTransmission)
	srv.GET("/download/aria2", uiCtrl.SendToAria2)
	srv.GET("/download/deluge", uiCtrl.SendToDeluge)
	srv.GET("/download/qbittorrent", uiCtrl.SendToQBittorrent)

	srv.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := srv.Group("/api")
	{
		api.GET("/search", uiCtrl.APISearch)
		api.GET("/stats", uiCtrl.APIStats)
		api.GET("/categories", uiCtrl.APICategories)
		api.GET("/latest", uiCtrl.APILatest)
	}

	css, _ := fs.Sub(static, "static/css")
	js, _ := fs.Sub(static, "static/js")

	srv.StaticFS("/css", http.FS(css))
	srv.StaticFS("/js", http.FS(js))

	err := srv.Run(configuration.Address)
	if err != nil {
		return
	}
}
