package certacme

import "testing"

func TestExtractEmailFromContacts(t *testing.T) {
	tests := []struct {
		name     string
		contacts []string
		want     string
	}{
		{"empty", nil, ""},
		{"mailto", []string{"mailto:user@example.com"}, "user@example.com"},
		{"mixed", []string{"tel:+1", "mailto:u@x.com"}, "u@x.com"},
		{"whitespace", []string{"  mailto:  spaced@x.com  "}, "spaced@x.com"},
		{"no mailto", []string{"user@example.com"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractEmailFromContacts(tt.contacts)
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestResolveImportDirectoryURL(t *testing.T) {
	dir, err := resolveImportDirectoryURL("letsencrypt", "")
	if err != nil {
		t.Fatal(err)
	}
	if dir != "https://acme-v02.api.letsencrypt.org/directory" {
		t.Fatalf("dir = %q", dir)
	}

	dir, err = resolveImportDirectoryURL("letsencrypt", "https://custom.example/directory")
	if err != nil {
		t.Fatal(err)
	}
	if dir != "https://custom.example/directory" {
		t.Fatalf("override dir = %q", dir)
	}

	_, err = resolveImportDirectoryURL("acmeca", "")
	if err == nil {
		t.Fatal("expected error for acmeca without dir")
	}
}
