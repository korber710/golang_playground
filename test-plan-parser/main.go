package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type TestExecution struct {
	Name    string
	Passive bool
	Tags    string
}

type TestExecutions struct {
	Tests []TestExecution
}

func main() {
	testSuitesFile, err := ioutil.ReadFile("test-suites.yml")
	if err != nil {
		panic(err)
	}

	type rawTestExecution map[string]map[string]interface{}
	testStuites := []rawTestExecution{}
	err = yaml.Unmarshal(testSuitesFile, &testStuites)
	if err != nil {
		panic(err)
	}

	tests := TestExecutions{}
	for _, v := range testStuites {
		// fmt.Printf("%s -> %d\n", k, v)
		// newTest := TestExecution{}
		fmt.Printf("v: %+v\n", v)
		for i, x := range v {
			fmt.Printf("i: %v, x: %+v\n", i, x)
			// fmt.Printf("%v\n", x["passive"].(bool))
			tests.Tests = append(tests.Tests,
				TestExecution{
					Name:    i,
					Passive: x["passive"].(bool),
					Tags:    x["tags"].(string)})
		}
	}

	fmt.Printf("tests: %+v\n", tests)
}
