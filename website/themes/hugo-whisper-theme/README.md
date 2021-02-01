# Hugo Whisper Theme

Whisper is a minimal documentation theme built for Hugo. The design and functionality is intentionally minimal.


[Live Demo](https://hugo-whisper.netlify.app/) |
[Zerostatic Themes](https://www.zerostatic.io/)

![Hugo Whisper Theme screenshot](https://github.com/JugglerX/hugo-whisper-theme/blob/master/images/screenshot-with-border.png)

## Theme features

### Content Types

- Docs (Markdown)
- Homepage

### Content Management

- This theme generates documentation from markdown files located in `content/docs`
- The "Home" page is not documentation, it can be used to introduce your project etc.

### Features

- Beautiful and clean typography for all semantic HTML elements

### SCSS

- SCSS (Hugo Pipelines)
- Responsive design
- Bootstrap 4 grid and media queries only

### Speed

- 100/100 Google Lighthouse speed score
- 21KB without images ⚡
- Vanilla JS only

### Menu

- Responsive mobile menu managed in `config.toml`

### Content

- Documentation examples included, using all markdown syntax

# Installation

To use this theme you will need to have Hugo installed. If you don't already have Hugo installed please follow the official [installation guide](https://gohugo.io/getting-started/installing/)

### Check Hugo version (Hugo 0.51+ Extended is required)

This theme uses [Hugo Pipes](https://gohugo.io/hugo-pipes/scss-sass/) to compile SCSS and minify assets. Please make sure you have the **Hugo Extended** version installed. If you are not using the extended version this theme will not not compile.

To check your version of Hugo, run:

```
hugo version
```

This will output the currently installed version of Hugo. Make sure you see `/extended` after the version number, for example `Hugo Static Site Generator v0.51/extended darwin/amd64 BuildDate: unknown` You do not need to use version v0.51 specifically, you can use any version of Hugo above 0.51. It just needs to have the `/extended` part

### Create a new Hugo site

```
hugo new site mynewsite
```

This will create a fresh Hugo site in the folder `mynewsite`.

### Install theme

Copy or git clone this theme into the sites themes folder `mynewsite/themes`

#### Install with Git

```
cd mynewsite
git clone https://github.com/jugglerx/hugo-whisper-theme.git themes/hugo-whisper-theme
```

#### Install from .zip file

You can download the .zip file located here https://github.com/JugglerX/hugo-whisper-theme/archive/master.zip.

Extract the downloaded .zip inside the `themes` folder. Rename the extracted folder from `hugo-whisper-theme-master` -> `hugo-whisper-theme`. You should end up with the following folder structure `mynewsite/themes/hugo-whisper-theme`

### Add example content

Copy the entire contents of the `mynewsite/themes/hugo-whisper-theme/exampleSite/` folder to root folder of your Hugo site, ie `mynewsite/`

To copy the files using terminal, make sure you are still in the projects root, ie the `mynewsite` folder.

```
cp -a themes/hugo-whisper-theme/exampleSite/. .
```

### Update config.toml

After you copy the `config.toml` into the root folder of your Hugo site you will need to update the `baseURL`, `themesDir` and `theme` values in `mynewsite/config.toml`

```
baseURL = "/"
themesDir = "themes"
theme = "hugo-whisper-theme"
```

### Run Hugo

After installing the theme for the first time, generate the Hugo site.

You run this command from the root folder of your Hugo site ie `mynewsite/`

```
hugo
```

For local development run Hugo's built-in local server.

```
hugo server
```

Now enter [`localhost:1313`](http://localhost:1313) in the address bar of your browser.

## License

If you fork or copy this theme the LICENSE file and the copyright notice on line 3 (where I am listed as the author) must not be changed. You cannot just replace the copyright line with your own name. Attribution in your README.md or on your site is also welcome but not required.
