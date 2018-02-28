package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

//Config functions
func SaveConfig(v interface{}) error {
	GL.Lock.Lock()
	str := mResponse(v)
	pathoffile := filepath.Join(megaWorkspace, configName)
	strbytes := []byte(str)
	err := ioutil.WriteFile(pathoffile, strbytes, 0700)
	strbytes = nil
	GL.Lock.Unlock()
	return err

}
func LoadConfig(targ interface{}) error {
	GL.Lock.Lock()
	pathoffile := filepath.Join(megaWorkspace, configName)

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

// Server request log functions
func SaveLog(name string, v interface{}) error {
	str := mResponse(v)
	pathoffile := filepath.Join(megaWorkspace, logDirectory, name)
	strbytes := []byte(str)
	err := ioutil.WriteFile(pathoffile, strbytes, 0700)
	strbytes = nil
	return err
}

func LoadLog(name string, targ interface{}) error {

	pathoffile := filepath.Join(megaWorkspace, logDirectory, name)

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
