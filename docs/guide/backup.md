---
title: "Backup"
description: How to backup an existing CHANGELOG file
---

If you are adding changie to the workflow for an existing project that includes a CHANGELOG
you can follow these steps to keep your existing changes.

1. Rename the existing `CHANGELOG.md`
1. Follow the [quick start guide](quick_start.md)
1. Move your backup changelog into the generated `changes` folder named as your latest release.
    * For example, if you just released v1.2.0 rename your changelog to `changes/v1.2.0.md`
    * If you are using another file extension and not markdown than adjust accordingly
1. Cut and paste the heading from your previous changelog into the generated `header.tpl.md` file
1. Run `changie merge` to regenerate your changelog to make sure it looks right

You can now use changie as normal, you do not need to recreate all previous version files.
