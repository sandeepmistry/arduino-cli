package resources

import (
	"fmt"
	"os"

	paths "github.com/arduino/go-paths-helper"
	"github.com/codeclysm/extract"
)

// Install installs the resource in three steps:
// - the archive is unpacked in a temporary subfolder of tempPath
// - there should be only one root folder in the unpacked content
// - the only root folder is moved/renamed to/as the destination directory
// Note that tempPath and destDir must be on the same filesystem partition
// otherwise the last step will fail.
func (release *DownloadResource) Install(downloadDir, tempPath, destDir *paths.Path) error {
	// Create a temporary folder to extract package
	if err := tempPath.MkdirAll(); err != nil {
		return fmt.Errorf("creating temp dir for extraction: %s", err)
	}
	tempDir, err := tempPath.MkTempDir("package-")
	if err != nil {
		return fmt.Errorf("creating temp dir for extraction: %s", err)
	}
	defer tempDir.RemoveAll()

	// Obtain the archive path and open it
	archivePath, err := release.ArchivePath(downloadDir)
	if err != nil {
		return fmt.Errorf("getting archive path: %s", err)
	}
	file, err := os.Open(archivePath.String())
	if err != nil {
		return fmt.Errorf("opening archive file: %s", err)
	}
	defer file.Close()

	// Extract into temp directory
	if err := extract.Archive(file, tempDir.String(), nil); err != nil {
		return fmt.Errorf("extracting archive: %s", err)
	}

	// Check package content and find package root dir
	root, err := findPackageRoot(tempDir)
	if err != nil {
		return fmt.Errorf("searching package root dir: %s", err)
	}

	// Ensure container dir exists
	destDirParent := destDir.Parent()
	if err := destDirParent.MkdirAll(); err != nil {
		return err
	}
	defer func() {
		if empty, err := IsDirEmpty(destDirParent); err == nil && empty {
			destDirParent.RemoveAll()
		}
	}()

	// Move/rename the extracted root directory in the destination directory
	if err := root.Rename(destDir); err != nil {
		return fmt.Errorf("moving extracted archive to destination dir: %s", err)
	}

	// TODO
	// // Create a package file
	// if err := createPackageFile(destDir); err != nil {
	// 	return err
	// }

	return nil
}

// IsDirEmpty returns true if the directory specified by path is empty.
func IsDirEmpty(path *paths.Path) (bool, error) {
	if files, err := path.ReadDir(); err != nil {
		return false, err
	} else {
		return len(files) == 0, nil
	}
}

func findPackageRoot(parent *paths.Path) (*paths.Path, error) {
	files, err := parent.ReadDir()
	if err != nil {
		return nil, fmt.Errorf("reading package root dir: %s", err)
	}
	var root *paths.Path
	for _, file := range files {
		if isdir, _ := file.IsDir(); !isdir {
			continue
		}
		if root == nil {
			root = file
		} else {
			return nil, fmt.Errorf("no unique root dir in archive, found '%s' and '%s'", root, file)
		}
	}
	return root, nil
}