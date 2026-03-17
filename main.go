// path: dct/main.go

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	var (
		srcDir   string
		name     string
		outRoot  string
		projRoot string
		keepSrc  bool
		quiet    bool
		version  string
	)

	flag.StringVar(&srcDir, "src", "", "Path to project directory (required)")
	flag.StringVar(&name, "name", "", "Project name (defaults to basename of -src)")
	flag.StringVar(&outRoot, "out", "publish", "Publish root directory")
	flag.StringVar(&projRoot, "projects", "projects", "Projects root directory (for moved source)")
	flag.BoolVar(&keepSrc, "keep", false, "Keep source in place (do not move into projects/)")
	flag.BoolVar(&quiet, "q", false, "Quiet zip recursion where applicable")
	flag.StringVar(&version, "version", "dev", "Release version")
	flag.Parse()

	if srcDir == "" {
		fail("missing -src (path to your project folder)")
	}

	absSrc, err := filepath.Abs(srcDir)
	check(err, "resolve src path")

	info, err := os.Stat(absSrc)
	if err != nil || !info.IsDir() {
		fail("source does not exist or is not a directory: %s", absSrc)
	}

	if name == "" {
		name = filepath.Base(absSrc)
	}

	if name == "." || name == "/" {
		fail("invalid project name derived from src: %q", name)
	}

	/* [AI:GPT | 2026-03-16 21:20:00 UTC] */
	// Added gpg requirement for signing support
	ensureTools("tar", "zip", "bzip2", "md5sum", "sha1sum", "gpg")
	/* [End AI:GPT] */

	fmt.Println("Checking directory structure…")

	checkMkDir("projects", 0o755)
	checkMkDir("publish", 0o755)

	pubDir := filepath.Join(outRoot, name)

	hashDir := filepath.Join(pubDir, "hash")
	hashZip := filepath.Join(hashDir, "zip")
	hashGz := filepath.Join(hashDir, "gzip")
	hashSha1 := filepath.Join(hashDir, "sha1")
	hashSha256 := filepath.Join(hashDir, "sha256")
	hashBz2 := filepath.Join(hashDir, "bz2")

	checkMkDir(pubDir, 0o755)
	checkMkDir(hashDir, 0o755)
	checkMkDir(hashZip, 0o755)
	checkMkDir(hashGz, 0o755)
	checkMkDir(hashSha1, 0o755)
	checkMkDir(hashSha256, 0o755)
	checkMkDir(hashBz2, 0o755)

	fmt.Println("Directory structure confirmed.")

	tarPath := name + ".tar"
	tgzPath := name + ".tar.gz"
	zipPath := name + ".zip"
	tbz2Path := name + ".tar.bz2"

	parent := filepath.Dir(absSrc)
	base := filepath.Base(absSrc)

	fmt.Println("Compressing your application…")

	runCmd("tar", "-cf", tarPath, "-C", parent, base)
	runCmd("tar", "-czf", tgzPath, "-C", parent, base)

	runCmd("zip", "-rq", zipPath, base)

	runCmdToFile(tbz2Path, "bzip2", "-z", "-c", "--fast", tarPath)

	fmt.Println("Hashing your application…")

	runHashToFile(filepath.Join(hashZip, "zip-md5.log"), "md5sum", zipPath)
	runHashToFile(filepath.Join(hashGz, "gzip-md5.log"), "md5sum", tgzPath)
	runHashToFile(filepath.Join(hashBz2, "bz2-md5.log"), "md5sum", tbz2Path)

	runHashToFile(filepath.Join(hashSha1, "zip-sha1.log"), "sha1sum", zipPath)
	runHashToFile(filepath.Join(hashSha1, "gzip-sha1.log"), "sha1sum", tgzPath)
	runHashToFile(filepath.Join(hashBz2, "bz2-sha1.log"), "sha1sum", tbz2Path)

	hashes := make(map[string]string)

	hashes[zipPath] = getHash("sha256sum", zipPath)
	hashes[tgzPath] = getHash("sha256sum", tgzPath)
	hashes[tbz2Path] = getHash("sha256sum", tbz2Path)

	/* [AI:GPT | 2026-03-16 21:20:00 UTC] */
	fmt.Println("Signing your application…")

	signFile(zipPath)
	signFile(tgzPath)
	signFile(tbz2Path)
	/* [End AI:GPT] */

	fmt.Println("Publishing your application…")

	moveFile(tgzPath, filepath.Join(pubDir, filepath.Base(tgzPath)))
	moveFile(zipPath, filepath.Join(pubDir, filepath.Base(zipPath)))
	moveFile(tbz2Path, filepath.Join(pubDir, filepath.Base(tbz2Path)))

	/* [AI:GPT | 2026-03-16 21:20:00 UTC] */
	moveFile(zipPath+".asc", filepath.Join(pubDir, filepath.Base(zipPath)+".asc"))
	moveFile(tgzPath+".asc", filepath.Join(pubDir, filepath.Base(tgzPath)+".asc"))
	moveFile(tbz2Path+".asc", filepath.Join(pubDir, filepath.Base(tbz2Path)+".asc"))
	/* [End AI:GPT] */

	/* [AI:GPT | 2026-03-16 21:20:00 UTC] */
	fmt.Println("Generating release manifest…")
	writeManifest(pubDir, name, version, hashes)
	/* [End AI:GPT] */

	_ = os.Remove(tarPath)

	if !keepSrc {
		checkMkDir(projRoot, 0o755)

		dest := filepath.Join(projRoot, name)
		if exists(dest) {
			dest = dest + "-" + time.Now().Format("20060102-150405")
		}

		check(os.Rename(absSrc, dest), "move source to projects/")
	}

	fmt.Println()
	fmt.Println("All done.")
	fmt.Printf("Published: %s\n", pubDir)

	if !keepSrc {
		fmt.Printf("Source moved to: %s\n", projRoot)
	}
}

/* ------------ helpers ------------ */

func check(err error, msg string) {
	if err != nil {
		fail("%s: %v", msg, err)
	}
}

func fail(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "ERR: "+format+"\n", a...)
	os.Exit(1)
}

func checkMkDir(path string, mode os.FileMode) {
	if err := os.MkdirAll(path, mode); err != nil {
		fail("mkdir %s: %v", path, err)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func ensureTools(names ...string) {
	for _, n := range names {
		if _, err := exec.LookPath(n); err != nil {
			fail("required tool not found in PATH: %s", n)
		}
	}
}

func runCmd(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fail("command failed: %s %v", name, args)
	}
}

func runCmdToFile(outPath string, name string, args ...string) {
	out, err := os.Create(outPath)
	check(err, "create "+outPath)
	defer out.Close()

	cmd := exec.Command(name, args...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fail("command failed: %s %v", name, args)
	}
}

func runHashToFile(outPath string, algo string, file string) {
	out, err := os.Create(outPath)
	check(err, "create "+outPath)
	defer out.Close()

	cmd := exec.Command(algo, file)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fail("hash failed: %s %s", algo, file)
	}
}

/* [AI:GPT | 2026-03-16 21:20:00 UTC] */
func getHash(algo, file string) string {
	out, err := exec.Command(algo, file).Output()
	if err != nil {
		fail("failed to hash %s with %s", file, algo)
	}

	parts := strings.Fields(string(out))
	return parts[0]
}
/* [End AI:GPT] */

/* [AI:GPT | 2026-03-16 21:20:00 UTC] */
func signFile(file string) {
	cmd := exec.Command("gpg", "--armor", "--detach-sign", file)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fail("gpg signing failed for %s", file)
	}
}
/* [End AI:GPT] */

/* [AI:GPT | 2026-03-16 21:20:00 UTC] */
func writeManifest(pubDir, name, version string, hashes map[string]string) {

	file := filepath.Join(pubDir, "release.manifest")

	f, err := os.Create(file)
	check(err, "create manifest")
	defer f.Close()

	now := time.Now().UTC().Format(time.RFC3339)

	fmt.Fprintf(f, "project: %s\n", name)
	fmt.Fprintf(f, "version: %s\n", version)
	fmt.Fprintf(f, "built: %s\n\n", now)

	fmt.Fprintf(f, "artifacts:\n\n")

	for fileName, hash := range hashes {

		base := filepath.Base(fileName)

		fmt.Fprintf(f, "- file: %s\n", base)
		fmt.Fprintf(f, "  sha256: %s\n", hash)
		fmt.Fprintf(f, "  signature: %s.asc\n\n", base)
	}
}
/* [End AI:GPT] */

func moveFile(src, dst string) {

	if _, err := os.Stat(src); err != nil {
		fail("cannot move, source missing: %s", src)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		fail("failed to create dest dir for %s: %v", dst, err)
	}

	if err := os.Rename(src, dst); err == nil {
		return
	}

	in, err := os.Open(src)
	if err != nil {
		fail("open src for move: %v", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		fail("create dst for move: %v", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		fail("copy data: %v", err)
	}

	out.Close()
	in.Close()

	if err := os.Remove(src); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: could not remove source after copy: %v\n", err)
	}
}
