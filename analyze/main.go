package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var verbs = []string{"can", "could", "able to", "must", "have to", "got to", "may", "might", "will", "'ll", "would", "going to", "gonna", "shall", "should", "ought to"}

type finalRound struct {
	name     string
	speaker1 speaker
	speaker2 speaker
}

func (fr finalRound) count(verb string) (count int) {
	for _, it := range fr.speaker1 {
		if it.verb == verb {
			count += 1
		}
	}
	for _, it := range fr.speaker2 {
		if it.verb == verb {
			count += 1
		}
	}
	return count
}

type speaker []modal

func (s speaker) count(verb string) (count int) {
	for _, it := range s {
		if it.verb == verb {
			count += 1
		}
	}
	return count
}

type modal struct {
	count int
	verb  string
	line  string
}

func main() {
	path := os.Args[1]
	fmt.Printf("given path: %v\n", path)

	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Fatalf("Failed to find given path: %s", err)
	}

	files := make([]string, 0)
	if fileInfo.IsDir() {
		dirEnts, err := os.ReadDir(path)
		if err != nil {
			log.Fatalf("failed reading files from dir due to %v", err)
		}
		fmt.Printf("found %v files in directory\n", len(dirEnts))
		for _, curr := range dirEnts {
			if !strings.HasSuffix(curr.Name(), ".txt") {
				fmt.Printf("ignoring file %v because not .txt\n", curr)
				continue
			}
			files = append(files, filepath.Join(path, curr.Name()))
		}
	}

	var rnds []finalRound
	var fileCtx string
	var roll rollingModal
	for _, curr := range files {
		file, err := os.Open(curr)
		if err != nil {
			log.Fatalf("Failed to open file: %s", err)
		}
		fileCtx = strings.TrimSuffix(filepath.Base(curr), filepath.Ext(curr))
		content, err := os.ReadFile(curr)
		if err != nil {
			log.Fatal(err)
		}

		contentStr := strings.ReplaceAll(string(content), "\r\n", "\n")
		contentStr = strings.ReplaceAll(contentStr, "\r", "\n")

		roll = rollingModal{"", speaker{}, speaker{}, speaker{}}
		lines := strings.Split(contentStr, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "[END]" {
				break
			}
			roll.update(line)
			for _, verb := range verbs {
				lowLine := strings.ToLower(line)
				if strings.Contains(lowLine, verb) {
					roll.verb(modal{strings.Count(lowLine, verb), verb, line})
				}
			}
		}

		rnds = append(rnds, finalRound{fileCtx, roll.speaker1, roll.speaker2})
		file.Close()
	}

	//for _, vb := range verbs {
	//	count := 0
	//	for _, rnd := range rnds {
	//		count += rnd.count(vb)
	//	}
	//	fmt.Printf("%v: %v\n", vb, count)
	//}

	//for _, rnd := range rnds {
	//	fmt.Println(rnd.name)
	//	fmt.Printf("Speaker1: \n")
	//	for _, vrb := range verbs {
	//		fmt.Printf("%v: %v\n", vrb, rnd.speaker1.count(vrb))
	//	}
	//	fmt.Printf("Speaker2: \n")
	//	for _, vrb := range verbs {
	//		fmt.Printf("%v: %v\n", vrb, rnd.speaker2.count(vrb))
	//	}
	//	fmt.Println()
	//}

	for _, rnd := range rnds {
		for _, vrb := range rnd.speaker1 {
			fmt.Printf("%v;%v;%v;%v;%q\n", rnd.name, "Speaker1", vrb.verb, vrb.count, vrb.line)
		}
		for _, vrb := range rnd.speaker2 {
			fmt.Printf("%v;%v;%v;%v;%q\n", rnd.name, "Speaker2", vrb.verb, vrb.count, vrb.line)
		}
	}
}

type rollingModal struct {
	speakerCtx string
	speaker1   speaker
	speaker2   speaker
	host       speaker
}

var speakerReg = regexp.MustCompile(`^(SPEAKER\s*1|SPEAKER\s*2|HOST)\s*:`)

func (rm *rollingModal) verb(modal modal) {
	switch rm.speakerCtx {
	case "SPEAKER1":
		rm.speaker1 = append(rm.speaker1, modal)
		return
	case "SPEAKER2":
		rm.speaker2 = append(rm.speaker2, modal)
		return
	case "HOST":
		rm.host = append(rm.host, modal)
		return
	}
}

func (rm *rollingModal) update(line string) {
	matches := speakerReg.FindStringSubmatch(line)
	if matches != nil {
		rm.speakerCtx = strings.ReplaceAll(strings.ToUpper(matches[1]), " ", "")
	}
}
