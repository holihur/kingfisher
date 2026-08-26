package app

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"kingfisher/core/config"
)

func setupSPA(r *gin.Engine, cfg *config.Config) {
	if cfg.Server.StaticDir == "" {
		return
	}
	dist := cfg.Server.StaticDir
	r.Static("/assets", filepath.Join(dist, "assets"))
	r.Static("/vite.svg", filepath.Join(dist, "vite.svg"))
	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path
		if strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/swagger") || strings.HasPrefix(p, "/uploads") || strings.HasPrefix(p, "/metrics") {
			c.JSON(404, gin.H{"code": 404, "message": "not found"})
			return
		}
		fp := filepath.Join(dist, filepath.Clean(p))
		if rel, relErr := filepath.Rel(dist, fp); relErr != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			c.JSON(404, gin.H{"code": 404, "message": "not found"})
			return
		}
		if fi, err := os.Stat(fp); err == nil && !fi.IsDir() {
			c.File(fp)
			return
		}
		c.File(filepath.Join(dist, "index.html"))
	})
}
