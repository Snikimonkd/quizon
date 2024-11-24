package logger

import (
	"log"
)

const (
	reset  string = "\033[0m"
	red    string = "\033[31m"
	green  string = "\033[32m"
	yellow string = "\033[33m"
)

func Infof(format string, v ...any) {
	if len(v) == 0 {
		log.Print(green + "[INFO] " + reset + format)
		return
	}

	log.Printf(green+"[INFO] "+reset+format, v)
}

func Warnf(format string, v ...any) {
	if len(v) == 0 {
		log.Print(yellow + "[WARN] " + reset + format)
		return
	}

	log.Printf(yellow+"[WARN] "+reset+format, v)
}

func Errorf(format string, v ...any) {
	if len(v) == 0 {
		log.Print(red + "[ERROR] " + reset + format)
		return
	}

	log.Printf(red+"[ERROR] "+reset+format, v)
}

func Fatalf(format string, v ...any) {
	if len(v) == 0 {
		log.Fatal(red + "[FATAL] " + reset + format)
		return
	}

	log.Fatalf(red+"[FATAL] "+reset+format, v)
}
