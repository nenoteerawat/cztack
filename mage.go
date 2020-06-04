// +build mage

package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ciJob struct {
	Name           string   `yaml:"name,omitempty"`
	RunsOn         string   `yaml:"runs-on,omitempty"`
	TimeoutMinutes int      `yaml:"timeout-minutes,omitempty"`
	Steps          []ciStep `yaml:"steps,omitempty"`
}

type ciStep struct {
	Name string            `yaml:"name,omitempty"`
	Run  string            `yaml:"run,omitempty"`
	Uses string            `yaml:"uses,omitempty"`
	With map[string]string `yaml:"with,omitempty"`
	// TODO
	// With ciStepWith
}

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
			name := strings.Replace(d, "/", "_", -1)
			j := ciJob{
				Name:           name,
				RunsOn:         "ubuntu-latest",
				TimeoutMinutes: 45,
				Steps: []ciStep{
					{
						Run: "env",
					},
					{
						Uses: "actions/checkout@v2",
					},
					{
						Name: "cache homebrew",
						Uses: "actions/cache@v2",
						With: map[string]string{
							"path": "/home/linuxbrew/.linuxbrew",
							"key":  "${{ runner.os }}-brew-${{ hashFiles('Brewfile.lock.json') }}",
						},
					},
					{
						Run: "brew bundle install",
					},
					{Run: `
                        cat << EOF > ~/.aws/credentials
                        [cztack-ci-1]
                        aws_access_key_id=${{ secrets.CI1_AWS_ACCESS_KEY_ID }}
                        aws_secret_access_key=${{ secrets.CI1_AWS_SECRET_ACCESS_KEY }}

                        [cztack-ci-2]
                        aws_access_key_id=${{ secrets.CI2_AWS_ACCESS_KEY_ID }}
                        aws_secret_access_key=${{ secrets.CI2_AWS_SECRET_ACCESS_KEY }}
                        EOF

                    `},
					// aws --profile cztack-ci-1 sts get-caller-identity;
					// aws --profile cztack-ci-2 sts get-caller-identity;
					// {Run: "tfenv install 0.12.24"},
					// {Run: "tfenv use 0.12.24"},
					// {Run: fmt.Sprintf("make test-ci TEST=./%s", p)},
				},
			}
			ci.Jobs[name] = j
		}
	}

	yml, err := yaml.Marshal(ci)
	if err != nil {
		return err
	}
	ioutil.WriteFile(filepath.Join(".github", "workflows", "ci2.yml"), yml, 0644)
	return nil
}
