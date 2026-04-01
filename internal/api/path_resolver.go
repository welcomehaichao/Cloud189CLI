package api

import (
	"fmt"
	"path"
	"strings"
	"sync"
)

type PathResolver struct {
	client     *Client
	cache      map[string]string
	cacheMutex sync.RWMutex
	isFamily   bool
}

func NewPathResolver(client *Client, isFamily bool) *PathResolver {
	return &PathResolver{
		client:   client,
		cache:    make(map[string]string),
		isFamily: isFamily,
	}
}

func (pr *PathResolver) ResolvePath(pathStr string) (string, error) {
	pathStr = strings.TrimSpace(pathStr)

	if pathStr == "" || pathStr == "/" {
		if pr.isFamily {
			return "", nil
		}
		return "-11", nil
	}

	pr.cacheMutex.RLock()
	if folderId, exists := pr.cache[pathStr]; exists {
		pr.cacheMutex.RUnlock()
		return folderId, nil
	}
	pr.cacheMutex.RUnlock()

	if !strings.HasPrefix(pathStr, "/") {
		rootId := "-11"
		if pr.isFamily {
			rootId = ""
		}
		return pr.resolveRelativePath(rootId, pathStr)
	}

	return pr.resolveAbsolutePath(pathStr)
}

func (pr *PathResolver) resolveAbsolutePath(absPath string) (string, error) {
	parts := strings.Split(strings.Trim(absPath, "/"), "/")

	currentId := "-11"
	if pr.isFamily {
		currentId = ""
	}

	for i, part := range parts {
		if part == "" {
			continue
		}

		files, err := pr.client.ListFiles(currentId, 1, 1000, "filename", "asc", pr.isFamily)
		if err != nil {
			return "", fmt.Errorf("failed to list files at '%s': %w", strings.Join(parts[:i+1], "/"), err)
		}

		found := false
		for _, file := range files {
			if file.Name == part && file.IsDir {
				currentId = file.ID
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("folder '%s' not found at path '%s'", part, strings.Join(parts[:i+1], "/"))
		}
	}

	pr.cacheMutex.Lock()
	pr.cache[absPath] = currentId
	pr.cacheMutex.Unlock()

	return currentId, nil
}

func (pr *PathResolver) resolveRelativePath(baseId string, relPath string) (string, error) {
	parts := strings.Split(relPath, "/")

	currentId := baseId

	for i, part := range parts {
		if part == "" || part == "." {
			continue
		}

		if part == ".." {
			return "", fmt.Errorf("parent directory navigation (..) not supported")
		}

		files, err := pr.client.ListFiles(currentId, 1, 1000, "filename", "asc", pr.isFamily)
		if err != nil {
			return "", fmt.Errorf("failed to list files: %w", err)
		}

		found := false
		for _, file := range files {
			if file.Name == part && file.IsDir {
				currentId = file.ID
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("folder '%s' not found", strings.Join(parts[:i+1], "/"))
		}
	}

	return currentId, nil
}

func (pr *PathResolver) GetParentPath(filePath string) (string, string) {
	dir := path.Dir(filePath)
	base := path.Base(filePath)

	if dir == "." {
		if pr.isFamily {
			return "", base
		}
		return "-11", base
	}

	return dir, base
}

func (pr *PathResolver) ClearCache() {
	pr.cacheMutex.Lock()
	defer pr.cacheMutex.Unlock()

	pr.cache = make(map[string]string)
}

func (pr *PathResolver) ResolvePathWithCreate(pathStr string, createIfNotExists bool) (string, error) {
	folderId, err := pr.ResolvePath(pathStr)
	if err == nil {
		return folderId, nil
	}

	if !createIfNotExists {
		return "", err
	}

	return pr.createPath(pathStr)
}

func (pr *PathResolver) createPath(pathStr string) (string, error) {
	pathStr = strings.TrimSpace(pathStr)

	if pathStr == "" || pathStr == "/" {
		if pr.isFamily {
			return "", nil
		}
		return "-11", nil
	}

	var parts []string
	var currentId string

	if strings.HasPrefix(pathStr, "/") {
		parts = strings.Split(strings.Trim(pathStr, "/"), "/")
		currentId = "-11"
		if pr.isFamily {
			currentId = ""
		}
	} else {
		parts = strings.Split(pathStr, "/")
		currentId = "-11"
		if pr.isFamily {
			currentId = ""
		}
	}

	for _, part := range parts {
		if part == "" {
			continue
		}

		files, err := pr.client.ListFiles(currentId, 1, 1000, "filename", "asc", pr.isFamily)
		if err != nil {
			return "", err
		}

		found := false
		for _, file := range files {
			if file.Name == part && file.IsDir {
				currentId = file.ID
				found = true
				break
			}
		}

		if !found {
			folder, err := pr.client.CreateFolder(currentId, part, pr.isFamily)
			if err != nil {
				return "", fmt.Errorf("failed to create folder '%s': %w", part, err)
			}
			currentId = folder.ID
		}
	}

	pr.cacheMutex.Lock()
	pr.cache[pathStr] = currentId
	pr.cacheMutex.Unlock()

	return currentId, nil
}
