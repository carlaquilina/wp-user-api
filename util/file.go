package util

import (
	"errors"
	"os"
)

func DeleteFile(path string) error {
	// delete file
	var err = os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}
func BackUpFile(oldName, newName string) error {
	//check if file exists
	if !FileExists(oldName) {
		return errors.New("File does not exist: " + oldName)
	}
	//if not return error
	err := os.Rename(oldName, newName)
	if err != nil {
		return err
	}
	return nil
}
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
