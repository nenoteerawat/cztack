// +build mage

package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v1"
)

type ciJob struct {
	Name           string   `yaml:"name,omitempty"`
	RunsOn         string   `yaml:"runs-on,omitempty"`
	TimeoutMinutes int      `yaml:"timeout-minutes,omitempty"`
	Steps          []ciStep `yaml:"steps,omitempty"`
}

type ciStep struct {
	Name string `yaml:"name,omitempty"`
	Run  string `yaml:"run,omitempty"`
	Uses string `yaml:"uses,omitempty"`
	// TODO
	// With ciStepWith
}

// type ciStepWith struct {
// 	Path string `yaml:"path"`
// 	Key  string `yaml:"key"`
// }

type ciConfig struct {
	Name string   `yaml:"name"`
	On   []string `yaml:"on"`
	Jobs map[string]ciJob
}

func newCi(name string) ciConfig {
	return ciConfig{
		Name: name,
		On:   []string{"push"},
		Jobs: map[string]ciJob{},
	}
}

func Ci() error {
	//  for each module directory
	//    write GHA ci file

	out, err := exec.Command("go", "list", "./...").Output()
	if err != nil {
		return err
	}

	fmt.Println(string(out))
	packages := strings.Split(string(out), "\n")

	fmt.Println(packages)
	fmt.Println(len(packages))

	ci := newCi("CI")

	for _, p := range packages {
		d := strings.Replace(p, "github.com/chanzuckerberg/cztack/", "", 1)
		if len(d) > 0 {
			fmt.Println(d)
			j := ciJob{
				Name:           fmt.Sprintf("%s", d),
				RunsOn:         "ubuntu-latest",
				TimeoutMinutes: 45,
				Steps: []ciStep{
					{
						Run: "env",
					},
				},
			}
			ci.Jobs[d] = j
		}
	}

	yml, err := yaml.Marshal(ci)
	if err != nil {
		return err
	}
	ioutil.WriteFile(filepath.Join(".github", "workflows", "ci2.yml"), yml, 0644)
	return nil
}
