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
	"strconv"
	"strings"
	"time"
)

const manifestURL = "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"

type manifest struct {
	Versions []struct{ ID, URL string } `json:"versions"`
}
type metadata struct {
	Downloads struct {
		Server         artifact `json:"server"`
		Client         artifact `json:"client"`
		ServerMappings artifact `json:"server_mappings"`
	} `json:"downloads"`
}
type artifact struct {
	SHA1, URL string
	Size      int64
}

func main() {
	version := flag.String("version", "26.2", "Minecraft version (ignored when -versions is set)")
	versions := flag.String("versions", "", "comma-separated Minecraft versions to generate")
	work := flag.String("work", "", "working directory (default: tools/mcgen/.work)")
	output := flag.String("output", "", "version-data output directory (default: internal/data/versions)")
	flag.Parse()
	root, err := findRoot()
	fatal(err)
	if *work == "" {
		*work = filepath.Join(root, "tools", "mcgen", ".work")
	}
	if *output == "" {
		*output = filepath.Join(root, "internal", "data", "versions")
	}
	fatal(os.MkdirAll(*work, 0o755))
	fatal(os.MkdirAll(*output, 0o755))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	client := &http.Client{Timeout: 5 * time.Minute}
	var mf manifest
	fatal(fetchJSON(ctx, client, manifestURL, &mf))
	targets := []string{*version}
	if *versions != "" {
		targets = splitVersions(*versions)
	}
	for _, target := range targets {
		fatal(generateVersion(ctx, client, mf, root, *work, *output, target))
	}
}

func splitVersions(raw string) []string {
	var versions []string
	seen := map[string]struct{}{}
	for _, version := range strings.Split(raw, ",") {
		version = strings.TrimSpace(version)
		if version == "" {
			continue
		}
		if _, ok := seen[version]; !ok {
			versions = append(versions, version)
			seen[version] = struct{}{}
		}
	}
	if len(versions) == 0 {
		fatal(errors.New("-versions did not contain a version"))
	}
	return versions
}

func generateVersion(ctx context.Context, client *http.Client, mf manifest, root, work, output, gameVersion string) error {
	metaURL := ""
	for _, v := range mf.Versions {
		if v.ID == gameVersion {
			metaURL = v.URL
			break
		}
	}
	if metaURL == "" {
		return fmt.Errorf("version %s not found", gameVersion)
	}
	var md metadata
	if err := fetchJSON(ctx, client, metaURL, &md); err != nil {
		return err
	}
	versionWork := filepath.Join(work, safeDir(gameVersion))
	if err := os.MkdirAll(versionWork, 0o755); err != nil {
		return err
	}
	server := filepath.Join(versionWork, "server-"+gameVersion+".jar")
	clientJar := filepath.Join(versionWork, "client-"+gameVersion+".jar")
	serverMappings := filepath.Join(versionWork, "server-"+gameVersion+".txt")
	if err := download(ctx, client, md.Downloads.Server, server); err != nil {
		return err
	}
	if err := download(ctx, client, md.Downloads.Client, clientJar); err != nil {
		return err
	}
	if md.Downloads.ServerMappings.URL != "" {
		if err := download(ctx, client, md.Downloads.ServerMappings, serverMappings); err != nil {
			return err
		}
	}
	reports := filepath.Join(versionWork, "reports")
	if err := os.MkdirAll(reports, 0o755); err != nil {
		return err
	}
	if err := run(ctx, reports, "java", "-DbundlerMainClass=net.minecraft.data.Main", "-jar", server, "--server", "--reports"); err != nil {
		return err
	}
	extracted := filepath.Join(versionWork, "extracted")
	if err := os.RemoveAll(extracted); err != nil {
		return err
	}
	if err := os.MkdirAll(extracted, 0o755); err != nil {
		return err
	}
	if err := extract(server, extracted, func(n string) bool {
		return strings.HasPrefix(n, "META-INF/versions/") || strings.HasPrefix(n, "META-INF/libraries/")
	}); err != nil {
		return err
	}
	inner, err := firstMatch(filepath.Join(extracted, "META-INF", "versions"), "server-*.jar")
	if err != nil {
		return err
	}
	if gameVersion == "1.21.11" {
		unsigned := inner + ".unsigned.jar"
		if err := stripJarSignatures(inner, unsigned); err != nil {
			return err
		}
		inner = unsigned
	}
	libs, err := allMatches(filepath.Join(extracted, "META-INF", "libraries"), "*.jar")
	if err != nil {
		return err
	}
	cp := strings.Join(append([]string{inner}, libs...), string(os.PathListSeparator))
	classes := filepath.Join(versionWork, "classes")
	if err := os.MkdirAll(classes, 0o755); err != nil {
		return err
	}
	dumpSource := filepath.Join(root, "tools", "mcgen", "Dump.java")
	if gameVersion == "1.21.11" {
		dumpSource = filepath.Join(versionWork, "Dump.java")
		if err := remapDumper(filepath.Join(root, "tools", "mcgen", "Dump.java"), serverMappings, dumpSource); err != nil {
			return fmt.Errorf("remap 1.21.11 dumper: %w", err)
		}
	}
	if err := run(ctx, root, "javac", "-cp", cp, "-d", classes, dumpSource); err != nil {
		return err
	}
	dump := filepath.Join(versionWork, "mcdump")
	if err := os.MkdirAll(dump, 0o755); err != nil {
		return err
	}
	if err := run(ctx, root, "java", "-cp", classes+string(os.PathListSeparator)+cp, "Dump", dump); err != nil {
		return err
	}
	assets := filepath.Join(versionWork, "mcsrc", "current")
	if err := os.RemoveAll(assets); err != nil {
		return err
	}
	if err := os.MkdirAll(assets, 0o755); err != nil {
		return err
	}
	if err := extract(inner, assets, func(n string) bool { return strings.HasPrefix(n, "data/minecraft/") }); err != nil {
		return err
	}
	if err := extract(clientJar, assets, func(n string) bool { return n == "assets/minecraft/lang/en_us.json" }); err != nil {
		return err
	}
	if err := copyFile(filepath.Join(assets, "assets", "minecraft", "lang", "en_us.json"), filepath.Join(assets, "en_us.json")); err != nil {
		return err
	}
	input := filepath.Join(versionWork, "input")
	generatedReports := filepath.Join(reports, "generated", "reports")
	for _, name := range []string{"blocks.json", "registries.json", "packets.json"} {
		if err := copyFile(filepath.Join(generatedReports, name), filepath.Join(input, name)); err != nil {
			return err
		}
	}
	itemsDir := filepath.Join(generatedReports, "minecraft", "components", "item")
	if _, err := os.Stat(itemsDir); err == nil {
		if err := copyTree(itemsDir, filepath.Join(input, "items")); err != nil {
			return err
		}
	} else {
		if err := copyFile(filepath.Join(generatedReports, "items.json"), filepath.Join(input, "items.json")); err != nil {
			return err
		}
	}
	if err := copyTree(dump, filepath.Join(input, "mcdump")); err != nil {
		return err
	}
	if err := copyFile(filepath.Join(assets, "en_us.json"), filepath.Join(input, "en_us.json")); err != nil {
		return err
	}
	for _, name := range []string{"component_metadata.include.json", "entity_metadata.include.json"} {
		if err := copyFile(filepath.Join(root, "internal", "data", "generate", name), filepath.Join(input, name)); err != nil {
			return err
		}
	}
	destination := filepath.Join(output, safeDir(gameVersion))
	if err := os.RemoveAll(destination); err != nil {
		return err
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return err
	}
	importRoot := "github.com/zeozeozeo/minego/internal/data/versions/" + safeDir(gameVersion)
	if err := copyRuntimeSources(filepath.Join(root, "internal", "data"), destination, importRoot); err != nil {
		return err
	}
	if err := adaptCopiedPacketSources(destination, gameVersion); err != nil {
		return err
	}
	protocol, ok := knownProtocol(gameVersion)
	if !ok {
		return fmt.Errorf("no verified protocol number configured for Minecraft %s", gameVersion)
	}
	if err := run(ctx, root, "go", "run", "./internal/data/generate", "-input", input, "-decompiled", assets, "-output", destination, "-import-root", importRoot, "-game-version", gameVersion, "-protocol", strconv.Itoa(int(protocol))); err != nil {
		return err
	}
	if err := run(ctx, root, "go", "run", "./internal/data/packets/generate.go", "-packet-ids-import", importRoot+"/packet_ids", filepath.Join(destination, "packets")); err != nil {
		return err
	}
	if err := run(ctx, root, "gofmt", "-w", destination); err != nil {
		return err
	}
	if err := writeAdapterReport(filepath.Join(destination, "PACKET_ADAPTERS.md"), gameVersion); err != nil {
		return err
	}
	fmt.Printf("generated isolated Minecraft %s data in %s\n", gameVersion, destination)
	return nil
}

func safeDir(gameVersion string) string {
	return "v" + strings.NewReplacer(".", "_", "-", "_").Replace(gameVersion)
}

func knownProtocol(gameVersion string) (int32, bool) {
	switch gameVersion {
	case "1.21.11":
		return 774, true
	case "26.1":
		return 775, true
	case "26.2":
		return 776, true
	default:
		return 0, false
	}
}

func writeAdapterReport(path, gameVersion string) error {
	return os.WriteFile(path, []byte("# Packet adapter status for Minecraft "+gameVersion+"\n\n"+
		"This data directory is generated independently. Packet structs are deliberately not copied from another release: add them in a version-owned Go package after reviewing the generated packet IDs and wire schemas. The normalized handshake/status/login packet-ID map may be added to `versions`, but configuration and play adapters require golden encode/decode coverage before a pack is registered in `versions.Supported`.\n"), 0o644)
}

// copyRuntimeSources carries the handwritten support code beside generated
// tables. Generated packages are not useful in isolation without these types
// (for example block-state lookup and item-component codecs). Imports between
// data packages are rewritten to stay inside this release's data root.
func copyRuntimeSources(source, destination, importRoot string) error {
	for _, dir := range []string{"blocks", "entities", "hitboxes", "hitboxes/blocks", "hitboxes/entities", "items", "lang", "packet_ids", "packets", "registries"} {
		from := filepath.Join(source, filepath.FromSlash(dir))
		entries, err := os.ReadDir(from)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_gen.go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(from, name))
			if err != nil {
				return err
			}
			data = []byte(strings.ReplaceAll(string(data), "github.com/zeozeozeo/minego/internal/data", importRoot))
			to := filepath.Join(destination, filepath.FromSlash(dir), name)
			if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(to, data, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

// adaptCopiedPacketSources records schema/name changes that have been reviewed
// against a release's official packets report. Keeping these transformations in
// the generator makes regenerated adapters deterministic and reviewable.
func adaptCopiedPacketSources(destination, gameVersion string) error {
	if gameVersion == "1.21.11" {
		path := filepath.Join(destination, "items", "item_display.go")
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var kept []string
		for _, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, "ComponentAdditionalTradeCost") || strings.Contains(line, "ComponentDye") || strings.Contains(line, "ComponentPigSoundVariant") || strings.Contains(line, "ComponentCowSoundVariant") || strings.Contains(line, "ComponentChickenSoundVariant") || strings.Contains(line, "ComponentCatSoundVariant") {
				continue
			}
			kept = append(kept, line)
		}
		return os.WriteFile(path, []byte(strings.Join(kept, "\n")), 0o644)
	}
	if gameVersion != "26.1" {
		return nil
	}
	path := filepath.Join(destination, "packets", "c2s_play.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	data = []byte(strings.ReplaceAll(string(data), "C2SSpectatorAction", "C2SSpectateEntity"))
	return os.WriteFile(path, data, 0o644)
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

func stripJarSignatures(source, destination string) error {
	r, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer r.Close()
	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	w := zip.NewWriter(out)
	for _, entry := range r.File {
		name := strings.ToUpper(entry.Name)
		if strings.HasPrefix(name, "META-INF/") && (strings.HasSuffix(name, ".SF") || strings.HasSuffix(name, ".RSA") || strings.HasSuffix(name, ".DSA")) {
			continue
		}
		header := entry.FileHeader
		writer, err := w.CreateHeader(&header)
		if err != nil {
			out.Close()
			return err
		}
		in, err := entry.Open()
		if err != nil {
			out.Close()
			return err
		}
		_, err = io.Copy(writer, in)
		in.Close()
		if err != nil {
			out.Close()
			return err
		}
	}
	if err := w.Close(); err != nil {
		out.Close()
		return err
	}
	return out.Close()
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
