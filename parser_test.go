package nginx

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2020 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"testing"
	"time"

	. "pkg.re/check.v1"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type NginxSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&NginxSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *NginxSuite) TestParsing(c *C) {
	config, err := Read("testdata/webkaos.conf", "")

	c.Assert(err, IsNil)
	c.Assert(config, NotNil)

	server := config.HTTP.FindServer("service.domain.com", "https")

	c.Assert(server, NotNil)
	c.Assert(server.Properties.Get("ssl_certificate"), Equals, "/etc/webkaos/ssl/my-chain.crt")
}

func (s *NginxSuite) TestPartParsing(c *C) {
	http, err := ReadPart("testdata/conf.d/service.conf", "")

	c.Assert(err, IsNil)
	c.Assert(http, NotNil)

	server := http.FindServer("service.domain.com", "https")

	c.Assert(server, NotNil)
	c.Assert(server.Properties.Get("ssl_certificate"), Equals, "/etc/webkaos/ssl/my-chain.crt")
}

func (s *NginxSuite) TestErrors(c *C) {
	config, err := Read("testdata/unknown.conf", "")

	c.Assert(err, NotNil)
	c.Assert(config, IsNil)

	http, err := ReadPart("testdata/unknown.conf", "")

	c.Assert(err, NotNil)
	c.Assert(http, IsNil)

	config, err = Read("testdata/webkaos-broken.conf", "")

	c.Assert(err, NotNil)
	c.Assert(config, IsNil)

	config, err = Read("testdata/webkaos-broken2.conf", "")

	c.Assert(err, NotNil)
	c.Assert(config, IsNil)

	data := []string{"events {", "worker_connections  8192;"}

	_, err = parseConfig(data)
	c.Assert(err, NotNil)

	data = []string{"stream {", "include stream.conf.d/*.conf;"}

	_, err = parseConfig(data)
	c.Assert(err, NotNil)

	data = []string{"http {", "server_tokens  off;"}

	_, err = parseConfig(data)
	c.Assert(err, NotNil)

	data = []string{"unknown {"}

	_, err = parseConfig(data)
	c.Assert(err, NotNil)
}

func (s *NginxSuite) TestAux(c *C) {
	c.Assert(getSafe([]string{"1", "2"}, 0), Equals, "1")
	c.Assert(getSafe([]string{"1", "2"}, 99), Equals, "")

	c.Assert(cleanData("  location = '{' { # TEST"), Equals, "location = '{' {")
	c.Assert(cleanData("resolver_timeout           10s;"), Equals, "resolver_timeout 10s")
}

func (s *NginxSuite) TestIfBlockParser(c *C) {
	data := []string{"if ($k == 1) {", "return 100;"}
	props := &ConditionalProperties{Data: make(map[string][]ConditionalProperty)}

	_, err := parseIfBlock(data, 1, props)
	c.Assert(err, NotNil)
}

func (s *NginxSuite) TestLocationBlockParser(c *C) {
	data := []string{"if ($k == 1) {", "return 100;"}

	_, _, err := parseLocationBlock(data, 0)
	c.Assert(err, NotNil)

	data = []string{"root /home;", "proxy_pass http://123.0.0.111:80/;"}

	_, _, err = parseLocationBlock(data, 0)
	c.Assert(err, NotNil)

	data = []string{"http {", "proxy_pass http://123.0.0.111:80/;"}

	_, _, err = parseLocationBlock(data, 0)
	c.Assert(err, NotNil)
}

func (s *NginxSuite) TestServerBlockParser(c *C) {
	data := []string{"location / {", "if ($k == 1) {", "return 100;"}

	_, _, err := parseServerBlock(data, 0)
	c.Assert(err, NotNil)

	data = []string{"if ($k == 1) {", "return 100;"}

	_, _, err = parseServerBlock(data, 0)
	c.Assert(err, NotNil)

	data = []string{"http {", "proxy_pass http://123.0.0.111:80/;"}

	_, _, err = parseServerBlock(data, 0)
	c.Assert(err, NotNil)

	data = []string{"location / {"}

	_, _, err = parseServerBlock(data, 1)
	c.Assert(err, NotNil)
}

func (s *NginxSuite) TestHTTPBlockParser(c *C) {
	data := []string{"types {", "text/xml xml;"}

	_, _, err := parseHTTPBlock(data, 0)
	c.Assert(err, NotNil)

	data = []string{"server {", "location / {"}

	_, _, err = parseHTTPBlock(data, 0)
	c.Assert(err, NotNil)

	data = []string{"upstream {", "server 127.0.0.1"}

	_, _, err = parseHTTPBlock(data, 0)
	c.Assert(err, NotNil)

	data = []string{"unknown {", "server 127.0.0.1"}

	_, _, err = parseHTTPBlock(data, 0)
	c.Assert(err, NotNil)

	data = []string{"upstream test123 {", "server 127.0.0.1"}

	_, _, err = parseHTTPBlock(data, 0)
	c.Assert(err, NotNil)
}

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *NginxSuite) TestHelpers(c *C) {
	config, err := Read("testdata/webkaos.conf", "")

	c.Assert(err, IsNil)
	c.Assert(config, NotNil)

	var http *HTTP
	var p Properties
	var cp *ConditionalProperties

	c.Assert(http.ServersNum(), Equals, 0)
	c.Assert(http.ServersList(), IsNil)
	c.Assert(http.FindServer("unknown", "http"), IsNil)
	c.Assert(cp.Get("unknown"), Equals, "")
	c.Assert(p.Get("unknown"), Equals, "")

	c.Assert(config.String(), Not(Equals), "")
	c.Assert(config.HTTP.ServersNum(), Equals, 3)
	c.Assert(
		config.HTTP.ServersList(), DeepEquals,
		[]string{"_:http", "service.domain.com:http", "service.domain.com:https"},
	)

	c.Assert(config.HTTP.FindServer("unknown", "http"), IsNil)
	c.Assert(config.HTTP.FindServer("service.domain.com", "https"), NotNil)
	c.Assert(config.HTTP.Properties.Get("client_body_timeout"), Equals, "15s")

	server := config.HTTP.FindServer("service.domain.com", "https")

	c.Assert(server.Properties.Get("unknown"), Equals, "")
	c.Assert(config.HTTP.Properties.Get("unknown"), Equals, "")
}

func (s *NginxSuite) TestPropsGetters(c *C) {
	p := make(Properties)
	p["bool1"] = append(p["bool1"], "on")
	p["bool2"] = append(p["bool2"], "off")
	p["bool3"] = append(p["bool3"], "auto")
	p["numeric1"] = append(p["numeric1"], "441")
	p["numeric2"] = append(p["numeric2"], "abc")
	p["buffer1"] = append(p["buffer1"], "4 16k")
	p["buffer2"] = append(p["buffer2"], "A 16k")
	p["buffer3"] = append(p["buffer3"], "4 A")
	p["buffer4"] = append(p["buffer4"], "123")
	p["size1"] = append(p["size1"], "160")
	p["size2"] = append(p["size2"], "512k")
	p["size3"] = append(p["size3"], "8M")
	p["size4"] = append(p["size4"], "2G")
	p["size5"] = append(p["size5"], "2J")
	p["time0"] = append(p["time0"], "48")
	p["time1"] = append(p["time1"], "230ms")
	p["time2"] = append(p["time2"], "120s")
	p["time3"] = append(p["time3"], "3m")
	p["time4"] = append(p["time4"], "6h")
	p["time5"] = append(p["time5"], "3d")
	p["time6"] = append(p["time6"], "2w")
	p["time7"] = append(p["time7"], "2M")
	p["time8"] = append(p["time8"], "1y")
	p["time9"] = append(p["time9"], "2d 6h 30m 15s")
	p["time10"] = append(p["time10"], "3u")

	c.Assert(p.Get("bool3"), Equals, "auto")

	b1, e1 := p.GetBool("bool1")
	b2, e2 := p.GetBool("bool2")
	b3, e3 := p.GetBool("bool3")

	c.Assert(b1, Equals, true)
	c.Assert(e1, IsNil)
	c.Assert(b2, Equals, false)
	c.Assert(e2, IsNil)
	c.Assert(b3, Equals, false)
	c.Assert(e3, NotNil)

	n1, e1 := p.GetInt("numeric1")
	n2, e2 := p.GetInt("numeric2")

	c.Assert(n1, Equals, int64(441))
	c.Assert(e1, IsNil)
	c.Assert(n2, Equals, int64(0))
	c.Assert(e2, NotNil)

	bfn1, bfs1, e1 := p.GetBuf("buffer1")
	bfn2, bfs2, e2 := p.GetBuf("buffer2")
	bfn3, bfs3, e3 := p.GetBuf("buffer3")
	bfn4, bfs4, e4 := p.GetBuf("buffer4")

	c.Assert(bfn1, Equals, int64(4))
	c.Assert(bfs1, Equals, int64(16384))
	c.Assert(e1, IsNil)
	c.Assert(bfn2, Equals, int64(0))
	c.Assert(bfs2, Equals, int64(0))
	c.Assert(e2, NotNil)
	c.Assert(bfn3, Equals, int64(0))
	c.Assert(bfs3, Equals, int64(0))
	c.Assert(e3, NotNil)
	c.Assert(bfn4, Equals, int64(0))
	c.Assert(bfs4, Equals, int64(0))
	c.Assert(e4, NotNil)

	s1, e1 := p.GetSize("size1")
	s2, e2 := p.GetSize("size2")
	s3, e3 := p.GetSize("size3")
	s4, e4 := p.GetSize("size4")
	s5, e5 := p.GetSize("size5")

	c.Assert(s1, Equals, int64(160))
	c.Assert(e1, IsNil)
	c.Assert(s2, Equals, int64(524288))
	c.Assert(e2, IsNil)
	c.Assert(s3, Equals, int64(8388608))
	c.Assert(e3, IsNil)
	c.Assert(s4, Equals, int64(2147483648))
	c.Assert(e4, IsNil)
	c.Assert(s5, Equals, int64(0))
	c.Assert(e5, NotNil)

	t0, e0 := p.GetTime("time0")
	t1, e1 := p.GetTime("time1")
	t2, e2 := p.GetTime("time2")
	t3, e3 := p.GetTime("time3")
	t4, e4 := p.GetTime("time4")
	t5, e5 := p.GetTime("time5")
	t6, e6 := p.GetTime("time6")
	t7, e7 := p.GetTime("time7")
	t8, e8 := p.GetTime("time8")
	t9, e9 := p.GetTime("time9")
	t10, e10 := p.GetTime("time10")

	c.Assert(t0, Equals, 48*time.Second)
	c.Assert(e0, IsNil)
	c.Assert(t1, Equals, 230*time.Millisecond)
	c.Assert(e1, IsNil)
	c.Assert(t2, Equals, 120*time.Second)
	c.Assert(e2, IsNil)
	c.Assert(t3, Equals, 3*time.Minute)
	c.Assert(e3, IsNil)
	c.Assert(t4, Equals, 6*time.Hour)
	c.Assert(e4, IsNil)
	c.Assert(t5, Equals, 72*time.Hour)
	c.Assert(e5, IsNil)
	c.Assert(t6, Equals, 14*24*time.Hour)
	c.Assert(e6, IsNil)
	c.Assert(t7, Equals, 2*30*24*time.Hour)
	c.Assert(e7, IsNil)
	c.Assert(t8, Equals, 365*24*time.Hour)
	c.Assert(e8, IsNil)
	c.Assert(t9, Equals, 54*time.Hour+30*time.Minute+15*time.Second)
	c.Assert(e9, IsNil)
	c.Assert(t10, Equals, time.Duration(0))
	c.Assert(e10, NotNil)

	_, err := p.GetBool("unknown")
	c.Assert(err, NotNil)
	_, err = p.GetInt("unknown")
	c.Assert(err, NotNil)
	_, _, err = p.GetBuf("unknown")
	c.Assert(err, NotNil)
	_, err = p.GetSize("unknown")
	c.Assert(err, NotNil)
	_, err = p.GetTime("unknown")
	c.Assert(err, NotNil)
}

func (s *NginxSuite) TestCondPropsGetters(c *C) {
	p := ConditionalProperties{[]string{}, make(map[string][]ConditionalProperty)}
	p.Data["bool1"] = append(p.Data["bool1"], ConditionalProperty{-1, "on"})
	p.Data["bool2"] = append(p.Data["bool2"], ConditionalProperty{-1, "off"})
	p.Data["bool3"] = append(p.Data["bool3"], ConditionalProperty{-1, "auto"})
	p.Data["numeric1"] = append(p.Data["numeric1"], ConditionalProperty{-1, "441"})
	p.Data["numeric2"] = append(p.Data["numeric2"], ConditionalProperty{-1, "abc"})
	p.Data["buffer1"] = append(p.Data["buffer1"], ConditionalProperty{-1, "4 16k"})
	p.Data["buffer2"] = append(p.Data["buffer2"], ConditionalProperty{-1, "123"})
	p.Data["size1"] = append(p.Data["size1"], ConditionalProperty{-1, "160"})
	p.Data["size2"] = append(p.Data["size2"], ConditionalProperty{-1, "512k"})
	p.Data["size3"] = append(p.Data["size3"], ConditionalProperty{-1, "8M"})
	p.Data["size4"] = append(p.Data["size4"], ConditionalProperty{-1, "2G"})
	p.Data["size5"] = append(p.Data["size5"], ConditionalProperty{-1, "2J"})
	p.Data["time0"] = append(p.Data["time0"], ConditionalProperty{-1, "48"})
	p.Data["time1"] = append(p.Data["time1"], ConditionalProperty{-1, "230ms"})
	p.Data["time2"] = append(p.Data["time2"], ConditionalProperty{-1, "120s"})
	p.Data["time3"] = append(p.Data["time3"], ConditionalProperty{-1, "3m"})
	p.Data["time4"] = append(p.Data["time4"], ConditionalProperty{-1, "6h"})
	p.Data["time5"] = append(p.Data["time5"], ConditionalProperty{-1, "3d"})
	p.Data["time6"] = append(p.Data["time6"], ConditionalProperty{-1, "2w"})
	p.Data["time7"] = append(p.Data["time7"], ConditionalProperty{-1, "2M"})
	p.Data["time8"] = append(p.Data["time8"], ConditionalProperty{-1, "1y"})
	p.Data["time9"] = append(p.Data["time9"], ConditionalProperty{-1, "2d 6h 30m 15s"})
	p.Data["time10"] = append(p.Data["time10"], ConditionalProperty{-1, "3u"})

	c.Assert(p.Get("bool3"), Equals, "auto")

	b1, e1 := p.GetBool("bool1")
	b2, e2 := p.GetBool("bool2")
	b3, e3 := p.GetBool("bool3")

	c.Assert(b1, Equals, true)
	c.Assert(e1, IsNil)
	c.Assert(b2, Equals, false)
	c.Assert(e2, IsNil)
	c.Assert(b3, Equals, false)
	c.Assert(e3, NotNil)

	n1, e1 := p.GetInt("numeric1")
	n2, e2 := p.GetInt("numeric2")

	c.Assert(n1, Equals, int64(441))
	c.Assert(e1, IsNil)
	c.Assert(n2, Equals, int64(0))
	c.Assert(e2, NotNil)

	bfn1, bfs1, e1 := p.GetBuf("buffer1")
	bfn2, bfs2, e2 := p.GetBuf("buffer2")
	bfn3, bfs3, e3 := p.GetBuf("buffer3")
	bfn4, bfs4, e4 := p.GetBuf("buffer4")

	c.Assert(bfn1, Equals, int64(4))
	c.Assert(bfs1, Equals, int64(16384))
	c.Assert(e1, IsNil)
	c.Assert(bfn2, Equals, int64(0))
	c.Assert(bfs2, Equals, int64(0))
	c.Assert(e2, NotNil)
	c.Assert(bfn3, Equals, int64(0))
	c.Assert(bfs3, Equals, int64(0))
	c.Assert(e3, NotNil)
	c.Assert(bfn4, Equals, int64(0))
	c.Assert(bfs4, Equals, int64(0))
	c.Assert(e4, NotNil)

	s1, e1 := p.GetSize("size1")
	s2, e2 := p.GetSize("size2")
	s3, e3 := p.GetSize("size3")
	s4, e4 := p.GetSize("size4")
	s5, e5 := p.GetSize("size5")

	c.Assert(s1, Equals, int64(160))
	c.Assert(e1, IsNil)
	c.Assert(s2, Equals, int64(524288))
	c.Assert(e2, IsNil)
	c.Assert(s3, Equals, int64(8388608))
	c.Assert(e3, IsNil)
	c.Assert(s4, Equals, int64(2147483648))
	c.Assert(e4, IsNil)
	c.Assert(s5, Equals, int64(0))
	c.Assert(e5, NotNil)

	t0, e0 := p.GetTime("time0")
	t1, e1 := p.GetTime("time1")
	t2, e2 := p.GetTime("time2")
	t3, e3 := p.GetTime("time3")
	t4, e4 := p.GetTime("time4")
	t5, e5 := p.GetTime("time5")
	t6, e6 := p.GetTime("time6")
	t7, e7 := p.GetTime("time7")
	t8, e8 := p.GetTime("time8")
	t9, e9 := p.GetTime("time9")
	t10, e10 := p.GetTime("time10")

	c.Assert(t0, Equals, 48*time.Second)
	c.Assert(e0, IsNil)
	c.Assert(t1, Equals, 230*time.Millisecond)
	c.Assert(e1, IsNil)
	c.Assert(t2, Equals, 120*time.Second)
	c.Assert(e2, IsNil)
	c.Assert(t3, Equals, 3*time.Minute)
	c.Assert(e3, IsNil)
	c.Assert(t4, Equals, 6*time.Hour)
	c.Assert(e4, IsNil)
	c.Assert(t5, Equals, 72*time.Hour)
	c.Assert(e5, IsNil)
	c.Assert(t6, Equals, 14*24*time.Hour)
	c.Assert(e6, IsNil)
	c.Assert(t7, Equals, 2*30*24*time.Hour)
	c.Assert(e7, IsNil)
	c.Assert(t8, Equals, 365*24*time.Hour)
	c.Assert(e8, IsNil)
	c.Assert(t9, Equals, 54*time.Hour+30*time.Minute+15*time.Second)
	c.Assert(e9, IsNil)
	c.Assert(t10, Equals, time.Duration(0))
	c.Assert(e10, NotNil)

	_, err := p.GetBool("unknown")
	c.Assert(err, NotNil)
	_, err = p.GetInt("unknown")
	c.Assert(err, NotNil)
	_, _, err = p.GetBuf("unknown")
	c.Assert(err, NotNil)
	_, err = p.GetSize("unknown")
	c.Assert(err, NotNil)
	_, err = p.GetTime("unknown")
	c.Assert(err, NotNil)
}
