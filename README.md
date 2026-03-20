# DCT — The Developer's Compression Tool
![GO](https://img.shields.io/badge/GO1-blue)
![Architecture](https://img.shields.io/badge/Architecture-GO-darkgreen)
![Status](https://img.shields.io/badge/Status-Active%20Development-orange)
![License](https://img.shields.io/badge/License-TBD-lightgrey)
![Sponsored](https://img.shields.io/badge/Sponsored-STN_Labz-blue)

Written in Go.

---

## Usage

```bash
dct -src target
```

Where target is your project directory.

## What It Does

DCT packages your application into multiple distribution formats:

 - .zip
 - .tar.gz
 - .tar.bz2

It then generates cryptographic hashes for verification:

 - MD5
 - SHA1
 - SHA256 (fallback to BLAKE3 if unavailable)

Finally, it signs all distributable artifacts using:
 - GPG (ASCII armored signatures)

**Output**

A complete publish-ready structure:
```bash
publish/<project>/

<project>.zip
<project>.zip.asc

<project>.tar.gz
<project>.tar.gz.asc

<project>.tar.bz2
<project>.tar.bz2.asc

hash/
  zip/
  gzip/
  bz2/
  sha1/
  sha256/
```

## What It Requires

 - Linux environment
 - GPG
 - tar
 - gzip
 - bzip2
 - zip
 - coreutils (md5sum, sha1sum, sha256sum or b3sum)

## Philosophy

DCT is built on a simple principle:

*Build once. Package cleanly. Verify everything. Trust nothing blindly*.

**It provides**:

 - Integrity (hashes)
 - Authenticity (GPG signatures)
 - A complete manifest
 - Consistency (structured output)

No external services. No vendor lock-in. No hidden processes.
