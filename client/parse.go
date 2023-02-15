package client

import (
	"bytes"
	"cf-tool/util"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/k0kubun/go-ansi"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func findSample(body []byte) (input [][]byte, output [][]byte, err error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	doc.Find(".sample-test .input").Each(func(_ int, s *goquery.Selection) {
		// For each item found, get the title
		inputCase := ""
		s.Find("pre").Contents().Each(func(_ int, s1 *goquery.Selection) {
			if !s1.Is("br") {
				inputCase += s1.Text() + "\n"
			}
		})
		for strings.HasSuffix(inputCase, "\n\n") {
			inputCase = inputCase[:len(inputCase)-1]
		}
		input = append(input, []byte(inputCase))
	})
	doc.Find(".sample-test .output").Each(func(_ int, s *goquery.Selection) {
		// For each item found, get the title
		outputCase := ""
		s.Find("pre").Contents().Each(func(_ int, s1 *goquery.Selection) {
			if !s1.Is("br") {
				outputCase += s1.Text() + "\n"
			}
		})
		for strings.HasSuffix(outputCase, "\n\n") {
			outputCase = outputCase[:len(outputCase)-1]
		}
		output = append(output, []byte(outputCase))

	})

	return
}

// ParseProblem parse problem to path. mu can be nil
func (c *Client) ParseProblem(URL, path string, mu *sync.Mutex) (samples int, standardIO bool, err error) {
	body, err := util.GetBody(c.client, URL)
	if err != nil {
		return
	}

	_, err = findHandle(body)
	if err != nil {
		return
	}

	input, output, err := findSample(body)
	if err != nil {
		return
	}

	standardIO = true
	if !bytes.Contains(body, []byte(`<div class="input-file"><div class="property-title">input</div>standard input</div><div class="output-file"><div class="property-title">output</div>standard output</div>`)) {
		standardIO = false
	}

	for i := 0; i < len(input); i++ {
		fileIn := filepath.Join(path, fmt.Sprintf("testI%v.txt", i+1))
		fileOut := filepath.Join(path, fmt.Sprintf("testO%v.txt", i+1))
		e := os.WriteFile(fileIn, input[i], 0644)
		if e != nil {
			if mu != nil {
				mu.Lock()
			}
			color.Red(e.Error())
			if mu != nil {
				mu.Unlock()
			}
		}
		e = os.WriteFile(fileOut, output[i], 0644)
		if e != nil {
			if mu != nil {
				mu.Lock()
			}
			color.Red(e.Error())
			if mu != nil {
				mu.Unlock()
			}
		}
	}
	return len(input), standardIO, nil
}

// Parse parse
func (c *Client) Parse(info Info) (problems []string, paths []string, err error) {
	color.Cyan("Parse " + info.Hint())

	problemID := info.ProblemID
	info.ProblemID = "%v"
	urlFormatter, err := info.ProblemURL(c.host)
	if err != nil {
		return
	}
	info.ProblemID = ""
	if problemID == "" {
		statics, err := c.Statis(info)
		if err != nil {
			return nil, nil, err
		}
		problems = make([]string, len(statics))
		for i, problem := range statics {
			problems[i] = problem.ID
		}
	} else {
		problems = []string{problemID}
	}
	contestPath := info.Path()
	ansi.Printf(color.CyanString("The problem(s) will be saved to %v\n"), color.GreenString(contestPath))

	wg := sync.WaitGroup{}
	wg.Add(len(problems))
	mu := sync.Mutex{}
	paths = make([]string, len(problems))
	for i, problemID := range problems {
		paths[i] = filepath.Join(contestPath, strings.ToLower(problemID))
		go func(problemID, path string) {
			defer wg.Done()
			mu.Lock()
			fmt.Printf("Parsing %v\n", problemID)
			mu.Unlock()

			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return
			}
			URL := fmt.Sprintf(urlFormatter, problemID)

			samples, standardIO, err := c.ParseProblem(URL, path, &mu)
			if err != nil {
				return
			}

			warns := ""
			if !standardIO {
				warns = color.YellowString("Non standard input output format.")
			}
			mu.Lock()
			if err != nil {
				color.Red("Failed %v. Error: %v", problemID, err.Error())
			} else {
				ansi.Printf("%v %v\n", color.GreenString("Parsed %v with %v samples.", problemID, samples), warns)
			}
			mu.Unlock()
		}(problemID, paths[i])
	}
	wg.Wait()
	return
}
