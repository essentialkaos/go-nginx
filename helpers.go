package nginx

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2020 ESSENTIAL KAOS                          //
//      Essential Kaos Open Source License <http://essentialkaos.com/ekol?en>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ////////////////////////////////////////////////////////////////////////////////// //

var errEmptyProp = fmt.Errorf("Value is empty")

// ////////////////////////////////////////////////////////////////////////////////// //

// String returns config as a string
func (c *Config) String() string {
	return "<" + c.File + ">"
}

// ////////////////////////////////////////////////////////////////////////////////// //

// ServersNum returns number of server directives
func (h *HTTP) ServersNum() int {
	if h == nil || h.Servers == nil {
		return 0
	}

	return len(h.Servers)
}

// ServersList returns slice with all servers names
func (h *HTTP) ServersList() []string {
	var result []string

	if h.ServersNum() == 0 {
		return result
	}

	for _, server := range h.Servers {
		names := server.GetNames()
		protocols := server.Properties.Get("listen")

		for _, name := range names {
			if strings.Contains(protocols, "http2") ||
				strings.Contains(protocols, "ssl") ||
				strings.Contains(protocols, "443") {
				result = append(result, name+":https")
			} else {
				result = append(result, name+":http")
			}
		}
	}

	return result
}

// FindServer tries to find server info
func (h *HTTP) FindServer(name, protocol string) *Server {
	if h.ServersNum() == 0 {
		return nil
	}

	for _, server := range h.Servers {
		names := server.GetNames()

		for _, serverName := range names {
			if serverName != name {
				continue
			}

			if isProtocolSupported(server.GetProtocols(), protocol) {
				return server
			}
		}
	}

	return nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// GetNames returns slice with servers names
func (s *Server) GetNames() []string {
	return strings.Fields(s.Properties.Get("server_name"))
}

// GetProtocols returns slice with supported protocols
func (s *Server) GetProtocols() []string {
	return strings.Fields(s.Properties.Get("listen"))
}

////////////////////////////////////////////////////////////////////////////////// //

// Get returns property with given name (if present)
func (p Properties) Get(name string) string {
	if p == nil {
		return ""
	}

	for k, v := range p {
		if k == name {
			return strings.Join(v, " ")
		}
	}

	return ""
}

// GetBool returns property with given name as bool
func (p Properties) GetBool(name string) (bool, error) {
	v := p.Get(name)

	if v == "" {
		return false, errEmptyProp
	}

	return parseBool(v)
}

// GetInt returns property with given name as int64
func (p Properties) GetInt(name string) (int64, error) {
	v := p.Get(name)

	if v == "" {
		return 0, errEmptyProp
	}

	return parseInt(v)
}

// GetBuf returns property with given name as buffer
func (p Properties) GetBuf(name string) (int64, int64, error) {
	v := p.Get(name)

	if v == "" {
		return 0, 0, errEmptyProp
	}

	return parseBuffers(v)
}

// GetSize parses and returns property with given name as size in bytes
func (p Properties) GetSize(name string) (int64, error) {
	v := p.Get(name)

	if v == "" {
		return 0, errEmptyProp
	}

	return parseSize(v)
}

// GetTime parses and returns property with given name as time duration
func (p Properties) GetTime(name string) (time.Duration, error) {
	v := p.Get(name)

	if v == "" {
		return 0, errEmptyProp
	}

	return parseTime(v)
}

////////////////////////////////////////////////////////////////////////////////// //

// Get returns property with given name (if present)
func (p *ConditionalProperties) Get(name string) string {
	if p == nil || p.Data == nil {
		return ""
	}

	var result []string

	for k, v := range p.Data {
		if k == name {
			for _, p := range v {
				result = append(result, p.Value)
			}

			return strings.Join(result, " ")
		}
	}

	return ""
}

// GetBool returns property with given name as bool
func (p *ConditionalProperties) GetBool(name string) (bool, error) {
	v := p.Get(name)

	if v == "" {
		return false, errEmptyProp
	}

	return parseBool(v)
}

// GetInt returns property with given name as int64
func (p *ConditionalProperties) GetInt(name string) (int64, error) {
	v := p.Get(name)

	if v == "" {
		return 0, errEmptyProp
	}

	return parseInt(v)
}

// GetBuf returns property with given name as buffer
func (p *ConditionalProperties) GetBuf(name string) (int64, int64, error) {
	v := p.Get(name)

	if v == "" {
		return 0, 0, errEmptyProp
	}

	return parseBuffers(v)
}

// GetSize parses and returns property with given name as size in bytes
func (p *ConditionalProperties) GetSize(name string) (int64, error) {
	v := p.Get(name)

	if v == "" {
		return 0, errEmptyProp
	}

	return parseSize(v)
}

// GetTime parses and returns property with given name as time duration
func (p *ConditionalProperties) GetTime(name string) (time.Duration, error) {
	v := p.Get(name)

	if v == "" {
		return 0, errEmptyProp
	}

	return parseTime(v)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// isProtocolSupported checks if protocol is supported
func isProtocolSupported(protocolList []string, protocol string) bool {
	for _, p := range protocolList {
		switch {
		case p == "http2" && (protocol == "http2" || protocol == "https"),
			p == "spdy" && (protocol == "spdy" || protocol == "https"),
			p == "ssl" && (protocol == "ssl" || protocol == "https"),
			p == "443" && (protocol == "443" || protocol == "https"),
			p == "80" && (protocol == "80" || protocol == "http"),
			p == protocol:
			return true
		}
	}

	return false
}

func parseBool(s string) (bool, error) {
	switch s {
	case "on":
		return true, nil
	case "off":
		return false, nil
	}

	return false, fmt.Errorf("Unsupported boolean value %s", s)
}

func parseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func parseBuffers(s string) (int64, int64, error) {
	if strings.Count(s, " ") != 1 {
		return 0, 0, fmt.Errorf("Wrong buffer format value")
	}

	buf := strings.Fields(s)
	num, err := parseInt(buf[0])

	if err != nil {
		return 0, 0, err
	}

	size, err := parseSize(buf[1])

	if err != nil {
		return 0, 0, err
	}

	return num, size, nil
}

func parseSize(s string) (int64, error) {
	mod := strings.Trim(s, "0123456789")
	val := strings.Trim(s, "kKmMgG")

	valInt, _ := strconv.ParseInt(val, 10, 64)

	switch mod {
	case "k", "K":
		return valInt * 1024, nil
	case "m", "M":
		return valInt * 1024 * 1024, nil
	case "g", "G":
		return valInt * 1024 * 1024 * 1024, nil
	case "":
		return valInt, nil
	}

	return 0, fmt.Errorf("Unsupported measurement unit %s", mod)
}

func parseTime(t string) (time.Duration, error) {
	var result time.Duration

	for _, part := range strings.Fields(t) {
		dur, err := parseTimePeriod(part)

		if err != nil {
			return 0, err
		}

		result += dur
	}

	return result, nil
}

func parseTimePeriod(p string) (time.Duration, error) {
	mod := strings.Trim(p, "0123456789")
	val := strings.Trim(p, "mshdwMy")

	valInt, _ := strconv.ParseInt(val, 10, 64)

	switch mod {
	case "ms":
		return time.Duration(valInt) * time.Millisecond, nil
	case "s", "":
		return time.Duration(valInt) * time.Second, nil
	case "m":
		return time.Duration(valInt) * time.Minute, nil
	case "h":
		return time.Duration(valInt) * time.Hour, nil
	case "d":
		return time.Duration(valInt) * time.Hour * 24, nil
	case "w":
		return time.Duration(valInt) * time.Hour * 24 * 7, nil
	case "M":
		return time.Duration(valInt) * time.Hour * 24 * 30, nil
	case "y":
		return time.Duration(valInt) * time.Hour * 24 * 365, nil
	}

	return 0, fmt.Errorf("Unsupported time unit %s", mod)
}
