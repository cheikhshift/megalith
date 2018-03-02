package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

func GenConfigName() string {
	return filepath.Join(megaWorkspace, configName)
}

//Config functions
func SaveConfig(v interface{}) error {
	ShouldLock()
	str := mResponse(v)
	pathoffile := GenConfigName()
	strbytes := []byte(str)
	err := ioutil.WriteFile(pathoffile, strbytes, 0700)
	strbytes = nil
	ShouldUnlock()
	return err

}
func LoadConfig(targ interface{}) error {
	ShouldLock()
	pathoffile := GenConfigName()
	data, err := ioutil.ReadFile(pathoffile)
	if err != nil {
		return err
	}
	strdata := string(data)
	bts := []byte(strdata)
	err = json.Unmarshal(bts, targ)
	ShouldUnlock()
	return err
}

func GenLogName(name string) string {
	return filepath.Join(megaWorkspace, logDirectory, name)
}

// Server request log functions
func SaveLog(name string, v interface{}) error {
	str := mResponse(v)
	pathoffile := GenLogName(name)
	strbytes := []byte(str)
	err := ioutil.WriteFile(pathoffile, strbytes, 0700)
	strbytes = nil
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
