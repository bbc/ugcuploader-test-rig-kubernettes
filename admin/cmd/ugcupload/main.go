package main

import (
	"encoding/gob"
	"net/http"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"

	aws "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/aws"
	"github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/controller"
	"github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/kubernetes"
	"golang.org/x/sync/errgroup"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"

	"net/http/httputil"

	"github.com/magiconair/properties"

	types "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/types"
)

var control = controller.Controller{KubeOps: kubernetes.Operations{}, S3: aws.S3Operations{}}
var props = properties.MustLoadFile("/etc/ugcupload/loadtest.conf", properties.UTF8)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	control.KubeOps.Init()
}

var (
	g errgroup.Group
)

//SetNoCacheHeader middle to prevent browser caching
func SetNoCacheHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", "no-store")
		c.Next()
	}
}

func main() {

	server01 := &http.Server{
		Addr:         ":1323",
		Handler:      router01(),
		ReadTimeout:  15 * time.Minute,
		WriteTimeout: 15 * time.Minute,
		IdleTimeout:  15 * time.Minute,
	}

	server02 := &http.Server{
		Addr:         ":1232",
		Handler:      router02(),
		ReadTimeout:  15 * time.Minute,
		WriteTimeout: 15 * time.Minute,
		IdleTimeout:  15 * time.Minute,
	}

	g.Go(func() error {
		return server01.ListenAndServe()
	})

	g.Go(func() error {
		return server02.ListenAndServe()
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}

}

func router01() http.Handler {
	// Gin instance
	r := gin.Default()

	gob.Register(types.UgcLoadRequest{})
	store, _ := redis.NewStore(10, "tcp", "localhost:6379", "", []byte("secret"))
	r.Use(sessions.Sessions("mysession", store))
	r.Use(SetNoCacheHeader())

	r.LoadHTMLGlob(props.MustGet("web") + "/templates/*")
	r.GET("/", func(c *gin.Context) {

		session := sessions.Default(c)
		var ugcLoadRequest types.UgcLoadRequest
		if ulr := session.Get("ugcLoadRequest"); ulr != nil {
			ugcLoadRequest = ulr.(types.UgcLoadRequest)
		} else {
			ugcLoadRequest = types.UgcLoadRequest{}
		}
		control.AddMonitorAndDashboard(&ugcLoadRequest)
		control.AddTenants(&ugcLoadRequest)
		c.HTML(http.StatusOK, "index.tmpl", ugcLoadRequest)
		session.Clear()
		if err := session.Save(); err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("Unable to save the session")
		}

	})

	r.GET("/update", func(c *gin.Context) {

		session := sessions.Default(c)
		var ugcLoadRequest types.UgcLoadRequest
		if ulr := session.Get("ugcLoadRequest"); ulr != nil {
			ugcLoadRequest = ulr.(types.UgcLoadRequest)
		} else {
			ugcLoadRequest = types.UgcLoadRequest{}
		}
		control.AddMonitorAndDashboard(&ugcLoadRequest)
		control.AddTenants(&ugcLoadRequest)
		c.PureJSON(http.StatusOK, ugcLoadRequest)
		session.Clear()
		if err := session.Save(); err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("Unable to save the session")
		}

	})
	r.Static("/script", props.MustGet("web"))
	r.POST("/start-test", control.Upload)
	r.POST("/stop-test", control.StopTest)
	r.POST("/delete-tenant", control.DeleteTenant)
	r.GET("/tenantReport", control.S3Tenants)
	r.POST("/genReport", control.GenerateReport)

	r.GET("/tenants", ReverseProxy())
	return r
}

//ReverseProxy taken from here:https://github.com/gin-gonic/gin/issues/686
func ReverseProxy() gin.HandlerFunc {
	return func(c *gin.Context) {

		target, _ := url.Parse("http://127.0.0.1:1232")
		proxy := httputil.NewSingleHostReverseProxy(target)
		realDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			log.Println(req.URL)
			realDirector(req)
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func router02() http.Handler {
	// Gin instance
	r := gin.Default()
	r.GET("/tenants", control.AllTenants)
	return r
}
