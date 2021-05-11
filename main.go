package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gernest/front"
)

type page struct {
	Weight int
	Title  string
	Path   string

	Children []*page
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	base := flag.String("path", "docs/docs", "")
	dir := flag.String("dir", "eventing", "")
	flag.Parse()

	m := front.NewMatter()
	m.Handle("---", front.YAMLHandler)

	dirs := make(map[string]*page)
	if err := filepath.WalkDir(filepath.Join(*base, *dir), func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		meta, _, err := m.Parse(f)
		if err != nil {
			log.Printf("frontmatter parse %s: %v", path, err)
			meta = make(map[string]interface{})
		}

		rel, err := filepath.Rel(*base, path)
		if err != nil {
			return err
		}

		p := &page{Weight: 0, Path: rel}
		title, ok := meta["title"]
		if !ok {
			log.Printf("no title for page at %s", path)
			return nil
		}

		p.Title = title.(string)
		if lt, ok := meta["linkTitle"]; ok {
			p.Title = lt.(string)
		}

		if weight, ok := meta["weight"]; ok {
			if p.Weight, ok = weight.(int); !ok {
				log.Printf("unexpected value for weight in path %s: %T", path, weight)
			}
		}

		dir, ok := dirs[filepath.Dir(path)]
		if !ok {
			dir = &page{}
			dirs[filepath.Dir(path)] = dir

			dirdir := filepath.Dir(filepath.Dir(path))
			if parent, ok := dirs[dirdir]; ok {
				parent.Children = append(parent.Children, dir)
			}
		}

		if d.Name() == "index.md" || d.Name() == "README.md" {
			dir.Title = p.Title
			dir.Weight = p.Weight
			dir.Path = p.Path
		} else {
			dir.Children = append(dir.Children, p)
		}

		return nil
	}); err != nil {
		return err
	}

	root := dirs[filepath.Clean(filepath.Join(*base, *dir))]
	walk(root, 0)

	return nil
}

func walk(p *page, indent int) {
	if len(p.Children) == 0 {
		fmt.Println(strings.Repeat("  ", indent) + "- " + p.Title + ": " + p.Path)
		return
	}

	fmt.Println(strings.Repeat("  ", indent) + "- " + p.Title + ":")
	fmt.Println(strings.Repeat("  ", indent) + "  - Overview: " + p.Path)
	sort.Stable(byWeight(p.Children))
	for _, c := range p.Children {
		walk(c, indent+1)
	}
}

type byWeight []*page

func (bw byWeight) Len() int {
	return len(bw)
}

func (bw byWeight) Less(i, j int) bool {
	return bw[i].Weight < bw[j].Weight
}

func (bw byWeight) Swap(i, j int) {
	bw[i], bw[j] = bw[j], bw[i]
}
