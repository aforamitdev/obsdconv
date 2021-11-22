---
aliases:
- existing-alias
- sample file for -std (= -cptag -rmtag -title -alias -link -cmmt -strictref) <<  >>
draft: false
publish: true
tags:
- existing-tag
- obsidian
- will_be_removed_from_text
- will_be_removed_in_title_and_alias
title: sample file for -std (= -cptag -rmtag -title -alias -link -cmmt -strictref)
  <<  >>
---
# sample file for -std (= -cptag -rmtag -title -alias -link -cmmt -strictref) <<  >>

## Copy tags
Tags will be copied to `tags` field in front matter.
<<  >> <- `#obsidian` will be copied (and be removed).

### Not tags
Tags are escaped in the following.

#### Code Block
```
	#code-block
```

#### Math Block
$$
	#math-block
$$

#### Comment Block

(↑ this comment block will be removed)

#### Inline Code
`#inline-code`

#### Inline Math
$#inline-math$


## Set titles
H1 content will be copied to `title` field in front matter.
In this case,
- tags are removed,
- internal links and external links are converted to display names only.

## Set aliases
H1 content will be copied to `aliases` field in front matter.
H1 content will be processed like `title`.

### Remove Tags
<<  >> <- `#will_be_removed_from_text` will be removed

## Convert Links
### Internal Links
[blank](blank.md)

### Obsidian URL
[obsidian url](blank.md)

### Embeds
![image.png](image.png)

## Remove Obsidian Comment Blocks


## Publish
The generated file will contain `draft: false`, since this file contains `publish: true` in its front matter.
On the other hand, `blank.md` in the same directory will not be generated, since it contains no `publish` nor `draft` field in its frontmatter.
