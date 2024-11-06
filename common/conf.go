package common

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
)

type Config struct {
    PVM_PATH          string `json:"pvm_path"`
    PVM_VERSIONS_PATH string `json:"pvm_versions_path"`
}

func InitConfigFile(baseDir string) (*Config) {
    confPath := filepath.Join(baseDir, "pvm.json")

    conf := new(Config)
    err := conf.load(confPath)
    if err == nil {
        return conf
    }
    conf.PVM_PATH = filepath.ToSlash(filepath.Join(baseDir, "pvm.exe"))
    conf.PVM_VERSIONS_PATH = filepath.ToSlash(filepath.Join(baseDir, "versions"))

    err = conf.Save(confPath)
    if err != nil {
        Log.Error(err)
    }

    return conf
}


func (c *Config) load(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        Log.Error(err)
        return err
    }
    defer file.Close()
    decoder := json.NewDecoder(file)
    err = decoder.Decode(c)
    if err != nil {
        Log.Error(err)
    }
    return err
}

func (c *Config) ToJSON() (string, error) {
    jsonBin, err := json.Marshal(c)
    if err != nil {
        return "", err
    }
    var str bytes.Buffer
    _ = json.Indent(&str, jsonBin, "", "  ")
    return str.String(), nil
}

func (c *Config) Save(saveAs string) error {
    file, err := os.Create(saveAs)
    if err != nil {
        Log.Error(err)
        return err
    }
    defer file.Close()
    data, err := json.MarshalIndent(c, "", "    ")
    if err != nil {
        Log.Error(err)
        return err
    }
    _, err = file.Write(data)
    if err != nil {
        Log.Error(err)
    }
    return err
}
