package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/qawatake/obsdconv/convert"
)

const (
	FLAG_SOURCE             = "src"
	FLAG_DESTINATION        = "dst"
	FLAG_TARGET             = "tgt"
	FLAG_REMOVE_TAGS        = "rmtag"
	FLAG_COPY_TAGS          = "cptag"
	FLAG_SYNC_TAGS          = "synctag"
	FLAG_COPY_TITLE         = "title"
	FLAG_COPY_ALIASES       = "alias"
	FLAG_SYNC_TITLE_ALIASES = "synctlal"
	FLAG_CONVERT_LINKS      = "link"
	FLAG_REMOVE_COMMENT     = "cmmt"
	FLAG_PUBLISHABLE        = "pub"
	FLAG_REMOVE_H1          = "rmh1"
	FLAG_REMAP_META_KEYS    = "remapkey"
	FLAG_FILTER             = "filter"
	// FLAG_BASE_URL           = "baseUrl"
	FLAG_REMAP_PATH_PREFIX = "remapPathPrefix"
	FLAG_FORMAT_LINK       = "formatLink"
	FLAG_FORMAT_ANCHOR     = "formatAnchor"
	FLAG_STRICT_REF        = "strictref"
	FLAG_OBSIDIAN_USAGE    = "obs"
	FLAG_STANDARD_USAGE    = "std"
	FLAG_VERSION           = "version"
	FLAG_DEBUG             = "debug"
)

type configuration struct {
	src         string
	dst         string
	tgt         string
	rmtag       bool
	cptag       bool
	synctag     bool
	title       bool
	alias       bool
	synctlal    bool
	link        bool
	cmmt        bool
	publishable bool
	rmH1        bool
	strictref   bool
	remapkey    string
	filter      string
	// baseUrl         string
	remapPathPrefix string
	formatLink      bool
	formatAnchor    string
	obs             bool
	std             bool
	ver             bool
	debug           bool
}

type mainErrKind int

const (
	MAIN_ERR_UNEXPECTED = iota + 1
	MAIN_ERR_KIND_SOURCE_NOT_SET
	MAIN_ERR_KIND_DESTINATION_NOT_SET
	MAIN_ERR_KIND_STRICTREF_NEEDS_LINK
	MAIN_ERR_KIND_INVALID_SOURCE_FORMAT
	MAIN_ERR_KIND_INVALID_DESTINATION_FORMAT
	MAIN_ERR_KIND_TARGET_IS_MARKDOWN_FILE_BUT_DESTINATION_IS_NOT
	MAIN_ERR_KIND_DESTINATION_IS_MARKDOWN_FILE_BUT_TARGET_IS_NOT
	MAIN_ERR_KIND_INVALID_REMAP_FORMAT
	MAIN_ERR_KIND_INVALID_FILTER_FORMAT
	MAIN_ERR_KIND_FORMAT_LINK_NEEDS_LINK
	MAIN_ERR_KIND_INVALID_ANCHOR_FORMATTING_STYLE
	MAIN_ERR_KIND_REMAP_PATH_PREFIX_NEEDS_LINK
	MAIN_ERR_KIND_INVALID_REMAP_PATH_PREFIX_FORMAT
	// MAIN_ERR_KIND_BASE_URL_NEEDS_LINK
)

type mainErr interface {
	error
	Kind() mainErrKind
}

type mainErrImpl struct {
	kind    mainErrKind
	message string
}

func (e *mainErrImpl) Error() string {
	return e.message
}

func (e *mainErrImpl) Kind() mainErrKind {
	return e.kind
}

func newMainErr(kind mainErrKind) mainErr {
	err := new(mainErrImpl)
	switch kind {
	case MAIN_ERR_KIND_SOURCE_NOT_SET:
		err.message = fmt.Sprintf("%s was not set", FLAG_SOURCE)
	case MAIN_ERR_KIND_DESTINATION_NOT_SET:
		err.message = fmt.Sprintf("%s was not set", FLAG_DESTINATION)
	case MAIN_ERR_KIND_STRICTREF_NEEDS_LINK:
		err.message = fmt.Sprintf("%s set but not %s", FLAG_STRICT_REF, FLAG_CONVERT_LINKS)
	case MAIN_ERR_KIND_INVALID_SOURCE_FORMAT:
		err.message = fmt.Sprintf("%s shouldn't begin with \"-\"", FLAG_SOURCE)
	case MAIN_ERR_KIND_INVALID_DESTINATION_FORMAT:
		err.message = fmt.Sprintf("%s shouldn't begin with \"-\"", FLAG_DESTINATION)
	case MAIN_ERR_KIND_TARGET_IS_MARKDOWN_FILE_BUT_DESTINATION_IS_NOT:
		err.message = fmt.Sprintf("%s is a markdown file but %s is not", FLAG_TARGET, FLAG_DESTINATION)
	case MAIN_ERR_KIND_DESTINATION_IS_MARKDOWN_FILE_BUT_TARGET_IS_NOT:
		err.message = fmt.Sprintf("%s is a markdown file but %s is not", FLAG_DESTINATION, FLAG_TARGET)
	case MAIN_ERR_KIND_INVALID_FILTER_FORMAT:
		err.message = fmt.Sprintf("%s has an invalid format", FLAG_FILTER)
	case MAIN_ERR_KIND_INVALID_ANCHOR_FORMATTING_STYLE:
		err.message = fmt.Sprintf("%s is invalid. must choose from %s", FLAG_FORMAT_ANCHOR, strings.Join(convert.ANCHOR_FORMATTING_STYLES, ", "))
	// case MAIN_ERR_KIND_BASE_URL_NEEDS_LINK:
	// 	err.message = fmt.Sprintf("%s set but not %s", FLAG_BASE_URL, FLAG_CONVERT_LINKS)
	case MAIN_ERR_KIND_FORMAT_LINK_NEEDS_LINK:
		err.message = fmt.Sprintf("%s set but not %s", FLAG_FORMAT_LINK, FLAG_CONVERT_LINKS)
	case MAIN_ERR_KIND_REMAP_PATH_PREFIX_NEEDS_LINK:
		err.message = fmt.Sprintf("%s set but not %s", FLAG_REMAP_PATH_PREFIX, FLAG_CONVERT_LINKS)
	default:
		err.kind = MAIN_ERR_UNEXPECTED
		err.message = "unexpected error"
		return err
	}
	err.kind = kind
	return err
}

func newMainErrf(kind mainErrKind, format string, a ...interface{}) mainErr {
	err := new(mainErrImpl)
	err.kind = kind
	err.message = fmt.Sprintf(format, a...)
	return err
}

func initFlags(flagset *flag.FlagSet, config *configuration) {
	flagset.StringVar(&config.src, FLAG_SOURCE, "", "source directory")
	flagset.StringVar(&config.dst, FLAG_DESTINATION, "", "destination directory")
	flagset.StringVar(&config.tgt, FLAG_TARGET, "", "the path that will be processed. It can be a file or a directory. The default value of tgt = the directory specified by src flag. This option will be used when you want to process only a subset of a vault but resolve refs by the entire vault.")
	flagset.BoolVar(&config.rmtag, FLAG_REMOVE_TAGS, false, "remove tag")
	flagset.BoolVar(&config.cptag, FLAG_COPY_TAGS, false, "copy tag to tags field of front matter")
	flagset.BoolVar(&config.synctag, FLAG_SYNC_TAGS, false, "remove all tags in front matter and then copy tags from text")
	flagset.BoolVar(&config.title, FLAG_COPY_TITLE, false, "copy h1 content to title field of front matter")
	flagset.BoolVar(&config.alias, FLAG_COPY_ALIASES, false, "copy add h1 content to aliases field of front matter")
	flagset.BoolVar(&config.synctlal, FLAG_SYNC_TITLE_ALIASES, false, "remove an alias appearing also in title field and then copy h1 content to title and aliases fields")
	flagset.BoolVar(&config.link, FLAG_CONVERT_LINKS, false, "convert obsidian internal and external links to external links in the usual format")
	flagset.BoolVar(&config.cmmt, FLAG_REMOVE_COMMENT, false, "remove obsidian comment")
	flagset.BoolVar(&config.publishable, FLAG_PUBLISHABLE, false, "process only files with publish: true or draft: false. For files with publish: true, add draft: false.")
	flagset.BoolVar(&config.rmH1, FLAG_REMOVE_H1, false, "remove H1")
	flagset.BoolVar(&config.strictref, FLAG_STRICT_REF, false, fmt.Sprintf("return error when ref target is not found. available only when %s is on", FLAG_CONVERT_LINKS))
	flagset.StringVar(&config.remapkey, FLAG_REMAP_META_KEYS, "", "remap keys in front matter. format: \"old1:new1,old2:new2\". If a new key is not specified (i.e., empty string), then the field will be removed.")
	flagset.StringVar(&config.filter, FLAG_FILTER, "", "process only files with specified conditions. Example: -filter=\"(key1||!key2)&&key3\". Each field must be boolean and each key must match /[a-zA-Z-_]+/.")
	// flagset.StringVar(&config.baseUrl, FLAG_BASE_URL, "", "prefix resolved internal links and format it. Example (-baseUrl=https://example.com/): sample -> https://example.com/sample")
	flagset.StringVar(&config.remapPathPrefix, FLAG_REMAP_PATH_PREFIX, "", "remap prefixes in paths. Example (-remapPrefix=static/>images/|notes/>posts/): static/sample.png -> images/sample.png, notes/sample.md -> posts/sample.md")
	flagset.BoolVar(&config.formatLink, FLAG_FORMAT_LINK, false, "trim suffix .md and complete link. Example: #section -> path/to/sample#section")
	flagset.StringVar(&config.formatAnchor, FLAG_FORMAT_ANCHOR, convert.FORMAT_ANCHOR_HUGO, fmt.Sprintf("anchor formatting style. Available styles: %s", strings.Join(convert.ANCHOR_FORMATTING_STYLES, ", ")))
	flagset.BoolVar(&config.obs, FLAG_OBSIDIAN_USAGE, false, "alias of -cptag -title -alias")
	flagset.BoolVar(&config.std, FLAG_STANDARD_USAGE, false, "alias of -cptag -rmtag -title -alias -link -cmmt -strictref")
	flagset.BoolVar(&config.ver, FLAG_VERSION, false, "display the version currently installed")
	flagset.BoolVar(&config.debug, FLAG_DEBUG, false, "display error message for developers")
}

// const FORMAT_ANCHOR_HUGO = "hugo"
// const FORMAT_ANCHOR_MARKDOWN_IT = "markdownit"

// var ANCHOR_FORMATTING_STYLES = []string{FORMAT_ANCHOR_HUGO, FORMAT_ANCHOR_MARKDOWN_IT}

// 実行前に↓が必要
// 1. initFlags(flagset, flags)
// 2. flag の値の設定
// 	- flag.CommandLine => flag.Parse()
//	- それ以外 => flagset.Set("フラグ名", "フラグの値を表す文字列")
func setConfig(flagset *flag.FlagSet, config *configuration) {
	orgFlag := *config
	setflags := make(map[string]struct{})
	flagset.Visit(func(f *flag.Flag) {
		setflags[f.Name] = struct{}{}
	})

	if config.obs || config.std {
		config.cptag = true
		config.title = true
		config.alias = true
	}
	if config.std {
		config.rmtag = true
		config.link = true
		config.cmmt = true
		config.strictref = true
	}
	config.tgt = config.src

	if _, ok := setflags[FLAG_COPY_TAGS]; ok {
		config.cptag = orgFlag.cptag
	}
	if _, ok := setflags[FLAG_COPY_TITLE]; ok {
		config.title = orgFlag.title
	}
	if _, ok := setflags[FLAG_COPY_ALIASES]; ok {
		config.alias = orgFlag.alias
	}
	if _, ok := setflags[FLAG_REMOVE_TAGS]; ok {
		config.rmtag = orgFlag.rmtag
	}
	if _, ok := setflags[FLAG_CONVERT_LINKS]; ok {
		config.link = orgFlag.link
	}
	if _, ok := setflags[FLAG_REMOVE_COMMENT]; ok {
		config.cmmt = orgFlag.cmmt
	}
	if _, ok := setflags[FLAG_PUBLISHABLE]; ok {
		config.publishable = orgFlag.publishable
	}
	if _, ok := setflags[FLAG_STRICT_REF]; ok {
		config.strictref = orgFlag.strictref
	}
	if _, ok := setflags[FLAG_TARGET]; ok {
		config.tgt = orgFlag.tgt
	}
}

func verifyConfig(config *configuration) error {
	if config.src == "" {
		return newMainErr(MAIN_ERR_KIND_SOURCE_NOT_SET)
	}
	if config.dst == "" {
		return newMainErr(MAIN_ERR_KIND_DESTINATION_NOT_SET)
	}
	if strings.HasPrefix(config.src, "-") {
		return newMainErr(MAIN_ERR_KIND_INVALID_SOURCE_FORMAT)
	}
	if strings.HasPrefix(config.dst, "-") {
		return newMainErr(MAIN_ERR_KIND_INVALID_DESTINATION_FORMAT)
	}
	if config.strictref && !config.link {
		return newMainErr(MAIN_ERR_KIND_STRICTREF_NEEDS_LINK)
	}
	// if config.baseUrl != "" && !config.link {
	// 	return newMainErr(MAIN_ERR_KIND_BASE_URL_NEEDS_LINK)
	// }
	var validAnchorFormattingStyle bool
	for _, style := range convert.ANCHOR_FORMATTING_STYLES {
		if config.formatAnchor == style {
			validAnchorFormattingStyle = true
			break
		}
	}
	if !validAnchorFormattingStyle {
		return newMainErr(MAIN_ERR_KIND_INVALID_ANCHOR_FORMATTING_STYLE)
	}

	if config.remapPathPrefix != "" && !config.link {
		return newMainErr(MAIN_ERR_KIND_INVALID_REMAP_FORMAT)
	}
	if config.formatLink && !config.link {
		return newMainErr(MAIN_ERR_KIND_FORMAT_LINK_NEEDS_LINK)
	}
	// check roughly if tgt and dst are the same type (regular file or directory)
	if filepath.Ext(config.tgt) == ".md" && filepath.Ext(config.dst) != ".md" {
		return newMainErrf(MAIN_ERR_KIND_TARGET_IS_MARKDOWN_FILE_BUT_DESTINATION_IS_NOT, "%s is a markdown file but %s is not", config.tgt, config.dst)
	}
	if filepath.Ext(config.tgt) != ".md" && filepath.Ext(config.dst) == ".md" {
		return newMainErrf(MAIN_ERR_KIND_DESTINATION_IS_MARKDOWN_FILE_BUT_TARGET_IS_NOT, "%s is a markdown file but %s is not", config.dst, config.tgt)
	}
	return nil
}
