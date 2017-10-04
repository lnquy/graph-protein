package main

import (
	"os"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"path"
	"strings"
	"fmt"
)

var (
	pMap    = make(map[string]*Protein)
	setTmpl = `mutation {
	set {%v
	}
}`
)

type Protein struct {
	Eid          string
	Name         string
	FullName     string
	Interactions []string
}

func main() {
	parseYeastSNetData()
	parseYeastLNetData()
	setCmd := generateSetCommand()

	wd, _ := os.Getwd()
	if err := ioutil.WriteFile(path.Join(wd, "data", "cmds", "set.txt"), []byte(setCmd), 0666); err != nil {
		logrus.Panic(err)
	}
}

func parseYeastSNetData() {
	raw, err := readFile("YeastS.net")
	if err != nil {
		logrus.Panic(err)
	}
	lines := strings.Split(string(raw), "\n")
	isVertice, isEdge := false, false

	for _, line := range lines {
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "*edges") {
			isEdge = true
			isVertice = false
			continue
		}
		if strings.HasPrefix(line, "*vertices") {
			isVertice = true
			isEdge = false
			continue
		}

		if isVertice {
			parts := strings.Split(strings.TrimSpace(line), " ")
			pMap[parts[0]] = &Protein{
				Eid: parts[0],
				Name: strings.TrimSpace(parts[1]),
				Interactions: make([]string, 0),
			}
			continue
		}

		if isEdge {
			parts := strings.Split(line, " ")
			prot := pMap[parts[0]]
			prot.Interactions = append(prot.Interactions, strings.TrimSpace(parts[len(parts) - 1]))
			continue
		}
	}
}

func parseYeastLNetData() {
	raw, err := readFile("YeastL.net")
	if err != nil {
		logrus.Panic(err)
	}
	lines := strings.Split(string(raw), "\n")
	isVertice := false

	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "*edges") {
			break
		}
		if strings.HasPrefix(line, "*vertices") {
			isVertice = true
			continue
		}

		if isVertice {
			parts := strings.Split(strings.TrimSpace(line), " ")
			pMap[parts[0]].FullName = strings.TrimSpace(strings.Join(parts[1:], " "))
			continue
		}
	}
}

func generateSetCommand() string {
	set := ""
	for _, prot := range pMap {
		set += fmt.Sprintf("\n\t\t_:p%s <protein.eid> \"%s\" .\n", prot.Eid, prot.Eid)
		set += fmt.Sprintf("\t\t_:p%s <protein.name> %s .\n", prot.Eid, prot.Name)
		set += fmt.Sprintf("\t\t_:p%s <protein.full_name> %s .\n", prot.Eid, prot.FullName)
		for _, v := range prot.Interactions {
			set += fmt.Sprintf("\t\t_:p%s <protein.interaction> _:p%s .\n", prot.Eid, v)
		}
	}
	return fmt.Sprintf(setTmpl, set)
}

func readFile(name string) ([]byte, error) {
	wd, err := os.Getwd()
	if err != nil {
		logrus.Panic(err)
	}
	return ioutil.ReadFile(path.Join(wd, "data", "raw", name))
}
