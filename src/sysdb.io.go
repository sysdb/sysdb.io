//
// SysDB - sysdb.io
// Copyright (C) 2014 Sebastian 'tokkee' Harl <sh@tokkee.org>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// ``AS IS'' AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED
// TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
// PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDERS OR
// CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
// EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS;
// OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY,
// WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR
// OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF
// ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//

// sysdb.io is a template-based generator of the SysDB website
package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
)

var force = flag.Bool("force", false, "force overwriting pages")
var configfile = flag.String("configfile", "sysdb.conf", "config file name")
var output = flag.String("output", "/var/www/sysdb.io", "output directory")

type templ struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Template    string   `json:"template"`
	JS          []string `json:"javascript"`
	Left        string   `json:"left"`
	Right       string   `json:"right"`
}

type config struct {
	Defaults templ            `json:"defaults"`
	Pages    map[string]templ `json:"pages"`
}

func (c config) title(page string) string {
	if t, ok := c.Pages[page]; ok && t.Title != "" {
		return t.Title
	}
	if c.Defaults.Title == "" {
		log.Fatalf("Title q for page %q not found.", page)
	}
	return c.Defaults.Title
}

func (c config) description(page string) string {
	if t, ok := c.Pages[page]; ok && t.Description != "" {
		return t.Description
	}
	if c.Defaults.Description == "" {
		log.Fatalf("Description q for page %q not found.", page)
	}
	return c.Defaults.Description
}

func (c config) js(page string) []string {
	js := c.Defaults.JS
	if t, ok := c.Pages[page]; ok && len(t.JS) != 0 {
		js = append(js, t.JS...)
	}
	return js
}

func (c config) template(page string) string {
	if t, ok := c.Pages[page]; ok && t.Template != "" {
		return t.Template
	}
	if c.Defaults.Template == "" {
		log.Fatalf("Template %q for page not found.", page)
	}
	return c.Defaults.Template
}

func (c config) left(page string) string {
	if t, ok := c.Pages[page]; ok && t.Left != "" {
		return t.Left
	}
	if c.Defaults.Left == "" {
		log.Fatalf("Left template for page %q not found.", page)
	}
	return c.Defaults.Left
}

func (c config) right(page string) string {
	if t, ok := c.Pages[page]; ok && t.Right != "" {
		return t.Right
	}
	if c.Defaults.Right == "" {
		log.Fatalf("Right template for page %q not found.", page)
	}
	return c.Defaults.Right
}

func parseTempl(templ *template.Template, page, name, filename string) *template.Template {
	var t *template.Template
	if templ == nil {
		templ = template.New(name)
		t = templ
	} else {
		t = templ.New(name)
	}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read template %q: %v", filename, err)
	}

	log.Printf("Parsing template %q for page %q:%q ...", filename, page, name)
	_, err = t.Parse(string(content))
	if err != nil {
		log.Fatalf("Failed to parse template %q: %v", filename, err)
	}
	return templ
}

func writePage(templ *template.Template, page string, c *config) {
	file := path.Join(*output, page)

	if _, err := os.Stat(file); err == nil && !*force {
		log.Fatalf("Page %q exists already.", file)
	}

	if err := os.MkdirAll(path.Dir(file), 0755); err != nil {
		log.Fatalf("Failed to create directory %q: %v", path.Dir(file), err)
	}

	out, err := os.Create(file)
	if err != nil {
		log.Fatalf("Failed to write page %q: %v", file, err)
	}
	defer out.Close()

	p := &struct {
		Title       string
		Description string
		JS          []string
	}{
		Title:       c.title(page),
		Description: c.description(page),
		JS:          c.js(page),
	}

	log.Printf("Writing page %q ...", file)
	err = templ.Execute(out, p)
	if err != nil {
		log.Fatalf("Failed to execute template for page %q: %v", file, err)
	}
}

func main() {
	flag.Parse()

	f, err := os.Open(*configfile)
	if err != nil {
		log.Fatalf("Failed to open config file %q: %v", *configfile, err)
	}
	defer f.Close()
	log.Printf("Reading config file %q ...", *configfile)

	c := &config{}

	dec := json.NewDecoder(f)
	err = dec.Decode(c)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	for page := range c.Pages {
		templ := parseTempl(nil, page, "main", c.template(page))
		parseTempl(templ, page, "left", c.left(page))
		parseTempl(templ, page, "right", c.right(page))

		writePage(templ, page, c)
	}
}
