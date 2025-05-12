package logger

import (
	"io"
	"log"
	"os"
)

var (
	Info *log.Logger
	Err  *log.Logger
	Debug *log.Logger
)

func DoConsoleLog() {
	initLoggers(os.Stdout)
}

func LogToFile(filename string) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open server.log file ", ":", err)
	}
	log.Printf("All logs now are saved in logger/%s", filename)

	initLoggers(file)
}

func initLoggers(out io.Writer) {
	Info = log.New(out, "[INFO]: ", log.Ldate|log.Ltime)
	Err = log.New(out, "[ERROR]: ", log.Ldate|log.Ltime|log.Lshortfile)
	Debug = log.New(out, "[DEBUG]: ", log.Ldate|log.Ltime|log.Lshortfile)
}