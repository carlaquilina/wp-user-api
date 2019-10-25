package logfile

import (
	"log"
	"os"

	"github.com/fatih/color"
)

type LogFile struct {
	File       *os.File
	PathToFile string
}

func InitialiseLogFile(pathToFile string) (LogFile, error) {
	f, err := os.Create(pathToFile)
	if err != nil {
		return LogFile{}, err
	}
	logFile := LogFile{File: f, PathToFile: pathToFile}
	return logFile, nil
}
func (f *LogFile) LogMessage(message string) {
	color := color.New(color.FgBlue)
	writeFile(message, f.File, true, color)
}
func (f *LogFile) LogMessageWithLog(message string) {
	color := color.New(color.FgBlue)
	writeFile(message, f.File, false, color)
	log.Println(message)
}
func (f *LogFile) LogMessageWithFatal(message string) {
	color := color.New(color.FgRed)
	writeFile(message, f.File, false, color)
	log.Fatal(message)
}

func writeFile(message string, file *os.File, echo bool, color *color.Color) {
	// write some text line-by-line to file
	_, err := file.WriteString(message + "\n")
	if err != nil {
		log.Fatal(err)
	}
	// save changes
	err = file.Sync()
	if err != nil {
		log.Fatal(err)
	}
	if echo {
		color.Println(message)
	}
}
