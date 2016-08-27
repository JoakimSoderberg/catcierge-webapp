package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"log"
	"strings"
	"encoding/json"
)

func UnzipEvent(src, dest string) error {
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

	// Find the event JSON and set the prefix based on that.
	// (Because some ZIP files have full paths in the zip)
	var pathPrefix string
	var eventName string
	var data CatEventDataV1 // TODO: Change to pares header only first

	for _, f := range r.File {
		if filepath.Ext(f.Name) == ".json" {
			// Get the filename without extension.
			_, eventName = filepath.Split(strings.TrimSuffix(f.Name, filepath.Ext(f.Name)))
			pathPrefix = filepath.Dir(f.Name)

			log.Printf("Event name: %s\n", eventName)
			log.Printf("Path prefix: %s\n", pathPrefix)

			// Decode the JSON.
			rc, err := f.Open()
			if err != nil {
				// TODO: Return extended error message we can return in REST API.
				return err
			}

			defer rc.Close()

			// TODO: Fails on unmarshalling the dates
			err = json.NewDecoder(rc).Decode(&data)
			if err != nil {
				log.Printf("Failed to decode event JSON: %s", err)
				return err
			}

			break
		}
	}

	// Unpack the files.
	for _, f := range r.File {
		f.Name = strings.TrimPrefix(f.Name, pathPrefix)
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
