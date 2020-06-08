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
	Name string            `yaml:"name,omitempty"`
	Run  string            `yaml:"run,omitempty"`
	Uses string            `yaml:"uses,omitempty"`
	With map[string]string `yaml:"with,omitempty"`
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

// Ci generates a ci configurations for this repo.
func Ci() error {
	out, err := exec.Command("go", "list", "./...").Output()
	if err != nil {
		return err
	}

	packages := strings.Split(string(out), "\n")

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
						Uses: "hashicorp/setup-terraform@v1",
						With: map[string]string{
							"terraform_version": "0.12.24",
							"terraform_wrapper": "false",
						},
					},
					{
						Uses: "actions/setup-go@v2",
						With: map[string]string{
							"go-version": "1.14.3",
						},
					},
					{Run: "aws configure set aws_access_key_id ${{ secrets.CI1_AWS_ACCESS_KEY_ID }} --profile cztack-ci-1"},
					{Run: "aws configure set aws_secret_access_key ${{ secrets.CI1_AWS_SECRET_ACCESS_KEY }} --profile cztack-ci-1"},
					{Run: "aws --profile cztack-ci-1 sts get-caller-identity"},
					{Run: "aws configure set aws_access_key_id ${{ secrets.CI2_AWS_ACCESS_KEY_ID }} --profile cztack-ci-2"},
					{Run: "aws configure set aws_secret_access_key ${{ secrets.CI2_AWS_SECRET_ACCESS_KEY }} --profile cztack-ci-2"},
					{Run: "aws --profile cztack-ci-2 sts get-caller-identity"},
					{Run: fmt.Sprintf("make test-ci TEST=./%s", d)},
				},
			}
			ci.Jobs[name] = j
		}
	}

	yml, err := yaml.Marshal(ci)
	if err != nil {
		return err
	}
	ioutil.WriteFile(filepath.Join(".github", "workflows", "ci.yml"), yml, 0644)
	return nil
}
