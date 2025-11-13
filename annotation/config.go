package annotation

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Meta struct {
		Description string `yaml:"description"`
	} `yaml:"meta"`
	Tasks          []*ConfigTask          `yaml:"tasks"`
	Authentication map[string]*ConfigAuth `yaml:"auth"`
	I18N           []ConfigI18N           `yaml:"i18n"`
}

type ConfigI18N struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type ConfigAuth struct {
	Password string `yaml:"password"`
}

type ConfigTask struct {
	ID        string                  `yaml:"id"`
	Name      string                  `yaml:"name"`
	ShortName string                  `yaml:"short_name"`
	Type      string                  `yaml:"type"`
	If        map[string]string       `yaml:"if"`
	Classes   map[string]*ConfigClass `yaml:"classes"`
}

type ConfigClass struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Examples    []string `yaml:"examples"`
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
	_taskDict := map[string]string{}
	for _, task := range ret.Tasks {
		taskName := task.ID
		_, ok := _taskDict[taskName]
		if ok {
			return nil, fmt.Errorf("task with %s is defined twice", taskName)
		}
		_taskDict[taskName] = ""
		if task.Type == "" {
			task.Type = "class"
		}
		if task.ShortName == "" {
			task.ShortName = task.Name
		}
		if task.Classes == nil {
			task.Classes = getClassesFromClassType(task.Type)
		}
		if task.Classes == nil {
			return nil, fmt.Errorf("task %s does not have any classes or a compatible type", taskName)
		}
	}
	if len(ret.Authentication) == 0 {
		return nil, fmt.Errorf("no users specified")
	}
	// Note: I18N configuration in YAML is deprecated.
	// Use annotation/locales/*.json files for translations instead.
	if len(ret.I18N) > 0 {
		log.Printf("Warning: i18n configuration in YAML is deprecated. Use annotation/locales/*.json files instead.")
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
				Name:        "OK",
				Description: "Not rotated",
			},
			"h_inv": {
				Name:        "Invert X",
				Description: "Invert in horizontal axis",
			},
			"v_inv": {
				Name:        "Invert Y",
				Description: "Invert in vertical axis",
			},
			"+90": {
				Name:        "+90deg",
				Description: "Rotate 90 degrees horary",
			},
			"-90": {
				Name:        "-90deg",
				Description: "Rotate 90 degrees antihorary",
			},
			"180": {
				Name:        "180deg",
				Description: "Rotate 180 degrees",
			},
		}
	default:
		return nil
	}
}
