// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const markdownExtension = ".md"
const introPage = "00.md"

func main() {
	fmt.Print(`<!--This sidebar was automatically generated by 'make site-gen-sidebar'-->
- [Installation](installation/)
    - [Docker](installation/docker/)
    - [gcloud](installation/gcloud/)
    - [Homebrew](installation/homebrew/)
    - [Source](installation/source/)
    - [Binaries](installation/binaries/)
- [Book](book/)
`)
	printBookOutline()
	fmt.Print(`- [Guides](guides/)
    - [For Consumers](guides/consumer/)
        - [Get](guides/consumer/get/)
        - [Apply](guides/consumer/apply/)
        - [Update](guides/consumer/update/)
        - [Running Functions](guides/consumer/function/)
            - Function Catalog
                - [Generators](guides/consumer/function/catalog/generators/)
                - [Sinks](guides/consumer/function/catalog/sinks/)
                - [Sources](guides/consumer/function/catalog/sources/)
                - [Transformers](guides/consumer/function/catalog/transformers/)
                - [Validators](guides/consumer/function/catalog/validators/)
            - [Exporting a Workflow](guides/consumer/function/export/)
                - [GitHub Actions](guides/consumer/function/export/github-actions/)
                - [GitLab CI](guides/consumer/function/export/gitlab-ci/)
                - [Jenkins](guides/consumer/function/export/jenkins/)
                - [Cloud Build](guides/consumer/function/export/cloud-build/)
                - [CircleCI](guides/consumer/function/export/circleci/)
                - [Tekton](guides/consumer/function/export/tekton/)
    - [For Producers](guides/producer/)
        - [Init](guides/producer/init/)
        - [Functions](guides/producer/functions/)
            - [Container Runtime](guides/producer/functions/container/)
            - [Exec Runtime](guides/producer/functions/exec/)
            - [Starlark Runtime](guides/producer/functions/starlark/)
            - [Go Function Libraries](guides/producer/functions/golang/)
            - [Typescript Function SDK](guides/producer/functions/ts/)
                - [Quickstart](guides/producer/functions/ts/quickstart/)
                - [Developer Guide](guides/producer/functions/ts/develop/)
        - [Packages](guides/producer/packages/)
        - [Bootstrapping](guides/producer/bootstrap/)
    - [Ecosystem](guides/ecosystem/)
        - [Kustomize](guides/ecosystem/kustomize/)
        - [Helm](guides/ecosystem/helm/)
        - [Open Application Model (OAM)](guides/ecosystem/oam/)
- [Reference](reference/)
    - [pkg](reference/pkg/)
        - [diff](reference/pkg/diff/)
        - [get](reference/pkg/get/)
        - [init](reference/pkg/init/)
        - [update](reference/pkg/update/)
    - [fn](reference/fn/)
        - [export](reference/fn/export/)
        - [run](reference/fn/run/)
        - [sink](reference/fn/sink/)
        - [source](reference/fn/source/)
    - [live](reference/live/)
        - [apply](reference/live/apply/)
        - [destroy](reference/live/destroy/)
        - [diff](reference/live/diff/)
        - [fetch-k8s-schema](reference/live/fetch-k8s-schema/)
        - [init](reference/live/init/)
        - [preview](reference/live/preview/)
        - [status](reference/live/status/)
- [Concepts](concepts/)
    - [API Conventions](concepts/api-conventions/)
    - [Architecture](concepts/architecture/)
    - [Functions](concepts/functions/)
    - [Packaging](concepts/packaging/)
- [FAQ](faq/)
- [Contact](contact/)`)
}

func printBookOutline() {
	sourcePath := "site/book"
	chapters := collectChapters(sourcePath)

	printChapters(chapters)
}

func collectChapters(source string) []chapter {
	chapters := make([]chapter, 0)
	chapterDirs, err := ioutil.ReadDir(source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	for _, dir := range chapterDirs {
		if dir.IsDir() {
			chapters = append(chapters, getChapter(dir.Name(), filepath.Join(source, dir.Name())))
		}
	}

	return chapters
}

func getChapter(chapterDirName string, chapterDirPath string) chapter {
	chapterBuilder := chapter{}

	// Split into chapter number and hyphenated name
	splitDirName := strings.SplitN(chapterDirName, "-", 2)
	chapterBuilder.Number = splitDirName[0]
	chapterBuilder.Name = strings.Title(strings.ReplaceAll(splitDirName[1], "-", " "))

	pageFiles, err := ioutil.ReadDir(chapterDirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	for _, pageFile := range pageFiles {
		if filepath.Ext(pageFile.Name()) == markdownExtension {
			chapterBuilder.Pages = append(chapterBuilder.Pages,
				getPage(pageFile.Name(), chapterBuilder.Name, chapterDirPath))
		}
	}

	return chapterBuilder
}

func getPage(pageFileName string, defaultName string, parentPath string) page {
	// Split into page number and hyphenated name.
	splitPageName := strings.SplitN(pageFileName, "-", 2)

	pageName := defaultName
	if pageFileName != introPage {
		// Strip page number and extension from file name.
		pageTitle := regexp.MustCompile(`^\d\d-?`).ReplaceAll([]byte(pageFileName), []byte(""))
		pageName = strings.Title(strings.ReplaceAll(strings.ReplaceAll(string(pageTitle), ".md", ""), "-", " "))
	}

	return page{
		Number: splitPageName[0],
		Name:   pageName,
		Path:   filepath.Join(parentPath, pageFileName),
	}
}

func printChapters(chapters []chapter) {
	// Sort chapters in ascending order by chapter number.
	sort.Slice(chapters, func(i, j int) bool { return chapters[i].Number < chapters[j].Number })

	for _, chapterEntry := range chapters {
		for pageNumber, pageEntry := range chapterEntry.Pages {
			// Make path relative to site directory.
			path := strings.Replace(pageEntry.Path, "site/", "", 1)

			// Print non-chapter intro pages as children of chapter intro page.
			if pageNumber == 0 {
				fmt.Printf("\t- [%s](%s)\n", pageEntry.Name, path)
			} else {
				fmt.Printf("\t\t- [%s](%s)\n", pageEntry.Name, path)
			}
		}
	}
}

type chapter struct {
	Name   string
	Pages  []page
	Number string
}

type page struct {
	Name   string
	Path   string
	Number string
}
