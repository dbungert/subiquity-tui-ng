package ui

import "testing"

func TestDetectHalfBlockSupport(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{
			name: "blocks override forces true even on linux tty",
			env:  map[string]string{"SUBIQUITY_NG_HEADER": "blocks", "TERM": "linux"},
			want: true,
		},
		{
			name: "plain override forces false on utf-8 xterm",
			env:  map[string]string{"SUBIQUITY_NG_HEADER": "plain", "TERM": "xterm-256color", "LANG": "en_US.UTF-8"},
			want: false,
		},
		{
			name: "linux tty falls back",
			env:  map[string]string{"TERM": "linux", "LANG": "en_US.UTF-8"},
			want: false,
		},
		{
			name: "dumb terminal falls back",
			env:  map[string]string{"TERM": "dumb", "LANG": "en_US.UTF-8"},
			want: false,
		},
		{
			name: "empty TERM falls back",
			env:  map[string]string{"TERM": "", "LANG": "en_US.UTF-8"},
			want: false,
		},
		{
			name: "utf-8 xterm enables blocks",
			env:  map[string]string{"TERM": "xterm-256color", "LANG": "en_US.UTF-8"},
			want: true,
		},
		{
			name: "non-utf-8 locale on capable terminal falls back",
			env:  map[string]string{"TERM": "xterm-256color", "LANG": "C"},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetLocaleEnv(t)
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			if got := detectHalfBlockSupport(); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsUTF8Locale(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"all empty", map[string]string{}, false},
		{"LC_ALL utf-8", map[string]string{"LC_ALL": "en_US.UTF-8"}, true},
		{"LC_ALL=C masks LANG=UTF-8", map[string]string{"LC_ALL": "C", "LANG": "en_US.UTF-8"}, false},
		{"LC_CTYPE utf8 when LC_ALL empty", map[string]string{"LC_CTYPE": "en_US.utf8"}, true},
		{"LANG utf-8 only", map[string]string{"LANG": "C.UTF-8"}, true},
		{"LANG POSIX", map[string]string{"LANG": "POSIX"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetLocaleEnv(t)
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			if got := isUTF8Locale(); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func resetLocaleEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"SUBIQUITY_NG_HEADER", "TERM", "LANG", "LC_ALL", "LC_CTYPE"} {
		t.Setenv(k, "")
	}
}
