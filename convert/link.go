package convert

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"golang.org/x/text/unicode/norm"
)

type PathDB interface {
	Get(fileId string) (path string, err error)
}

type pathDbImpl struct {
	vault     string
	vaultdict map[string][]string
}

func NewPathDB(vault string) PathDB {
	db := new(pathDbImpl)
	db.vault = vault
	db.vaultdict = make(map[string][]string)
	filepath.Walk(vault, func(path string, info fs.FileInfo, err error) error {
		// if vault was not found, info will be nil
		if info == nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		base := norm.NFC.String(filepath.Base(path))
		db.vaultdict[base] = append(db.vaultdict[base], path)
		return nil
	})
	return db
}

func (f *pathDbImpl) Get(fileId string) (path string, err error) {
	var filename string
	if filepath.Ext(fileId) == "" {
		filename = fileId + ".md"
	} else {
		filename = fileId
	}

	base := filepath.Base(filename)
	paths, ok := f.vaultdict[base]
	if !ok || len(paths) == 0 {
		return "", nil
	}

	bestscore := -1
	bestmatch := ""
	for _, pth := range paths {
		if score := pathMatchScore(pth, filename); score < 0 {
			continue
		} else if bestmatch == "" || score < bestscore || (score == bestscore && strings.Compare(pth, bestmatch) < 0) {
			bestscore = score
			bestmatch = pth
			continue
		}
	}

	if bestscore < 0 {
		return "", nil
	}
	path, err = filepath.Rel(f.vault, bestmatch)
	if err != nil {
		return "", newErrTransformf(ERR_KIND_UNEXPECTED, "filepath.Rel failed: %v", err)
	}
	return filepath.ToSlash(path), nil
}

func pathMatchScore(path string, filename string) int {
	// 書記素クラスタに対応
	path = norm.NFC.String(path)
	filename = norm.NFC.String(filename)

	pp := strings.Split(filepath.ToSlash(path), "/")
	ff := strings.Split(filepath.ToSlash(filename), "/")
	lpp := len(pp)
	lff := len(ff)
	if lpp < lff {
		return -1
	}
	cur := 0
	for ; cur < lff; cur++ {
		if pp[lpp-1-cur] != ff[lff-1-cur] {
			return -1
		}
	}
	return lpp - cur
}

type pathDBWrapperImplReturningNotFoundPathError struct {
	original PathDB
}

func (w *pathDBWrapperImplReturningNotFoundPathError) Get(fileId string) (path string, err error) {
	if w.original == nil {
		panic("original PathDB not set but used")
	}
	path, err = w.original.Get(fileId)
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", newErrTransformf(ERR_KIND_PATH_NOT_FOUND, "failed to resolve ref \"%s\"", fileId)
	}
	return path, nil
}

func WrapForReturningNotFoundPathError(original PathDB) PathDB {
	return &pathDBWrapperImplReturningNotFoundPathError{original: original}
}

func splitDisplayName(fullname string) (identifier string, displayname string) {
	position := strings.Index(fullname, "|")
	if position < 0 {
		return fullname, ""
	} else {
		identifier := strings.Trim(string(fullname[:position]), " \t")
		displayname := strings.TrimLeft(string(fullname[position:]), "|")
		displayname = strings.Trim(displayname, " \t")
		return identifier, displayname
	}
}

func splitFragments(identifier string) (fileId string, fragments []string, err error) {
	strs := strings.Split(identifier, "#")
	if len(strs) == 1 {
		return strs[0], nil, nil
	}
	fileId = strs[0]
	if len(strings.TrimRight(fileId, " \t")) != len(fileId) {
		return "", nil, newErrTransformf(ERR_KIND_INVALID_INTERNAL_LINK_CONTENT, "invalid internal link content: %q", identifier)
	}

	fragments = make([]string, len(strs)-1)
	for id, f := range strs[1:] {
		f := strings.Trim(f, " \t")
		fragments[id] = f
	}
	return fileId, fragments, nil
}

func buildLinkText(displayName string, fileId string, fragments []string) (linktext string) {
	if displayName != "" {
		return displayName
	}

	if fileId != "" {
		linktext = fileId
		for _, f := range fragments {
			linktext += fmt.Sprintf(" > %s", f)
		}
		return linktext
	} else {
		return strings.Join(fragments, " > ")
	}
}
