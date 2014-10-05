//
// SysDB - newsconv
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

// newsconv is the RSS, Atom, and news generator of the SysDB website. It
// convert the JSON representation of the news items into the target formats.
package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gorilla/feeds"
)

var force = flag.Bool("force", false, "force overwriting pages")
var configfile = flag.String("configfile", "news/news.conf", "config file name")
var output = flag.String("output", "/var/www/sysdb.io", "output directory")

func main() {
	flag.Parse()

	f, err := os.Open(*configfile)
	if err != nil {
		log.Fatalf("Failed to open config file %q: %v", *configfile, err)
	}
	defer f.Close()
	log.Printf("Reading config file %q ...", *configfile)

	news := &feeds.Feed{
		Title:       "SysDB — System DataBase",
		Link:        &feeds.Link{Href: "https://sysdb.io"},
		Description: "The system management and inventory collection service",
		Author:      &feeds.Author{"Sebastian 'tokkee' Harl", "tokkee@sysdb.io"},
		Created:     time.Date(2014, time.June, 1, 17, 56, 19, 0, time.FixedZone("Europe/Berlin", 7200)),
		Copyright:   "Copyright © 2014 SysDB Project",
	}
	releases := &feeds.Feed{
		Title:       "SysDB Releases",
		Link:        &feeds.Link{Href: "https://sysdb.io/download"},
		Description: "Releases of the system management and inventory collection service",
		Author:      &feeds.Author{"Sebastian 'tokkee' Harl", "tokkee@sysdb.io"},
		Created:     time.Date(2014, time.June, 1, 17, 56, 19, 0, time.FixedZone("Europe/Berlin", 7200)),
		Copyright:   "Copyright © 2014 SysDB Project",
	}

	dec := json.NewDecoder(f)
	err = dec.Decode(&news.Items)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	news.Updated = news.Created
	releases.Updated = releases.Created
	for _, item := range news.Items {
		if item.Updated.IsZero() {
			item.Updated = item.Created
		}
		if item.Updated.After(news.Updated) {
			news.Updated = item.Updated
		}

		if !strings.HasPrefix(item.Link.Href, "http") {
			item.Link.Href = "https://sysdb.io/" + item.Link.Href
		}

		if item.Author == nil {
			item.Author = news.Author
		}

		// In order to be able to use the feeds.Item type for decoding, we
		// misuse the Id field to mark release entries.
		if item.Id == "release" {
			releases.Items = append(releases.Items, item)
			item.Id = ""

			if item.Updated.After(releases.Updated) {
				releases.Updated = item.Updated
			}
		}
	}

	for _, page := range []struct {
		file string
		gen  func(io.Writer) error
	}{
		{*output + "/news.rss", news.WriteRss},
		{*output + "/news.atom", news.WriteAtom},
		{*output + "/releases.rss", releases.WriteRss},
		{*output + "/releases.atom", releases.WriteAtom},
	} {
		if _, err := os.Stat(page.file); err == nil && !*force {
			log.Fatalf("Page %q exists already", page.file)
		}

		if err := os.MkdirAll(path.Dir(page.file), 0755); err != nil {
			log.Fatalf("Failed to create directory %q: %v", path.Dir(page.file), err)
		}

		out, err := os.Create(page.file)
		if err != nil {
			log.Fatalf("Failed to write page %q: %v", page.file, err)
		}
		defer out.Close()

		log.Printf("Writing page %q ...", page.file)
		if err := page.gen(out); err != nil {
			log.Fatalf("Failed to generate page %q: %v", page.file, err)
		}
	}
}
