package annotation

import (
	"io"
    "os"
    "fmt"

	"gopkg.in/yaml.v3"
)

type Config struct {
    Meta struct {
        Description string `yaml:"description"`
    } `yaml:"meta"`
    Tasks map[string]*ConfigTask `yaml:"tasks"`
    Authentication map[string]*ConfigAuth `yaml:"auth"`
}

type ConfigAuth struct {
    Password string `yaml:"password"`
}

type ConfigTask struct {
    Name string `yaml:"name"`
    ShortName string `yaml:"short_name"`
    Type string `yaml:"type"`
    If map[string]string `yaml:"if"`
    Classes map[string] *ConfigClass `yaml:"classes"`
}

type ConfigClass struct {
    Name string `yaml:"string"`
    Description string `yaml:"description"`
    Examples []string `yaml:"examples"`
}

func LoadConfig(filename string) (*Config, error) {
    var ret Config
    f, err := os.Open(filename)
    defer f.Close()
    if err != nil {
        return nil, err
    }
    data, err := io.ReadAll(f)
    if err != nil {
        return nil, err
    }
    err = yaml.Unmarshal(data, &ret)
    if err != nil {
        return nil, err
    }
    for taskName := range ret.Tasks {
        if ret.Tasks[taskName].Type == "" {
            ret.Tasks[taskName].Type = "class"
        }
        if ret.Tasks[taskName].Classes == nil {
            ret.Tasks[taskName].Classes = getClassesFromClassType(ret.Tasks[taskName].Type)
        }
        if ret.Tasks[taskName].Classes == nil {
            return nil, fmt.Errorf("task %s does not have any classes or a compatible type", taskName)
        }
    }
    if len(ret.Authentication) == 0 {
        return nil, fmt.Errorf("no users specified")
    }
    for user := range ret.Authentication {
        if ret.Authentication[user].Password == "" {
            return nil, fmt.Errorf("user %s has a null password", user)
        }
    }
    return &ret, nil
}

func getClassesFromClassType(classType string) map[string]*ConfigClass {
    switch classType {
    case "boolean":
        return map[string]*ConfigClass{
            "true": {
                Name: "Yes",
            },
            "false": {
                Name: "No",
            },
        }
    case "rotation":
        return map[string]*ConfigClass{
            "ok": {
                Name: "OK",
                Description: "Not rotated",
            },
            "h_inv": {
                Name: "Invert X",
                Description: "Invert in horizontal axis",
            },
            "v_inv": {
                Name: "Invert Y",
                Description: "Invert in vertical axis",
            },
            "+90": {
                Name: "+90deg",
                Description: "Rotate 90 degrees horary",
            },
            "-90": {
                Name: "-90deg",
                Description: "Rotate 90 degrees antihorary",
            },
            "180": {
                Name: "180deg",
                Description: "Rotate 180 degrees",
            },
        }
    default:
        return nil
    }
}
