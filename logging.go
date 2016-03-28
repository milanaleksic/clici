package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type blackHoleWriter struct {
}

func (w *blackHoleWriter) Write(p []byte) (n int, err error) {
	err = errors.New("black hole writer")
	return
}

var logFile *os.File

func setupLog() {
	if options.Application.DoLog {
		fmt.Println("using " + filepath.Join(filepath.Dir(os.Args[0]), "clici.log"))
		var err error
		logFile, err = os.OpenFile(filepath.Join(filepath.Dir(os.Args[0]), "clici.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Error opening log file: %v", err)
		}
		log.SetOutput(logFile)
	} else {
		log.SetOutput(&blackHoleWriter{})
	}
}
