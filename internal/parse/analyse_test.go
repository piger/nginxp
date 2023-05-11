package parse

import (
	"reflect"
	"testing"
)

func TestNewConfiguration(t *testing.T) {
	filename := "testdata/nginx.conf"
	files, err := Unpack(filename)
	if err != nil {
		t.Fatalf("failed unpacking %q: %s", filename, err)
	}

	tree, err := Parse(filename, files[filename])
	if err != nil {
		t.Fatalf("failed parsing %q: %s", filename, err)
	}

	cfg, err := NewConfiguration(tree)
	if err != nil {
		t.Fatalf("failed analysing %q: %s", filename, err)
	}

	if cfg.Filename != filename {
		t.Fatalf("cfg.Filename: expected %q, got %q", filename, cfg.Filename)
	}

	findDirective := func(name string, dirs []*Directive) *Directive {
		for _, dir := range dirs {
			if dir.Name == name {
				return dir
			}
		}
		return nil
	}

	httpDirective := findDirective("http", cfg.Directives)
	if httpDirective == nil {
		t.Fatal("directive 'http' not found")
	}

	var servers []*Directive
	for _, dir := range httpDirective.Block {
		if dir.Name == "server" {
			servers = append(servers, dir)
		}
	}

	if len(servers) != 3 {
		t.Fatalf("expected 3 servers, found %d", len(servers))
	}

	server0 := servers[0]
	listenDirective := findDirective("listen", server0.Block)
	if listenDirective == nil {
		t.Fatalf("directive 'listen' not found in %+v", server0)
	}

	if listenDirective.Args[0] != "80" {
		t.Fatalf("expected listen 80, found: %q", listenDirective.Args)
	}

	domains := []string{"domain1.com", "www.domain1.com"}
	serverNameDirective := findDirective("server_name", server0.Block)
	if serverNameDirective == nil {
		t.Fatalf("directive 'server_name' not found in %+v", server0)
	}

	if !reflect.DeepEqual(serverNameDirective.Args, domains) {
		t.Fatalf("server_name: expected %q, got %q", domains, serverNameDirective.Args)
	}
}
