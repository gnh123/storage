package main

import (
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/gnh123/storage"
	"github.com/gnh123/storage/cmd/benchmark"
	"github.com/guonaihong/clop"
	"github.com/guonaihong/gutil/file"
)

type Storage struct {
	Server              `clop:"subcommand" usage:"server sub command"`
	benchmark.Benchmark `clop:"subcommand" usage:"benchmark"`
}

type Server struct {
	Dir  string       `clop:"short;long" usage:"dir" valid:"required"`
	Size storage.Size `clop:"short;long;callback=ParseSize" usage:"Maximum capacity that can be stored, example:1G 1T" `
	s    storage.Storage
}

type query struct {
	Key string `form:"key"`
}

type data struct {
	Data []byte `json:"data"`
}

// clop的callback=ParseSize会调用
func (s *Server) ParseSize(val string) {
	size, err := file.ParseSize(val)
	if err != nil {
		fmt.Printf("parse size fail:%s\n", err)
		return
	}

	s.Size = storage.Size(size)
}

func (s *Server) createRaw(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{"code": 1, "message": err.Error()})
		return
	}

	index, err := s.s.Put(data)
	if err != nil {
		c.JSON(500, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 0, "message": "", "data": gin.H{"index": index}})
}

func (s *Server) create(c *gin.Context) {
	d := data{}
	err := c.ShouldBindJSON(&d)
	if err != nil {
		c.JSON(500, gin.H{"code": 1, "message": err.Error()})
		return
	}
	index, err := s.s.Put(d.Data)
	if err != nil {
		c.JSON(500, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 0, "message": "", "data": gin.H{"index": index}})
}

func (s *Server) delete(c *gin.Context) {

	var q query
	err := c.ShouldBindQuery(&q)
	if err != nil {
		c.JSON(500, gin.H{"code": 1, "message": err.Error()})
		return
	}

	s.s.Delete(q.Key)
	c.JSON(200, gin.H{"code": 0, "message": ""})
}

func (s *Server) get(c *gin.Context) {
	var q query
	err := c.ShouldBindQuery(&q)
	if err != nil {
		c.JSON(500, gin.H{"code": 1, "message": err.Error()})
		return
	}

	elem, ok, err := s.s.Get(q.Key)
	if err != nil {
		c.JSON(500, gin.H{"code": 0, "message": err.Error()})
		return
	}
	if !ok {
		c.JSON(500, gin.H{"code": 0, "message": "not found"})
		return

	}

	c.JSON(200, gin.H{"code": 0, "message": "", "data": elem})

}

func (s *Server) SubMain() {

	var err error
	r := gin.Default()

	s.s, err = storage.Open(s.Dir, s.Size)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	r.POST("/file", s.create)
	r.POST("/file/raw", s.createRaw)
	r.DELETE("/file", s.delete)
	r.GET("/file", s.get)

	r.Run()
}

func main() {

	var s Storage

	clop.Bind(&s)

}
