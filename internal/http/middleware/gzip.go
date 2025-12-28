package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to decode gzip body"})
				return
			}
			defer reader.Close()
			c.Request.Body = io.NopCloser(reader)
		}

		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if !(strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "text/html")) {
			c.Next()
			return
		}

		gw := &gzipResponseWriter{
			ResponseWriter: c.Writer,
			writer:         nil,
		}
		c.Writer = gw
		c.Header("Content-Encoding", "gzip")

		c.Next()

		if gw.writer != nil {
			gw.writer.Close()
		}
	}
}

type gzipResponseWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipResponseWriter) Write(data []byte) (int, error) {
	if g.writer == nil {
		g.writer = gzip.NewWriter(g.ResponseWriter)
	}
	return g.writer.Write(data)
}
