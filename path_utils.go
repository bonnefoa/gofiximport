package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// PathWalkFunc is called on each file found by PathWalk
type PathWalkFunc func(ctx interface{}, path string, info os.FileInfo, err error) error

// PathWalk recursively traverse all files, passing ctx as first argument
func PathWalk(ctx interface{}, root string, walkFn PathWalkFunc) error {
	info, err := os.Lstat(root)
	if err != nil {
		return walkFn(ctx, root, nil, err)
	}
	return pathWalk(ctx, root, info, walkFn)
}

// walk recursively descends path, calling walkFn.
func pathWalk(ctx interface{}, path string,
	info os.FileInfo, walkFn PathWalkFunc) error {
	err := walkFn(ctx, path, info, nil)
	if err != nil {
		if info.IsDir() && err == filepath.SkipDir {
			return nil
		}
		return err
	}

	if !info.IsDir() {
		return nil
	}

	list, err := ioutil.ReadDir(path)
	if err != nil {
		return walkFn(ctx, path, info, err)
	}

	for _, fileInfo := range list {
		err = pathWalk(ctx, filepath.Join(path, fileInfo.Name()),
			fileInfo, walkFn)
		if err != nil {
			if !fileInfo.IsDir() || err != filepath.SkipDir {
				return err
			}
		}
	}
	return nil
}

func mostRecentWalk(ctx interface{}, path string, info os.FileInfo, err error) error {
	t := ctx.(*time.Time)
	if info.ModTime().After(*t) {
		*t = info.ModTime()
	}
	return nil
}

func mostRecentModification(paths []string) time.Time {
	loc, _ := time.LoadLocation("UTC")
	t := time.Date(2001, 1, 1, 1, 1, 1, 0, loc)
	for _, path := range gopaths {
		PathWalk(&t, path, mostRecentWalk)
	}
	return t
}
