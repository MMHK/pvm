package commands

import (
	"archive/zip"
	"fmt"
	"hjbdev/pvm/common"
	"hjbdev/pvm/theme"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func Install(conf *common.Config, args []string) {
	if len(args) < 2 {
		theme.Error("You must specify a version to install.")
		return
	}

	desireThreadSafe := true
	if len(args) > 2 {
		if args[2] == "nts" {
			desireThreadSafe = false
		}
	}

	var threadSafeString string
	if desireThreadSafe {
		threadSafeString = "thread safe"
	} else {
		threadSafeString = "non-thread safe"
	}

	if desireThreadSafe {
		theme.Warning("Thread safe version will be installed")
	} else {
		theme.Warning("Non-thread safe version will be installed")
	}

	desiredVersionNumbers := common.GetVersion(args[1], desireThreadSafe, "")

	if desiredVersionNumbers == (common.Version{}) {
		theme.Error("Invalid version specified")
		return
	}

	// Get the desired version from the user input
	desiredMajorVersion := desiredVersionNumbers.Major
	desiredMinorVersion := desiredVersionNumbers.Minor
	desiredPatchVersion := desiredVersionNumbers.Patch

	// perform get request to https://windows.php.net/downloads/releases/archives/
	resp, err := http.Get("https://windows.php.net/downloads/releases/archives/")
	if err != nil {
		log.Fatalln(err)
	}
	// We Read the response body on the line below.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	// Convert the body to type string
	sb := string(body)

	// regex match
	re := regexp.MustCompile(`<A HREF="([a-zA-Z0-9./-]+)">([a-zA-Z0-9./-]+)</A>`)
	matches := re.FindAllStringSubmatch(sb, -1)

	versions := make([]common.Version, 0)

	for _, match := range matches {
		url := match[1]
		name := match[2]

		// check if name starts with "php-devel-pack-"
		if name != "" && len(name) > 15 && name[:15] == "php-devel-pack-" {
			continue
		}
		// check if name starts with "php-debug-pack-"
		if name != "" && len(name) > 15 && name[:15] == "php-debug-pack-" {
			continue
		}
		// check if name starts with "php-test-pack-"
		if name != "" && len(name) > 15 && name[:14] == "php-test-pack-" {
			continue
		}

		// check if name contains "src"
		if name != "" && strings.Contains(name, "src") {
			continue
		}

		// check if name does not end in zip
		if name != "" && !strings.HasSuffix(name, ".zip") {
			continue
		}

		threadSafe := true

		// check if name contains "nts" or "NTS"
		if name != "" && (strings.Contains(name, "nts") || strings.Contains(name, "NTS")) {
			threadSafe = false
		}

		// make sure we only get x64 versions
		if name != "" && !strings.Contains(name, "x64") {
			continue
		}

		// regex match name and push to versions
		versions = append(versions, common.GetVersion(name, threadSafe, url))
	}

	// find desired version
	var desiredVersion common.Version

	if desiredMajorVersion > -1 && desiredMinorVersion > -1 && desiredPatchVersion > -1 {
		desiredVersion = FindExactVersion(versions, desiredMajorVersion, desiredMinorVersion, desiredPatchVersion, desireThreadSafe)
	}

	if desiredMajorVersion > -1 && desiredMinorVersion > -1 && desiredPatchVersion == -1 {
		desiredVersion = FindLatestPatch(versions, desiredMajorVersion, desiredMinorVersion, desireThreadSafe)
	}

	if desiredMajorVersion > -1 && desiredMinorVersion == -1 && desiredPatchVersion == -1 {
		desiredVersion = FindLatestMinor(versions, desiredMajorVersion, desireThreadSafe)
	}

	if desiredVersion == (common.Version{}) {
		theme.Error(fmt.Sprintf("Could not find the desired version: %s %s", args[1], threadSafeString))
		return
	}

	fmt.Printf("Installing PHP %s\n", desiredVersion)


	// check if .pvm/versions folder exists
	if _, err := os.Stat(conf.PVM_VERSIONS_PATH); os.IsNotExist(err) {
		theme.Info(fmt.Sprintf("Creating %s folder", conf.PVM_VERSIONS_PATH))
		os.Mkdir(conf.PVM_VERSIONS_PATH, 0755)
	}

	theme.Info("Downloading")

	// Get the data
	downloadResponse, err := http.Get("https://windows.php.net" + desiredVersion.Url)
	if err != nil {
		log.Fatalln(err)
	}

	defer downloadResponse.Body.Close()

	// zip filename from url
	zipFileName := strings.Split(desiredVersion.Url, "/")[len(strings.Split(desiredVersion.Url, "/"))-1]
    fullZipFilePath := filepath.Join(conf.PVM_VERSIONS_PATH, zipFileName)

	// check if zip already exists
	if _, err := os.Stat(fullZipFilePath); err == nil {
		theme.Error(fmt.Sprintf("PHP %s already exists", desiredVersion))
		return
	}

	// Create the file
	out, err := os.Create(fullZipFilePath)
	if err != nil {
		log.Fatalln(err)
	}

	// Write the body to file
	_, err = io.Copy(out, downloadResponse.Body)

	if err != nil {
		out.Close()
		log.Fatalln(err)
	}

	// Close the file
	out.Close()

	// extract the zip file to a folder
	theme.Info("Unzipping")

	Unzip(fullZipFilePath, strings.Replace(fullZipFilePath, ".zip", "", -1))

	// remove the zip file
	theme.Info("Cleaning up")
	err = os.Remove(fullZipFilePath)
	if err != nil {
		log.Fatalln(err)
	}

	theme.Success(fmt.Sprintf("Finished installing PHP %s", desiredVersion))
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func FindExactVersion(versions []common.Version, major int, minor int, patch int, threadSafe bool) common.Version {
	for _, version := range versions {
		if version.ThreadSafe != threadSafe {
			continue
		}
		if version.Major == major && version.Minor == minor && version.Patch == patch {
			return version
		}
	}

	return common.Version{}
}

func FindLatestPatch(versions []common.Version, major int, minor int, threadSafe bool) common.Version {
	latestPatch := common.Version{}

	for _, version := range versions {
		if version.ThreadSafe != threadSafe {
			continue
		}
		if version.Major == major && version.Minor == minor {
			if latestPatch.Patch == -1 || version.Patch > latestPatch.Patch {
				latestPatch = version
			}
		}
	}

	return latestPatch
}

func FindLatestMinor(versions []common.Version, major int, threadSafe bool) common.Version {
	latestMinor := common.Version{}

	for _, version := range versions {
		if version.ThreadSafe != threadSafe {
			continue
		}
		if version.Major == major {
			if latestMinor.Minor == -1 || version.Minor > latestMinor.Minor {
				if latestMinor.Patch == -1 || version.Patch > latestMinor.Patch {
					latestMinor = version
				}
			}
		}
	}

	return latestMinor
}
