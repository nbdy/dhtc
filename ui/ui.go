package ui

import (
	"dhtc/config"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/leekchan/gtf"
	"github.com/ostafen/clover/v2"
)

type Controller struct {
	Database      *clover.DB
	Configuration *config.Configuration
}

func loadTemplates() multitemplate.Render {
	renderer := multitemplate.New()

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
		tpl := template.Must(gtf.New(viewName).ParseFS(templates, append(includeFiles, "templates/view/"+viewFileName)...))
		renderer.Add(viewName, tpl)
	}

	return renderer
}

func RunWebServer(configuration *config.Configuration, database *clover.DB) {
	// gin.SetMode(gin.ReleaseMode)

	srv := gin.Default()
	srv.HTMLRender = loadTemplates()
	_ = srv.SetTrustedProxies(nil)

	uiCtrl := Controller{
		Database:      database,
		Configuration: configuration,
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

	css, _ := fs.Sub(static, "static/css")
	js, _ := fs.Sub(static, "static/js")

	srv.StaticFS("/css", http.FS(css))
	srv.StaticFS("/js", http.FS(js))

	err := srv.Run(configuration.Address)

	if err != nil {
		return
	}
}
