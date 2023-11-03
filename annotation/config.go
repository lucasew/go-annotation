package annotation

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
    Meta struct {
        Description string `yaml:"description"`
    }
    Classes []struct{
        ID string `yaml:"id"`
        Name string `yaml:"name"`
        ShortName string `yaml:"short_name"`
        Type string `yaml:"string"`
        If map[string]string `yaml:"if"`
        Classes map[string]struct{
            Name string `yaml:"string"`
            Description string `yaml:"description"`
            Examples []string `yaml:"examples"`
        }
    }
}

func LoadConfig(filename string) (*Config, error) {
    var ret Config
    data, err := ioutil.ReadAll(filename)
    if err != nil {
        return nil, err
    }
    err = yaml.Unmarshal(data, &ret)
    if err != nil {
        return nil, err
    }
    return &ret, nil

}
