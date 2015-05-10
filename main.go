package main

import (
	"fmt"
	"log"
	//"runtime"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var (
	workersCount = 1 //runtime.NumCPU()
	kLinksDir    = "latest"
	kContentDir  = "content"
)

// NOTE: `init` is called implicitly by runtime.
func init() {
}

//------------------------------------------------------------------------------

func main() {
	start := time.Now()
	fmt.Println(fibonacci(38))
	log.Printf("%2fs total\n", time.Since(start).Seconds())

	server := &Server{Dir: "data"}
	err := server.Start(":8080")
	if err != nil {
		log.Panic(err)
	}
}

type Server struct {
	Dir string
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		s.handlePost(w, r)
	} else {
		s.handleGet(w, r)
	}
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	// TODO: make error response in case of failure
	file, header, err := r.FormFile("key")
	if err != nil {
		log.Println("form", err)
		return
	}
	defer file.Close()

	filename := r.URL.Path
	log.Println("handling POST:", filename, header)
	targetPath := s.getContentPath(filename)
	content, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Println("error to create file:", err)
		return
	}
	io.Copy(content, file)
	content.Close()

	err = renameSymlink(s.getReadPath(filename), targetPath)
	if err != nil {
		log.Println("error to update file:", err)
	}
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	// Model some delay.
	fibonacci(38)

	filename := r.URL.Path
	http.ServeFile(w, r, s.getReadPath(filename))
}

func (s *Server) getReadPath(filename string) string {
	return filepath.Join(s.Dir, "latest", filename)
}

func (s *Server) getContentPath(filename string) string {
	return filepath.Join("..", "content", filename+timestamp())
}

// Helper functions.

func renameSymlink(link, target string) error {
	return exec.Command("ln", "-s", target, link).Run()
}

func timestamp() string {
	return time.Now().UTC().Format("20060102150405")
}

func fibonacci(n int) int {
	if n <= 1 {
		return 1
	}
	return fibonacci(n-1) + fibonacci(n-2)
}
