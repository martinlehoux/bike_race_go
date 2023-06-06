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

func extractKeys() map[string]int {
	extractedKeys := make(map[string]int, 0)

	reg, err := regexp.Compile(`{{\s*call \$?\.T\s*"(.+)"\s*(?:([.\w]+)\s*)*}}`)
	core.Expect(err, "error compiling regexp")

	filepath.Walk("templates", func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".html" {
			content, err := os.ReadFile(path)
			core.Expect(err, "error reading file")
			matches := reg.FindAllStringSubmatch(string(content), -1)
			for _, match := range matches {
				key := match[1]
				parts := strings.Split(strings.TrimSpace(match[0][2:len(match[0])-2]), " ")
				extractedKeys[key] = len(parts) - 3
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

	extractedKeys := extractKeys()
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
