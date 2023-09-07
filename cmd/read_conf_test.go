package cmd

import (
	"testing"
)

func TestPastaConfValidate(t *testing.T) {
	tests := []struct {
		name    string
		conf    *pastaConf
		wantErr bool
	}{
		{
			name: "valid config",
			conf: &pastaConf{
				Deps: []copierConf{
					{
						URL:     "https://example.com",
						From:    "path/to/source/",
						To:      "path/to/destination/",
						Include: "*.go",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing url",
			conf: &pastaConf{
				Deps: []copierConf{
					{
						From: "path/to/source/",
						To:   "path/to/destination/",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing from",
			conf: &pastaConf{
				Deps: []copierConf{
					{
						URL: "https://example.com",
						To:  "path/to/destination/",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing to",
			conf: &pastaConf{
				Deps: []copierConf{
					{
						URL:  "https://example.com",
						From: "path/to/source/",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid from",
			conf: &pastaConf{
				Deps: []copierConf{
					{
						URL:  "https://example.com",
						From: "/path/to/source/",
						To:   "path/to/destination/",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid to",
			conf: &pastaConf{
				Deps: []copierConf{
					{
						URL:  "https://example.com",
						From: "path/to/source/",
						To:   "/path/to/destination/",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid dot to with files",
			conf: &pastaConf{
				Deps: []copierConf{
					{
						URL:   "https://example.com",
						From:  "path/to/source/",
						To:    ".",
						Files: []string{"file1.txt", "file2.txt"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid to with files",
			conf: &pastaConf{
				Deps: []copierConf{
					{
						URL:   "https://example.com",
						From:  "path/to/source/",
						To:    "path/to/destination/",
						Files: []string{"file1.txt", "file2.txt"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid to with leading slash",
			conf: &pastaConf{
				Deps: []copierConf{
					{
						URL:  "https://example.com",
						From: "path/to/source/",
						To:   "/path/to/destination/",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid to with dot",
			conf: &pastaConf{
				Deps: []copierConf{
					{
						URL:  "https://example.com",
						From: "path/to/source/",
						To:   ".",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid to with leading slash and dot",
			conf: &pastaConf{
				Deps: []copierConf{
					{
						URL:  "https://example.com",
						From: "path/to/source/",
						To:   "/.",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid files with include and exclude",
			conf: &pastaConf{
				Deps: []copierConf{
					{
						URL:     "https://example.com",
						From:    "path/to/source/",
						To:      "path/to/destination/",
						Include: "*.go",
						Exclude: "*.txt",
						Files:   []string{"file1.txt", "file2.txt"},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conf.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
