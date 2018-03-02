package main

import (
	"encoding/json"
	"fmt"
	"github.com/theckman/go-flock"
	"io/ioutil"
	"os"
	"path/filepath"
)

func GenConfigName() string {
	return filepath.Join(megaWorkspace, configName)
}

//Config functions
func SaveConfig(v interface{}) error {
	GL.Lock.Lock()
	fileLock := flock.NewFlock(fmt.Sprintf(urlformat, GenConfigName(), LockExt))
	fileLock.Lock()
	str := mResponse(v)
	pathoffile := GenConfigName()
	strbytes := []byte(str)
	err := ioutil.WriteFile(pathoffile, strbytes, 0700)
	strbytes = nil
	fileLock.Unlock()
	GL.Lock.Unlock()
	return err

}
func LoadConfig(targ interface{}) error {
	GL.Lock.Lock()
	pathoffile := GenConfigName()

	data, err := ioutil.ReadFile(pathoffile)
	if err != nil {
		return err
	}
	strdata := string(data)
	bts := []byte(strdata)
	err = json.Unmarshal(bts, targ)
	GL.Lock.Unlock()
	return err
}

func GenLogName(name string) string {
	return filepath.Join(megaWorkspace, logDirectory, name)
}

// Server request log functions
func SaveLog(name string, v interface{}) error {
	fileLock := flock.NewFlock(fmt.Sprintf(urlformat, GenLogName(name), LockExt))
	fileLock.Lock()
	str := mResponse(v)
	pathoffile := GenLogName(name)
	strbytes := []byte(str)
	err := ioutil.WriteFile(pathoffile, strbytes, 0700)
	strbytes = nil
	fileLock.Unlock()
	return err
}

func LoadLog(name string, targ interface{}) error {
	pathoffile := GenLogName(name)
	data, err := ioutil.ReadFile(pathoffile)
	if err != nil {
		return err
	}
	strdata := string(data)
	bts := []byte(strdata)
	err = json.Unmarshal(bts, targ)
	return err
}

func DeleteLog(name string) {
	pathoffile := filepath.Join(megaWorkspace, logDirectory, name)
	os.Remove(pathoffile)
}
