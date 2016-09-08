package main

import (
	"archive/zip"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// CatJSONError An error for failing to parse Cat event JSON files.
type CatJSONError struct {
	error
}

// CatJSONHeaderError An error for failing to parse Cat event JSON header.
type CatJSONHeaderError struct {
	error
}

// CatJSONVersionError Inidicates if the version is not supported.
type CatJSONVersionError struct {
	error
}

func isSupportedEventVersion(h *CatEventHeader) bool {
	switch h.EventJSONVersion {
	case "1.0":
		return true
	}

	return false
}

// UnzipEvent Unzips a catcierge event ZIP file.
func UnzipEvent(src, dest string) (*CatEventHeader, *CatEventDataV1, error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return nil, nil, err
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

	var pathPrefix string
	var eventName string
	var header CatEventHeader
	var data CatEventDataV1

	// TODO: Add known versions parsing
	// TODO: Verify all files referenced in the json file exists at the correct paths.
	// TODO: Based on content-type version  give error if what we try to parse does not match
	// TODO: If content-type is zip, assume it is the latest version

	// Find the event JSON and set the prefix based on that.
	// (Because some ZIP files have full paths in the zip)
	for _, f := range r.File {
		if filepath.Ext(f.Name) == ".json" {
			// Get the filename without extension.
			_, eventName = filepath.Split(strings.TrimSuffix(f.Name, filepath.Ext(f.Name)))
			pathPrefix = filepath.Dir(f.Name)

			log.Printf("Event ID: %s\n", eventName)
			log.Printf("Path prefix: %s\n", pathPrefix)

			// Decode the JSON.
			rc, err := f.Open()
			if err != nil {
				return nil, nil, err
			}
			defer rc.Close()

			// Start by decoding the header so we can get the version info.
			err = json.NewDecoder(rc).Decode(&header)
			if err != nil {
				log.Printf("Failed to decode event header: %s", err)
				return nil, nil, CatJSONHeaderError{err}
			}

			if !isSupportedEventVersion(&header) {
				log.Printf("Unsupported version for event %s: %s", header.ID, header.Version)
				return &header, nil, CatJSONVersionError{err}
			}

			// We can't seek in a zip file so we re-open it.
			rc.Close()
			rc, err = f.Open()
			if err != nil {
				return nil, nil, err
			}
			defer rc.Close()

			// The version is supported to parse it.
			// TODO: Add a map of parsers for different versions.
			err = json.NewDecoder(rc).Decode(&data)
			if err != nil {
				log.Printf("Failed to decode JSON (v%s) for event %s: %s", header.EventJSONVersion, header.ID, err)
				return &header, nil, CatJSONError{err}
			}
			break
		}
	}

	// Unpack the files.
	for _, f := range r.File {
		f.Name = strings.TrimPrefix(f.Name, pathPrefix)
		err := extractAndWriteFile(f)
		if err != nil {
			return nil, nil, err
		}
	}

	return &header, &data, nil
}
