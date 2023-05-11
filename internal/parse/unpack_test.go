package parse

import (
	"testing"
)

func TestUnpackFilenames(t *testing.T) {
	tests := []struct {
		Filename      string
		ExpectedFiles []string
	}{
		{
			Filename: "testdata/docker_nginx_t.conf",
			ExpectedFiles: []string{
				"/etc/nginx/nginx.conf",
				"/etc/nginx/mime.types",
				"/etc/nginx/conf.d/default.conf",
			},
		},
		{
			Filename: "testdata/nginx.conf",
			ExpectedFiles: []string{
				"testdata/nginx.conf",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Filename, func(t *testing.T) {
			files, err := Unpack(tt.Filename)
			if err != nil {
				t.Fatalf("error unpacking %q: %s", tt.Filename, err)
			}

			for _, expected := range tt.ExpectedFiles {
				if _, ok := files[expected]; !ok {
					t.Fatalf("expected %q in %s not found", expected, tt.Filename)
				}
			}
		})
	}
}
