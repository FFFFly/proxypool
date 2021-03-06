package api

import (
	"html/template"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/zu1k/proxypool/config"
	binhtml "github.com/zu1k/proxypool/internal/bindata/html"
	"github.com/zu1k/proxypool/internal/cache"
	"github.com/zu1k/proxypool/pkg/provider"
)

const version = "v0.3.4"

var router *gin.Engine

func setupRouter() {
	gin.SetMode(gin.ReleaseMode)
	router = gin.New()
	router.Use(gin.Recovery())
	temp, err := loadTemplate()
	if err != nil {
		panic(err)
	}
	router.SetHTMLTemplate(temp)

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/index.html", gin.H{
			"domain":               config.Config.Domain,
			"getters_count":        cache.GettersCount,
			"all_proxies_count":    cache.AllProxiesCount,
			"ss_proxies_count":     cache.SSProxiesCount,
			"ssr_proxies_count":    cache.SSRProxiesCount,
			"vmess_proxies_count":  cache.VmessProxiesCount,
			"trojan_proxies_count": cache.TrojanProxiesCount,
			"useful_proxies_count": cache.UsefullProxiesCount,
			"last_crawl_time":      cache.LastCrawlTime,
			"version":              version,
		})
	})

	router.GET("/clash", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/clash.html", gin.H{
			"domain": config.Config.Domain,
		})
	})

	router.GET("/surge", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/surge.html", gin.H{
			"domain": config.Config.Domain,
		})
	})

	router.GET("/clash/config", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/clash-config.yaml", gin.H{
			"domain": config.Config.Domain,
		})
	})

	router.GET("/surge/config", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/surge.conf", gin.H{
			"domain": config.Config.Domain,
		})
	})

	router.GET("/clash/proxies", func(c *gin.Context) {
		proxyTypes := c.DefaultQuery("type", "")
		proxyCountry := c.DefaultQuery("c", "")
		proxyNotCountry := c.DefaultQuery("nc", "")
		text := ""
		if proxyTypes == "" && proxyCountry == "" && proxyNotCountry == "" {
			text = cache.GetString("clashproxies")
			if text == "" {
				proxies := cache.GetProxies("proxies")
				clash := provider.Clash{Proxies: proxies}
				text = clash.Provide()
				cache.SetString("clashproxies", text)
			}
		} else if proxyTypes == "all" {
			proxies := cache.GetProxies("allproxies")
			clash := provider.Clash{
				Proxies:    proxies,
				Types:      proxyTypes,
				Country:    proxyCountry,
				NotCountry: proxyNotCountry,
			}
			text = clash.Provide()
		} else {
			proxies := cache.GetProxies("proxies")
			clash := provider.Clash{
				Proxies:    proxies,
				Types:      proxyTypes,
				Country:    proxyCountry,
				NotCountry: proxyNotCountry,
			}
			text = clash.Provide()
		}
		c.String(200, text)
	})
	router.GET("/surge/proxies", func(c *gin.Context) {
		text := cache.GetString("surgeproxies")
		if text == "" {
			proxies := cache.GetProxies("proxies")
			surge := provider.Surge{Proxies: proxies}
			text = surge.Provide()
			cache.SetString("surgeproxies", text)
		}
		c.String(200, text)
	})
}

func Run() {
	setupRouter()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}

func loadTemplate() (t *template.Template, err error) {
	t = template.New("")
	for _, fileName := range binhtml.AssetNames() {
		data := binhtml.MustAsset(fileName)
		t, err = t.New(fileName).Parse(string(data))
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}
