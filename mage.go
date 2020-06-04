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

func newCi(name string) *ciConfig {
	return &ciConfig{
		Name: name,
		On:   []string{"push"},
		Jobs: map[string]ciJob{
			"build": {
				Name:   "build and test",
				RunsOn: "ubuntu-latest",
				Steps: []ciStep{
					{
						Run: "env",
					},
				},
			},
		},
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
	for _, p := range packages {
		d := strings.Replace(p, "github.com/chanzuckerberg/cztack/", "", 1)
		if len(d) > 0 {
			fmt.Println(d)

			ci := newCi(fmt.Sprintf("CI %s", d))

			out, err := yaml.Marshal(ci)

			if err != nil {
				return err
			}

			fileName := fmt.Sprintf("ci-%s.yml", strings.Replace(d, "/", "_", 1))
			filePath := filepath.Join(".github", "workflows", fileName)
			fmt.Println(filePath)
			ioutil.WriteFile(filePath, out, 0644)
		}
	}

	return nil
}
