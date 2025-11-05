package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"slices"
	"strings"
)

var rounds = []string{"[SPLIT/STEAL]", "[ROUND 1]", "[ROUND 2]", "[ROUND 3]", "[END]"}

func main() {
	filename := os.Args[1]

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open file: %s", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	count := 0
	for scanner.Scan() {
		count++
		line := scanner.Text()
		err := validateLine(line)
		if err != nil {
			fmt.Printf("error in line %v: %v\n", count, err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
}

func validateLine(line string) error {
	return errors.Join(validateSpeaker(line), validateMarkers(line), validateNumbers(line), validateWhitespaces(line), validateOverlaps(line))
}

func validateOverlaps(line string) error {
	if strings.Count(line, "[OVERLAP START]") != strings.Count(line, "[OVERLAP END]") {
		return fmt.Errorf("not all Overlaps have been properly closed: %v", line)
	}
	return nil
}

func validateSpeaker(line string) error {
	if slices.Contains(rounds, line) {
		return nil
	}
	valid := regexp.MustCompile(`^(HOST|SPEAKER1|SPEAKER2|SPEAKER3|SPEAKER4): `)
	if !valid.MatchString(line) {
		return fmt.Errorf("invalid speaker indication: %v", line)
	}
	return nil
}

func validateMarkers(line string) error {
	vals := append(rounds, "[LAUGHTER]", "[CUT]", "[OVERLAP START]", "[OVERLAP END]", "[?]", "[INAUDIBLE]")
	valid := regexp.MustCompile(`\[[^\[\]]+]`)
	matches := valid.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}
	if slices.Contains(vals, matches[0]) {
		return nil
	}
	return fmt.Errorf("invalid marker label: %v", matches[0])
}

func validateNumbers(line string) error {
	currs := []string{"$", "€", "£"}
	for _, curr := range currs {
		if strings.ContainsAny(line, curr) {
			return fmt.Errorf("currency symbol detected: %v", curr)
		}
	}

	money := regexp.MustCompile(`\b(\d+|\d+(?:[.,]\d+))\b`)
	if money.MatchString(line) {
		return fmt.Errorf("number not spelled out: %v", line)
	}
	return nil
}

func validateWhitespaces(line string) error {
	whit := regexp.MustCompile(`(\t| {2,})`)
	if whit.MatchString(line) {
		return fmt.Errorf("illegal whitespace detected: %v", line)
	}
	return nil
}
