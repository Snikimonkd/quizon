package logger

import "log"

var reset = "\033[0m"

var red = "\033[31m"
var green = "\033[32m"
var yellow = "\033[33m"

func Infof(format string, v ...any) {
	if len(v) == 0 {
		log.Print(green + "[INFO] " + reset + format)
	} else {
		log.Printf(green+"[INFO] "+reset+format, v)
	}
}

func Warnf(format string, v ...any) {
	if len(v) == 0 {
		log.Print(yellow + "[WARN] " + reset + format)
	} else {
		log.Printf(yellow+"[WARN] "+reset+format, v)
	}
}

func Errorf(format string, v ...any) {
	if len(v) == 0 {
		log.Print(red + "[ERROR] " + reset + format)
	} else {
		log.Printf(red+"[ERROR] "+reset+format, v)
	}
}

func Fatalf(format string, v ...any) {
	if len(v) == 0 {
		log.Fatal(red + "[FATAL] " + reset + format)
	} else {
		log.Fatalf(red+"[FATAL] "+reset+format, v)
	}
}
