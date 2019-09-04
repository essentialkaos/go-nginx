// Package nginx provides methods for reading Nginx configuration files
package nginx

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2019 Essential Kaos                         //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ////////////////////////////////////////////////////////////////////////////////// //

type Properties map[string][]string

type Config struct {
	Root string
	File string

	Core   Properties
	Events Properties
	Stream Properties
	HTTP   *HTTP
}

type HTTP struct {
	Properties Properties
	Types      Properties
	Servers    []*Server
	Upstreams  map[string]*Upstream
}

type Server struct {
	Properties *ConditionalProperties
	Locations  []*Location
	Parent     *HTTP
}

type Location struct {
	Modifier   string
	URI        string
	Properties *ConditionalProperties
	Parent     *Server
}

type ConditionalProperties struct {
	Conditions []string
	Data       map[string][]ConditionalProperty
}

type ConditionalProperty struct {
	ConditionID int
	Value       string
}

type Upstream struct {
	Properties Properties
	Parent     *HTTP
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Read read and parse nginx config file
func Read(file string) (*Config, error) {
	filePath, _ := filepath.Abs(file)
	root := path.Dir(filePath)

	data, err := readFile(root, filePath)

	if err != nil {
		return nil, err
	}

	config, err := parseConfig(data)

	if err != nil {
		return nil, err
	}

	config.Root = root
	config.File = filePath

	return config, nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// readFile reads full config data (with all includes)
func readFile(root, file string) ([]string, error) {
	var filePath = file

	if !path.IsAbs(file) {
		filePath = root + "/" + file
	}

	fd, err := os.OpenFile(filePath, os.O_RDONLY, 0)

	if err != nil {
		return nil, err
	}

	defer fd.Close()

	reader := bufio.NewReader(fd)
	scanner := bufio.NewScanner(reader)

	var data, fullLine []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "include") {
			_, includesPath := parseProperty(line)
			includes, err := getInclude(root, includesPath)

			if err != nil {
				return nil, err
			}

			if len(includes) == 0 {
				continue
			}

			for _, include := range includes {
				includeData, err := readFile(root, include)

				if err != nil {
					return nil, err
				}

				data = append(data, includeData...)
			}

			continue
		}

		// "{ }" / "{ prop; }" / "prop; prop; ..."
		if !isFullLine(line) && !isBlockPart(line) {
			fullLine = append(fullLine, line)
		} else {
			if len(fullLine) != 0 {
				fullLine = append(fullLine, line)
				data = append(data, strings.Join(fullLine, " "))
				fullLine = nil
			} else {
				data = append(data, line)
			}
		}
	}

	return data, nil
}

// getInclude returns paths to file to include
func getInclude(root, file string) ([]string, error) {
	if strings.Contains(file, "*") {
		glob := file

		if !path.IsAbs(glob) {
			glob = root + "/" + file
		}

		matches, err := filepath.Glob(glob)

		if err != nil {
			return nil, fmt.Errorf("Can't read glob %s: %v", file, err)
		}

		return matches, nil
	}

	return []string{file}, nil
}

// parseConfig parses config data
func parseConfig(data []string) (*Config, error) {
	var err error
	var cursor int

	config := &Config{Core: make(Properties)}

	for cursor < len(data) {
		line := data[cursor]

		if isBlockStart(line) {
			blockName, _ := getBlockName(line)

			switch blockName {
			case "events":
				var events Properties
				cursor, events, err = parseSimpleBlock(data, cursor+1)

				if err != nil {
					return nil, err
				}

				config.Events = events

				continue

			case "stream":
				var stream Properties
				cursor, stream, err = parseSimpleBlock(data, cursor+1)

				if err != nil {
					return nil, err
				}

				config.Stream = stream

				continue

			case "http":
				var http *HTTP
				cursor, http, err = parseHTTPBlock(data, cursor+1)

				if err != nil {
					return nil, err
				}

				config.HTTP = http

				continue

			default:
				return nil, fmt.Errorf("Unsupported block %s inside config body", blockName)
			}
		}

		propName, propValue := parseProperty(line)
		config.Core[propName] = append(config.Core[propName], propValue)

		cursor++
	}

	return config, nil
}

// parseSimpleBlock parses any simple block
func parseSimpleBlock(data []string, cursor int) (int, Properties, error) {
	result := make(Properties)
	dataLen := len(data)

	for {
		if cursor >= dataLen {
			break
		}

		line := data[cursor]

		if isBlockEnd(line) {
			return cursor + 1, result, nil
		}

		propName, propValue := parseProperty(line)
		result[propName] = append(result[propName], propValue)

		cursor++
	}

	return -1, nil, fmt.Errorf("Can't find block end")
}

// parseHTTPBlock parses http block
func parseHTTPBlock(data []string, cursor int) (int, *HTTP, error) {
	var err error

	http := &HTTP{Properties: make(Properties), Upstreams: make(map[string]*Upstream)}
	dataLen := len(data)

	for {
		if cursor >= dataLen {
			break
		}

		line := data[cursor]

		if isBlockEnd(line) {
			return cursor + 1, http, nil
		}

		if isBlockStart(line) {
			blockName, blockArgs := getBlockName(line)

			switch blockName {
			case "types":
				var types Properties
				cursor, types, err = parseSimpleBlock(data, cursor+1)

				if err != nil {
					return -1, nil, err
				}

				http.Types = types

				continue

			case "server":
				var server *Server
				cursor, server, err = parseServerBlock(data, cursor+1)

				if err != nil {
					return -1, nil, err
				}

				server.Parent = http
				http.Servers = append(http.Servers, server)

				continue

			case "upstream":
				upstreamName := getSafe(blockArgs, 0)

				if upstreamName == "" {
					return -1, nil, fmt.Errorf("Unsupported upstream block doesn't have the name")
				}

				var props Properties
				cursor, props, err = parseSimpleBlock(data, cursor+1)

				if err != nil {
					return -1, nil, err
				}

				http.Upstreams[upstreamName] = &Upstream{props, http}

				continue

			default:
				return -1, nil, fmt.Errorf("Unsupported block %s inside http block", blockName)
			}
		}

		propName, propValue := parseProperty(line)
		http.Properties[propName] = append(http.Properties[propName], propValue)

		cursor++
	}

	return -1, nil, fmt.Errorf("Can't find block end")
}

// parseServerBlock parses server block
func parseServerBlock(data []string, cursor int) (int, *Server, error) {
	var err error

	server := &Server{Properties: &ConditionalProperties{
		Data: make(map[string][]ConditionalProperty),
	}}

	dataLen := len(data)

	for {
		if cursor >= dataLen {
			break
		}

		line := data[cursor]

		if isBlockEnd(line) {
			return cursor + 1, server, nil
		}

		if isBlockStart(line) {
			blockName, blockArgs := getBlockName(line)

			switch blockName {
			case "if":
				cursor, err = parseIfBlock(data, cursor+1, server.Properties)

				if err != nil {
					return -1, nil, err
				}

				continue

			case "location":
				var location *Location
				cursor, location, err = parseLocationBlock(data, cursor+1)

				if err != nil {
					return -1, nil, err
				}

				location.Parent = server
				location.URI, location.Modifier = parseLocationArgs(blockArgs)
				server.Locations = append(server.Locations, location)

				continue

			default:
				return -1, nil, fmt.Errorf("Unsupported block %s inside server block", blockName)
			}
		}

		propName, propValue := parseProperty(line)
		server.Properties.Data[propName] = append(server.Properties.Data[propName], ConditionalProperty{-1, propValue})

		cursor++
	}

	return -1, nil, fmt.Errorf("Can't find block end")
}

// parseLocationBlock parses location block
func parseLocationBlock(data []string, cursor int) (int, *Location, error) {
	var err error

	location := &Location{Properties: &ConditionalProperties{
		Data: make(map[string][]ConditionalProperty),
	}}

	dataLen := len(data)

	for {
		if cursor >= dataLen {
			break
		}

		line := data[cursor]

		if isBlockEnd(line) {
			return cursor + 1, location, nil
		}

		if isBlockStart(line) {
			blockName, _ := getBlockName(line)

			switch blockName {
			case "if":
				cursor, err = parseIfBlock(data, cursor+1, location.Properties)

				if err != nil {
					return -1, nil, err
				}

				continue

			default:
				return -1, nil, fmt.Errorf("Unsupported block %s inside location block", blockName)
			}
		}

		propName, propValue := parseProperty(line)
		location.Properties.Data[propName] = append(location.Properties.Data[propName], ConditionalProperty{-1, propValue})

		cursor++
	}

	return -1, nil, fmt.Errorf("Can't find block end")
}

// parseIfBlock parses condition block
func parseIfBlock(data []string, cursor int, props *ConditionalProperties) (int, error) {
	dataLen := len(data)

	condition := getCondition(data[cursor-1])
	conditionID := len(props.Conditions)

	props.Conditions = append(props.Conditions, condition)

	for {
		if cursor >= dataLen {
			break
		}

		line := data[cursor]

		if isBlockEnd(line) {
			return cursor + 1, nil
		}

		propName, propValue := parseProperty(line)
		props.Data[propName] = append(props.Data[propName], ConditionalProperty{conditionID, propValue})

		cursor++
	}

	return -1, fmt.Errorf("Can't find block end")
}

// parseProperty parses property and returns name and value
func parseProperty(data string) (string, string) {
	data = cleanData(data)

	firstSpaceIndex := strings.Index(data, " ")

	if firstSpaceIndex == -1 || len(data) < firstSpaceIndex+1 {
		return data, ""
	}

	return data[:firstSpaceIndex], data[firstSpaceIndex+1:]
}

// parseLocationArgs parses location args
func parseLocationArgs(data []string) (string, string) {
	switch len(data) {
	case 2:
		return data[1], data[0]
	default:
		return data[0], ""
	}
}

// getCondition parses condition
func getCondition(data string) string {
	cStart := strings.Index(data, "(")
	cEnd := strings.LastIndex(data, ")")

	return data[cStart+1 : cEnd-1]
}

// getBlockName extracts block name
func getBlockName(data string) (string, []string) {
	data = cleanData(data)
	data = strings.TrimRight(data, " {")

	if strings.Count(data, " ") == 0 {
		return data, nil
	}

	dataSlice := strings.Fields(data)

	return dataSlice[0], dataSlice[1:]
}

// isFullLine returns true if given line terminated by symbol ;
func isFullLine(data string) bool {
	return strings.Contains(data, ";")
}

// isBlockStart returns true if given data contains block start symbol
func isBlockStart(data string) bool {
	var hasBracket, escaped bool

	for _, r := range data {
		switch r {
		case '"', '\'':
			if escaped {
				escaped = false
			} else {
				escaped = true
			}

		case '{':
			if !escaped {
				hasBracket = true
			}

		case '#':
			if !escaped {
				break
			}
		}
	}

	return hasBracket && !escaped
}

// isBlockEnd returns true if given data contains block end symbol
func isBlockEnd(data string) bool {
	var hasBracket, escaped bool

	for _, r := range data {
		switch r {
		case '"', '\'':
			if escaped {
				escaped = false
			} else {
				escaped = true
			}

		case '}':
			if !escaped {
				hasBracket = true
			}

		case '#':
			if !escaped {
				break
			}
		}
	}

	return hasBracket && !escaped
}

// isBlockPart returns true if given data is a block part
func isBlockPart(data string) bool {
	return isBlockStart(data) || isBlockEnd(data)
}

func cleanData(data string) string {
	// Remove spaces from both sides
	data = strings.TrimSpace(data)

	// Spaces deduplication
	for strings.Contains(data, "  ") {
		data = strings.Replace(data, "  ", " ", -1)
	}

	// Remove inline comments
	var end int
	var escaped bool

	for i, r := range data {
		switch r {
		case '"', '\'':
			if escaped {
				escaped = false
			} else {
				escaped = true
			}
		case '#':
			if end == 0 && !escaped {
				end = i - 1
			}
		case ';':
			if !escaped {
				end = i
				break
			}
		}
	}

	if end != 0 {
		return data[:end]
	}

	return data
}

// getSafe reads value from slice
func getSafe(data []string, index int) string {
	if index < len(data) {
		return data[index]
	}

	return ""
}
