package logging

import (
	"fmt"
	"log"
	"os"
)

func LogToFile(message interface{}) {
	if message != nil {
		file, fatal_error := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

		if fatal_error != nil {
			log.Fatalf("error opening file: %v", fatal_error)
		}
		defer file.Close()

		log.SetOutput(file)
		log.Println(log.Lshortfile, message)
	}
}

func HandleErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
