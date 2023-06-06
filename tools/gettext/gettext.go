package main

import (
	"bike_race/core"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

func isIdentifierRune(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '.'
}

func extractKeys(content string) map[string]int {
	extractedKeys := make(map[string]int, 0)
	reg, err := regexp.Compile(`\{\{\s*call \$?\.T\s*"(\w+)"([\w.\s",:()]*)}}`)
	core.Expect(err, "error compiling regexp")
	matches := reg.FindAllStringSubmatch(string(content), -1)
	for _, match := range matches {
		key := match[1]
		args := strings.TrimSpace(match[2])
		argsCount := 0
		currentBlock := ""
		currentNesting := 0
		currentStringLiteral := false
		for _, char := range args {
			if isIdentifierRune(char) || currentStringLiteral {
				currentBlock += string(char)
			} else if char == ' ' && !currentStringLiteral {
				if currentBlock != "" {
					if currentNesting == 0 {
						argsCount++
					}
					currentBlock = ""
				}
			} else if char == '(' && !currentStringLiteral {
				currentBlock = ""
				currentNesting++
			} else if char == ')' && !currentStringLiteral {
				currentBlock = ""
				currentNesting--
			} else if char == '"' {
				currentStringLiteral = !currentStringLiteral
			}
		}
		if currentBlock != "" {
			argsCount++
		}
		extractedKeys[key] = argsCount
	}
	return extractedKeys
}

func extractAllKeys() map[string]int {
	extractedKeys := make(map[string]int, 0)

	filepath.Walk("templates", func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".html" {
			content, err := os.ReadFile(path)
			core.Expect(err, "error reading file")
			for key, value := range extractKeys(string(content)) {
				extractedKeys[key] = value
			}
		}
		return nil
	})

	return extractedKeys
}

func main() {
	write := flag.Bool("write", false, "write new locales")
	flag.Parse()
	baseLogger := slog.Default()
	langs := [...]string{"en-GB", "fr-FR"}

	extractedKeys := extractAllKeys()
	baseLogger.Info("extracted keys from templates", slog.Int("count", len(extractedKeys)))

	for _, lang := range langs {
		logger := baseLogger.With(slog.String("lang", lang))
		currentLocales := make(map[string]string, 0)
		newLocales := make(map[string]string, 0)

		locales, err := os.ReadFile(filepath.Join("locales", lang, "index.yml"))
		core.Expect(err, "error reading file")
		core.Expect(yaml.Unmarshal(locales, &currentLocales), "error unmarshalling yaml")

		correctLocales := 0
		for key, translation := range currentLocales {
			currentArgsCount := strings.Count(translation, "%")
			expectedArgsCount, ok := extractedKeys[key]
			if !ok {
				logger.Info(`found unused key`, slog.String("key", key))
			} else if currentArgsCount != extractedKeys[key] {
				logger.Info(`found translation with incorrect number of arguments`, slog.String("key", key), slog.Int("current", currentArgsCount), slog.Int("expected", expectedArgsCount))
				newLocales[key] = ""
			} else if translation == "" {
				newLocales[key] = ""
			} else {
				newLocales[key] = currentLocales[key]
				correctLocales++
			}
		}

		for key := range extractedKeys {
			if _, ok := currentLocales[key]; !ok {
				logger.Info(`found missing key`, slog.String("key", key))
				newLocales[key] = ""
			}
		}

		logger.Info("finished checking locales", slog.Int("count", len(newLocales)), slog.Int("correct", correctLocales), slog.String("completion", fmt.Sprintf("%d%%", correctLocales*100/len(extractedKeys))))

		if *write {
			content, err := yaml.Marshal(newLocales)
			core.Expect(err, "error marshalling yaml")
			core.Expect(os.WriteFile(filepath.Join("locales", lang, "index.yml"), content, 0644), "error writing file")
		}
	}
}
