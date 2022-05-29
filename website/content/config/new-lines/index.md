---
title: "New lines"
date: 
draft: false
weight: 7
summary: Configure amount of new lines will be placed.
---

Number of new lines before and after parts of the change log can be adjusted from the config.

Example for 2 lines before and after kind header:

```yml
newlines:
  beforeKindHeader: 2
  afterKindHeader: 2
```

It will look like this:
```
## v1.0.0 - 2022-05-28


### Added


* commit massage
```

## Newlines Options:
For each of the options the type is int and default is 0:

type: `int` | default: `0`

-	beforeVersion
- afterVersion
- beforeComponent
- afterComponent
-	beforeHeader
-	afterHeader
-	beforeFooter
-	afterFooter
-	beforeHeaderFile
-	afterHeaderFile
-	beforeFooterFile
-	afterFooterFile
-	beforeKindHeader
-	afterKindHeader
-	beforeChange
-	afterChange