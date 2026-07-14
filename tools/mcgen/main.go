// Command mcgen reproducibly regenerates MineGo's version data from official
// Mojang artifacts. It intentionally uses only Go orchestration (plus the Java
// runtime/compiler required to execute Minecraft and Dump.java).
package main

import (
	"archive/zip"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const manifestURL = "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"

type manifest struct {
	Versions []struct{ ID, URL string } `json:"versions"`
}
type metadata struct {
	Downloads struct{ Server, Client artifact } `json:"downloads"`
}
type artifact struct {
	SHA1, URL string
	Size      int64
}

func main() {
	version := flag.String("version", "26.2", "Minecraft version")
	work := flag.String("work", "", "working directory (default: tools/mcgen/.work)")
	flag.Parse()
	root, err := findRoot()
	fatal(err)
	if *work == "" {
		*work = filepath.Join(root, "tools", "mcgen", ".work")
	}
	fatal(os.MkdirAll(*work, 0o755))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	client := &http.Client{Timeout: 5 * time.Minute}
	var mf manifest
	fatal(fetchJSON(ctx, client, manifestURL, &mf))
	metaURL := ""
	for _, v := range mf.Versions {
		if v.ID == *version {
			metaURL = v.URL
			break
		}
	}
	if metaURL == "" {
		fatal(fmt.Errorf("version %s not found", *version))
	}
	var md metadata
	fatal(fetchJSON(ctx, client, metaURL, &md))
	server := filepath.Join(*work, "server-"+*version+".jar")
	clientJar := filepath.Join(*work, "client-"+*version+".jar")
	fatal(download(ctx, client, md.Downloads.Server, server))
	fatal(download(ctx, client, md.Downloads.Client, clientJar))
	reports := filepath.Join(*work, "reports")
	fatal(os.MkdirAll(reports, 0o755))
	fatal(run(ctx, reports, "java", "-DbundlerMainClass=net.minecraft.data.Main", "-jar", server, "--server", "--reports"))
	extracted := filepath.Join(*work, "extracted")
	fatal(os.RemoveAll(extracted))
	fatal(os.MkdirAll(extracted, 0o755))
	fatal(extract(server, extracted, func(n string) bool {
		return strings.HasPrefix(n, "META-INF/versions/") || strings.HasPrefix(n, "META-INF/libraries/")
	}))
	inner, err := firstMatch(filepath.Join(extracted, "META-INF", "versions"), "server-*.jar")
	fatal(err)
	libs, err := allMatches(filepath.Join(extracted, "META-INF", "libraries"), "*.jar")
	fatal(err)
	cp := strings.Join(append([]string{inner}, libs...), string(os.PathListSeparator))
	classes := filepath.Join(*work, "classes")
	fatal(os.MkdirAll(classes, 0o755))
	fatal(run(ctx, root, "javac", "-cp", cp, "-d", classes, filepath.Join(root, "tools", "mcgen", "Dump.java")))
	dump := filepath.Join(*work, "mcdump")
	fatal(os.MkdirAll(dump, 0o755))
	fatal(run(ctx, root, "java", "-cp", classes+string(os.PathListSeparator)+cp, "Dump", dump))
	assets := filepath.Join(*work, "mcsrc", "current")
	fatal(os.RemoveAll(assets))
	fatal(os.MkdirAll(assets, 0o755))
	fatal(extract(inner, assets, func(n string) bool { return strings.HasPrefix(n, "data/minecraft/") }))
	fatal(extract(clientJar, assets, func(n string) bool { return n == "assets/minecraft/lang/en_us.json" }))
	fatal(copyFile(filepath.Join(assets, "assets", "minecraft", "lang", "en_us.json"), filepath.Join(assets, "en_us.json")))
	input := filepath.Join(root, "internal", "data", "generate")
	generatedReports := filepath.Join(reports, "generated", "reports")
	for _, name := range []string{"blocks.json", "registries.json", "packets.json"} {
		fatal(copyFile(filepath.Join(generatedReports, name), filepath.Join(input, name)))
	}
	fatal(copyTree(filepath.Join(generatedReports, "minecraft", "components", "item"), filepath.Join(input, "items")))
	fatal(copyTree(dump, filepath.Join(input, "mcdump")))
	fatal(copyFile(filepath.Join(assets, "en_us.json"), filepath.Join(input, "en_us.json")))
	fatal(run(ctx, root, "go", "run", "./internal/data/generate", input, assets))
	fatal(run(ctx, root, "go", "run", "./internal/data/packets/generate.go", filepath.Join(root, "internal", "data", "packets")))
	fatal(run(ctx, root, "go", "fmt", "./internal/data/..."))
	fmt.Printf("generated Minecraft %s data successfully\n", *version)
}

func findRoot() (string, error) {
	d, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, e := os.Stat(filepath.Join(d, "go.mod")); e == nil {
			return d, nil
		}
		p := filepath.Dir(d)
		if p == d {
			return "", errors.New("go.mod not found")
		}
		d = p
	}
}
func fetchJSON(ctx context.Context, c *http.Client, url string, out any) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("GET %s: %s", url, res.Status)
	}
	return json.NewDecoder(res.Body).Decode(out)
}
func download(ctx context.Context, c *http.Client, a artifact, path string) error {
	if data, err := os.ReadFile(path); err == nil && hash(data) == a.SHA1 {
		return nil
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, a.URL, nil)
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("GET %s: %s", a.URL, res.Status)
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	h := sha1.New()
	_, copyErr := io.Copy(io.MultiWriter(f, h), res.Body)
	closeErr := f.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != a.SHA1 {
		_ = os.Remove(tmp)
		return fmt.Errorf("SHA-1 mismatch for %s: %s != %s", path, got, a.SHA1)
	}
	return os.Rename(tmp, path)
}
func hash(b []byte) string { h := sha1.Sum(b); return hex.EncodeToString(h[:]) }
func run(ctx context.Context, dir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if runtime.GOOS == "windows" && name == "gofmt" {
		cmd = exec.CommandContext(ctx, "gofmt.exe", args...)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}
func extract(path, dst string, keep func(string) bool) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		n := filepath.ToSlash(f.Name)
		if !keep(n) || f.FileInfo().IsDir() {
			continue
		}
		out := filepath.Join(dst, filepath.FromSlash(n))
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		in, err := f.Open()
		if err != nil {
			return err
		}
		o, err := os.Create(out)
		if err != nil {
			in.Close()
			return err
		}
		_, e := io.Copy(o, in)
		in.Close()
		o.Close()
		if e != nil {
			return e
		}
	}
	return nil
}
func firstMatch(root, pattern string) (string, error) {
	all, err := allMatches(root, pattern)
	if err != nil {
		return "", err
	}
	if len(all) == 0 {
		return "", fmt.Errorf("%s not found under %s", pattern, root)
	}
	return all[0], nil
}
func allMatches(root, pattern string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if ok, _ := filepath.Match(pattern, d.Name()); ok {
				out = append(out, path)
			}
		}
		return nil
	})
	return out, err
}
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err = os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, e := io.Copy(out, in)
	if c := out.Close(); e == nil {
		e = c
	}
	return e
}
func copyTree(src, dst string) error {
	_ = os.RemoveAll(dst)
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}
func fatal(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "mcgen:", err)
		os.Exit(1)
	}
}
