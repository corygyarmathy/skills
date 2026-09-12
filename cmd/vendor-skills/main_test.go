package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeSkill lays out a vendored skill directory and returns its hash.
func writeSkill(t *testing.T, dir string, files map[string]string) string {
	t.Helper()
	for name, body := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	hash, err := hashSkill(dir)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}

func TestHashSkill(t *testing.T) {
	base := map[string]string{"SKILL.md": "body\n", "ref/notes.md": "notes\n"}

	tests := []struct {
		name  string
		files map[string]string
		same  bool
	}{
		{
			name:  "identical trees hash the same",
			files: map[string]string{"SKILL.md": "body\n", "ref/notes.md": "notes\n"},
			same:  true,
		},
		{
			name:  "a changed skill body is a different hash",
			files: map[string]string{"SKILL.md": "edited\n", "ref/notes.md": "notes\n"},
		},
		{
			name:  "a changed reference file beside it is a different hash",
			files: map[string]string{"SKILL.md": "body\n", "ref/notes.md": "edited\n"},
		},
		{
			name:  "a removed reference file is a different hash",
			files: map[string]string{"SKILL.md": "body\n"},
		},
		{
			name:  "an added file is a different hash",
			files: map[string]string{"SKILL.md": "body\n", "ref/notes.md": "notes\n", "extra.md": "x\n"},
		},
	}

	want := writeSkill(t, t.TempDir(), base)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeSkill(t, t.TempDir(), tt.files)
			if (got == want) != tt.same {
				t.Errorf("hash equal = %v, want %v", got == want, tt.same)
			}
		})
	}
}

func TestHashSkill_MissingDirectory(t *testing.T) {
	if _, err := hashSkill(filepath.Join(t.TempDir(), "absent")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("err = %v, want a not-exist error so verify can report it as missing", err)
	}
}

// consumer builds a repository whose vendored copy agrees with its lock, then
// lets the test break it one way at a time.
func consumer(t *testing.T) (root string, lock *Lock) {
	t.Helper()
	root = t.TempDir()
	vendor := filepath.Join(root, ".agents", "skills")
	links := filepath.Join(root, ".claude", "skills")
	if err := os.MkdirAll(links, 0o755); err != nil {
		t.Fatal(err)
	}
	hash := writeSkill(t, filepath.Join(vendor, "tdd"), map[string]string{"SKILL.md": "body\n"})
	if err := os.Symlink(filepath.Join("..", "..", ".agents", "skills", "tdd"), filepath.Join(links, "tdd")); err != nil {
		t.Fatal(err)
	}
	return root, &Lock{
		Version:    2,
		Source:     Source{Repo: "corygyarmathy/skills", URL: "https://example.invalid/skills.git", Prefix: "skills", Rev: "abc"},
		VendorPath: ".agents/skills",
		LinkPath:   ".claude/skills",
		Skills:     map[string]Skill{"tdd": {Path: "skills/tdd", Hash: hash}},
	}
}

func TestVerify(t *testing.T) {
	tests := []struct {
		name      string
		break_    func(t *testing.T, root string, lock *Lock)
		wantDrift bool
		wantLine  string
	}{
		{
			name:     "agreement",
			break_:   func(*testing.T, string, *Lock) {},
			wantLine: "ok        tdd",
		},
		{
			name: "a vendored copy edited in place",
			break_: func(t *testing.T, root string, _ *Lock) {
				p := filepath.Join(root, ".agents/skills/tdd/SKILL.md")
				if err := os.WriteFile(p, []byte("edited\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			wantDrift: true,
			wantLine:  "drifted   tdd",
		},
		{
			name: "a locked skill that was never vendored",
			break_: func(t *testing.T, root string, _ *Lock) {
				if err := os.RemoveAll(filepath.Join(root, ".agents/skills/tdd")); err != nil {
					t.Fatal(err)
				}
			},
			wantDrift: true,
			wantLine:  "missing   tdd",
		},
		{
			name: "a vendored skill the lock does not know about",
			break_: func(t *testing.T, root string, _ *Lock) {
				writeSkill(t, filepath.Join(root, ".agents/skills/bogus"), map[string]string{"SKILL.md": "x\n"})
			},
			wantDrift: true,
			wantLine:  "unlocked  bogus",
		},
		{
			name: "a skill the harness cannot see",
			break_: func(t *testing.T, root string, _ *Lock) {
				if err := os.Remove(filepath.Join(root, ".claude/skills/tdd")); err != nil {
					t.Fatal(err)
				}
			},
			wantDrift: true,
			wantLine:  "unlinked  tdd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, lock := consumer(t)
			tt.break_(t, root, lock)

			var out bytes.Buffer
			err := verify(root, lock, &out)

			if got := errors.Is(err, errDrift); got != tt.wantDrift {
				t.Errorf("drift = %v (err %v), want %v", got, err, tt.wantDrift)
			}
			if !strings.Contains(out.String(), tt.wantLine) {
				t.Errorf("output missing %q:\n%s", tt.wantLine, out.String())
			}
		})
	}
}

func TestReadLock_RejectsAnIncompleteLock(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, lockName), []byte(`{"version":2}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readLock(root); err == nil {
		t.Fatal("readLock accepted a lock with no source.url or vendorPath")
	}
}

func TestFindRoot_WalksUp(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, lockName), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	deep := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(deep)

	got, err := findRoot()
	if err != nil {
		t.Fatal(err)
	}
	// t.TempDir can hand back a symlinked path (/tmp -> /private/tmp), so
	// compare what the filesystem resolves rather than the strings.
	want, _ := filepath.EvalSymlinks(root)
	gotResolved, _ := filepath.EvalSymlinks(got)
	if gotResolved != want {
		t.Errorf("findRoot() = %q, want %q", gotResolved, want)
	}
}
