package gemtext

import (
	"bufio"
	"bytes"
	"log"
	"net/url"
	"strings"
)

const (
	LINK_TYPE    = "=>"
	HEADING_TYPE = "#"
	LIST_TYPE    = "*"
	BLOCK_TYPE   = ">"
)

type Gemtext struct {
	Lines       []string
	LineTypeMap map[string]string
	Links       []string
}

// GetLinks returns only gemini links
func GetLinks(text, path string) ([]string, error) {
	_, links, _ := breakdown(text)
	var parsedLinks []string
	for _, v := range links {
		v = strings.Trim(v, "=>")
		v = strings.TrimSpace(v)

		linkParts := strings.Fields(v)
		if len(linkParts) > 0 {
			link := linkParts[0]

			l, err := url.Parse(link)
			if err != nil {
				log.Println(err)
				continue
			}

			if l.Host == "" {
				link = strings.Trim(path, "/") + "/" + strings.Trim(l.Path, "/")
				l, err = url.Parse(link)
				if err != nil {
					log.Println(err)
					continue
				}
			}

			if l.Scheme == "gemini" {
				parsedLinks = append(parsedLinks, link)
			}
		}
	}

	return parsedLinks, nil
}

func breakdown(text string) ([]string, []string, map[string]string) {
	scanner := bufio.NewScanner(bytes.NewBufferString(text))
	var lines, links []string
	lineMap := make(map[string]string)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lines = append(lines, line)

		firstChar := strings.Split(line, " ")[0]

		switch firstChar {
		case LINK_TYPE:
			links = append(links, line)
			lineMap[line] = LINK_TYPE
		case HEADING_TYPE:
			lineMap[line] = HEADING_TYPE
		case LIST_TYPE:
			lineMap[line] = LIST_TYPE
		case BLOCK_TYPE:
			lineMap[line] = BLOCK_TYPE
		}
	}
	return lines, links, lineMap
}
