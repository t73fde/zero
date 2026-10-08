# oso

Package `oso` provides safe, atomic file writing for Go.

Content is written to a temporary file in the destination directory.
A successful `Close` syncs and closes the temporary file, renames it to the
destination, and syncs the directory. On any error, or if `Close` is never
reached, the temporary file is removed and the previous destination content
stays untouched.

## Usage

```go
func writeData(filename string, data []byte) error {
	f, err := oso.SafeWrite(filename)
	if err != nil {
		return err
	}
	defer f.RollbackIfNeeded()

	if _, err = f.Write(data); err != nil {
		return err
	}
	return f.Close()
}
```

## Limitations

- The destination is always replaced, never modified in place.
- The new file has mode 0600 (see `os.CreateTemp`), regardless of the
  permissions of an existing destination. Ownership, extended attributes, and
  hard links are not preserved. Set other permissions after `Close`, e.g. with
  `os.Chmod`.
- A symbolic link as destination is replaced, not followed.
- Atomicity of the rename depends on the file system.
- A `File` must not be used concurrently by multiple goroutines.

## Name

The package is named after the manufacturer of some safes owned by Scrooge McDuck.

Roughly based on [kjk/common/atomicfile](https://github.com/kjk/common/blob/main/atomicfile).