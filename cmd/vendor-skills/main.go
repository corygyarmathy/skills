// Command vendor-skills maintains a repository's vendored copy of the skills
// in github.com/corygyarmathy/skills.
//
// Harnesses discover skills from the directory they are launched in, so a
// consumer keeps a committed copy rather than pointing at this repository. This
// tool is what keeps that copy honest: skills-lock.json pins the commit the
// copy came from and hashes what was written, and `verify` reports any
// disagreement.
//
// It is run straight from the module, so a consumer carries no copy of it:
//
//	go run github.com/corygyarmathy/skills/cmd/vendor-skills@latest verify
//	go run github.com/corygyarmathy/skills/cmd/vendor-skills@latest pull
//	go run github.com/corygyarmathy/skills/cmd/vendor-skills@latest pull --rev master
//	go run github.com/corygyarmathy/skills/cmd/vendor-skills@latest pull --add tdd
//
// The tool is deliberately not pinned by the lock file. What must be
// reproducible is the vendored content, and that is pinned by commit and
// checked by hash, so a tool that wrote the wrong thing is caught by the next
// verify rather than trusted.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Lock is skills-lock.json: where the skills came from, where they were put,
// and what was written.
type Lock struct {
	Version    int              `json:"version"`
	Source     Source           `json:"source"`
	VendorPath string           `json:"vendorPath"`
	LinkPath   string           `json:"linkPath"`
	Skills     map[string]Skill `json:"skills"`
}

// Source identifies the repository and commit a vendored copy came from.
type Source struct {
	Repo   string `json:"repo"`
	URL    string `json:"url"`
	Prefix string `json:"prefix"`
	Rev    string `json:"rev"`
}

// Skill is one vendored skill: where it lives upstream, and the hash of what
// landed here.
type Skill struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

const lockName = "skills-lock.json"

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "vendor-skills:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	cmd := "verify"
	if len(args) > 0 {
		cmd, args = args[0], args[1:]
	}

	root, err := findRoot()
	if err != nil {
		return err
	}
	lock, err := readLock(root)
	if err != nil {
		return err
	}

	switch cmd {
	case "verify":
		return verify(root, lock, out)
	case "pull":
		return pull(root, lock, args, out)
	default:
		return fmt.Errorf("usage: vendor-skills [verify | pull [--rev REF] [--add NAME]]")
	}
}

// findRoot walks up from the working directory looking for the lock file,
// rather than asking git, so that the tool works the same in a worktree, a
// submodule or a plain directory.
func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, lockName)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no %s in this directory or any parent", lockName)
		}
		dir = parent
	}
}

func readLock(root string) (*Lock, error) {
	b, err := os.ReadFile(filepath.Join(root, lockName))
	if err != nil {
		return nil, err
	}
	var lock Lock
	if err := json.Unmarshal(b, &lock); err != nil {
		return nil, fmt.Errorf("%s: %w", lockName, err)
	}
	if lock.Source.URL == "" || lock.VendorPath == "" {
		return nil, fmt.Errorf("%s: needs source.url and vendorPath", lockName)
	}
	return &lock, nil
}

func writeLock(root string, lock *Lock) error {
	b, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, lockName), append(b, '\n'), 0o644)
}

// hashSkill hashes a skill's whole directory, not just its SKILL.md: a
// reference file beside the skill body is as load-bearing as the body, and a
// change to one of them is exactly the drift worth catching.
//
// The digest is over "<relative path> <file digest>\n" lines in sorted order,
// so it is stable across filesystems and reproducible outside this tool.
func hashSkill(dir string) (string, error) {
	var lines []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(b)
		lines = append(lines, fmt.Sprintf("%s %s\n", filepath.ToSlash(rel), hex.EncodeToString(sum[:])))
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(lines)
	h := sha256.New()
	for _, l := range lines {
		io.WriteString(h, l)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// errDrift reports that the vendored copies and the lock disagree. It is an
// error so that the exit status is usable from CI without parsing output.
var errDrift = errors.New("vendored skills do not match the lock")

func verify(root string, lock *Lock, out io.Writer) error {
	drift := false
	for _, name := range sortedNames(lock.Skills) {
		dir := filepath.Join(root, lock.VendorPath, name)
		switch have, err := hashSkill(dir); {
		case errors.Is(err, fs.ErrNotExist):
			fmt.Fprintf(out, "missing   %s (locked, not vendored)\n", name)
			drift = true
		case err != nil:
			return err
		case have != lock.Skills[name].Hash:
			fmt.Fprintf(out, "drifted   %s (vendored copy differs from the lock)\n", name)
			drift = true
		default:
			fmt.Fprintf(out, "ok        %s\n", name)
		}
		if lock.LinkPath != "" {
			link := filepath.Join(root, lock.LinkPath, name)
			if _, err := os.Stat(link); err != nil {
				fmt.Fprintf(out, "unlinked  %s (no %s/%s)\n", name, lock.LinkPath, name)
				drift = true
			}
		}
	}

	// A skill vendored but absent from the lock is drift in the other
	// direction, and the direction that accumulates quietly.
	entries, err := os.ReadDir(filepath.Join(root, lock.VendorPath))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, ok := lock.Skills[e.Name()]; !ok {
			fmt.Fprintf(out, "unlocked  %s (vendored, not in the lock)\n", e.Name())
			drift = true
		}
	}

	if drift {
		return errDrift
	}
	return nil
}

func pull(root string, lock *Lock, args []string, out io.Writer) error {
	fs := flag.NewFlagSet("pull", flag.ContinueOnError)
	rev := fs.String("rev", lock.Source.Rev, "commit, branch or tag to vendor from")
	add := fs.String("add", "", "skill to add before vendoring")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if rest := fs.Args(); len(rest) > 0 {
		return fmt.Errorf("unexpected argument %q", rest[0])
	}

	checkout, err := os.MkdirTemp("", "vendor-skills-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(checkout)

	if err := git("clone", "--quiet", lock.Source.URL, checkout); err != nil {
		return err
	}
	if err := git("-C", checkout, "checkout", "--quiet", *rev); err != nil {
		return fmt.Errorf("checkout %s: %w", *rev, err)
	}
	resolved, err := gitOutput("-C", checkout, "rev-parse", "HEAD")
	if err != nil {
		return err
	}

	names := sortedNames(lock.Skills)
	if *add != "" {
		if _, ok := lock.Skills[*add]; !ok {
			lock.Skills[*add] = Skill{Path: filepath.ToSlash(filepath.Join(lock.Source.Prefix, *add))}
			names = sortedNames(lock.Skills)
		}
	}

	for _, name := range names {
		src := filepath.Join(checkout, lock.Source.Prefix, name)
		if _, err := os.Stat(src); err != nil {
			return fmt.Errorf("no skill %q at %s", name, resolved)
		}
		dst := filepath.Join(root, lock.VendorPath, name)
		if err := os.RemoveAll(dst); err != nil {
			return err
		}
		if err := copyTree(src, dst); err != nil {
			return err
		}
		hash, err := hashSkill(dst)
		if err != nil {
			return err
		}
		lock.Skills[name] = Skill{
			Path: filepath.ToSlash(filepath.Join(lock.Source.Prefix, name)),
			Hash: hash,
		}
		fmt.Fprintf(out, "vendored  %s\n", name)
	}

	if err := relink(root, lock, names); err != nil {
		return err
	}

	lock.Source.Rev = resolved
	if err := writeLock(root, lock); err != nil {
		return err
	}
	fmt.Fprintf(out, "locked    %s at %s\n", lock.Source.Repo, resolved)
	return nil
}

// relink points the harness directory at the vendored tree. A symlinked
// directory *of* skills is not followed, which is why each skill is linked
// individually and the vendored tree stays the one copy under version control.
func relink(root string, lock *Lock, names []string) error {
	if lock.LinkPath == "" {
		return nil
	}
	linkDir := filepath.Join(root, lock.LinkPath)
	if err := os.MkdirAll(linkDir, 0o755); err != nil {
		return err
	}
	rel, err := filepath.Rel(linkDir, filepath.Join(root, lock.VendorPath))
	if err != nil {
		return err
	}
	for _, name := range names {
		link := filepath.Join(linkDir, name)
		if err := os.RemoveAll(link); err != nil {
			return err
		}
		if err := os.Symlink(filepath.Join(rel, name), link); err != nil {
			return err
		}
	}
	return nil
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		// The mode is carried across because at least one skill ships an
		// executable template script, and a template nobody can run is not one.
		return os.WriteFile(target, b, info.Mode().Perm())
	})
}

func git(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitOutput(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	return strings.TrimSpace(string(out)), err
}

func sortedNames(skills map[string]Skill) []string {
	names := make([]string, 0, len(skills))
	for name := range skills {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
