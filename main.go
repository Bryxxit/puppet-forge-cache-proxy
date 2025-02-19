package main

import (
	"bytes"
	"flag"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

var (
	cacheDir string
	port     string
)

func main() {
	flag.StringVar(&cacheDir, "cacheDir", "./cache", "Directory to store cached files")
	flag.StringVar(&port, "port", "8080", "Port to run the proxy server on")
	flag.Parse()

	r := gin.Default()

	r.Any("/*proxyPath", handleProxy)

	r.Run(":" + port)
}

func handleProxy(c *gin.Context) {
	proxyPath := c.Param("proxyPath")
	proxyURL := "https://forgeapi.puppet.com" + proxyPath

	if filepath.HasPrefix(proxyPath, "/v3/files") {
		if serveFromCache(c, proxyPath) {
			return
		}
	}

	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	parsedURL.RawQuery = c.Request.URL.RawQuery

	proxyReq, err := http.NewRequest(c.Request.Method, parsedURL.String(), c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	proxyReq.Header = c.Request.Header

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if filepath.HasPrefix(proxyPath, "/v3/files") {
		cacheAndServeFile(c, resp, proxyPath)
		return
	}

	for k, v := range resp.Header {
		c.Header(k, v[0])
	}

	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

func serveFromCache(c *gin.Context, proxyPath string) bool {
	cacheFilePath := filepath.Join(cacheDir, filepath.Base(proxyPath))

	if _, err := os.Stat(cacheFilePath); err == nil {
		c.File(cacheFilePath)
		return true
	}

	return false
}

func cacheAndServeFile(c *gin.Context, resp *http.Response, proxyPath string) {
	cacheFilePath := filepath.Join(cacheDir, filepath.Base(proxyPath))

	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out, err := os.Create(cacheFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer out.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	respBody := buf.Bytes()

	out.Write(respBody)
	c.Writer.Write(respBody)
}
