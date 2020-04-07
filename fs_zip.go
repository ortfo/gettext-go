// Copyright 2013 ChaiShushan <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gettext

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
)

type zipFS struct {
	root string
	name string
	r    *zip.Reader
}

func newZipFS(r *zip.Reader, name string) *zipFS {
	fs := &zipFS{r: r, name: name}
	fs.root = fs.zipName()
	return fs
}

func (p *zipFS) zipName() string {
	name := p.name
	if x := strings.LastIndexAny(name, `\/`); x != -1 {
		name = name[x+1:]
	}
	name = strings.TrimSuffix(name, ".zip")
	return name
}

func (p *zipFS) LocaleList() []string {
	var locals []string
	for s := range p.lsZip(p.r) {
		locals = append(locals, s)
	}
	sort.Strings(locals)
	return locals
}

func (p *zipFS) DomainList(locale string) []string {
	var domainMap = make(map[string]string)
	for _, f := range p.r.File {
		if strings.HasSuffix(f.Name, ".mo") || strings.HasSuffix(f.Name, ".po") {
			if strings.Contains(f.Name, locale+"/LC_MESSAGES") {
				domain := f.Name[strings.LastIndex(f.Name, "/")+1 : strings.LastIndex(f.Name, ".")]
				domainMap[domain] = domain
			}
		}
	}

	var keys []string
	for _, s := range domainMap {
		keys = append(keys, s)
	}
	sort.Strings(keys)
	return keys
}

func (p *zipFS) LoadMessagesFile(domain, local, ext string) ([]byte, error) {
	trName := p.makeMessagesFileName(domain, local, ext)
	for _, f := range p.r.File {
		if f.Name != trName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		rcData, err := ioutil.ReadAll(rc)
		rc.Close()
		return rcData, err
	}
	return nil, fmt.Errorf("not found")

}

func (p *zipFS) LoadResourceFile(domain, local, name string) ([]byte, error) {
	rcName := p.makeResourceFileName(domain, local, name)
	for _, f := range p.r.File {
		if f.Name != rcName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		rcData, err := ioutil.ReadAll(rc)
		rc.Close()
		return rcData, err
	}
	return nil, fmt.Errorf("not found")
}

func (p *zipFS) String() string {
	return "gettext.zipfs(" + p.name + ")"
}

func (p *zipFS) makeMessagesFileName(domain, local, ext string) string {
	return fmt.Sprintf("%s/%s/LC_MESSAGES/%s%s", p.root, local, domain, ext)
}

func (p *zipFS) makeResourceFileName(domain, local, name string) string {
	return fmt.Sprintf("%s/%s/LC_RESOURCE/%s/%s", p.root, local, domain, name)
}

func (p *zipFS) lsZip(r *zip.Reader) map[string]bool {
	ssMap := make(map[string]bool)
	for _, f := range r.File {
		if x := strings.Index(f.Name, "LC_MESSAGES"); x != -1 {
			s := strings.TrimRight(f.Name[:x], `\/`)
			if x = strings.LastIndexAny(s, `\/`); x != -1 {
				s = s[x+1:]
			}
			if s != "" {
				ssMap[s] = true
			}
			continue
		}
		if x := strings.Index(f.Name, "LC_RESOURCE"); x != -1 {
			s := strings.TrimRight(f.Name[:x], `\/`)
			if x = strings.LastIndexAny(s, `\/`); x != -1 {
				s = s[x+1:]
			}
			if s != "" {
				ssMap[s] = true
			}
			continue
		}
	}
	return ssMap
}
