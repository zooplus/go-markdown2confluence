# markdown2confluence

Push markdown files to Confluence Cloud

[![Build & Release](https://github.com/zooplus/go-markdown2confluence/actions/workflows/build-go.yml/badge.svg)](https://github.com/zooplus/go-markdown2confluence/actions/workflows/build-go.yml)

## Installation

Download the binary for your platform from the [latest
release](https://github.com/zooplus/go-markdown2confluence/releases/latest)
and add it to your `PATH`. Binaries are named
`markdown2confluence-<os>-<arch>` (`linux`, `macos`, `windows`; `amd64`,
`arm64`) and checksums are published in `sha512sums.txt`.

- Linux

  ```shell
  curl -L -o markdown2confluence https://github.com/zooplus/go-markdown2confluence/releases/latest/download/markdown2confluence-linux-amd64
  chmod +x markdown2confluence && sudo mv markdown2confluence /usr/local/bin/
  ```

- macOS

  ```shell
  curl -L -o markdown2confluence https://github.com/zooplus/go-markdown2confluence/releases/latest/download/markdown2confluence-macos-arm64
  chmod +x markdown2confluence && sudo mv markdown2confluence /usr/local/bin/
  ```

- Windows

  Download `markdown2confluence-windows-amd64.exe` from [the latest release](https://github.com/zooplus/go-markdown2confluence/releases/latest) and add it to your system `PATH`

## Building locally

```shell
go build -ldflags="-X main.version=dev" -o markdown2confluence .
```

## Environment Variables

For best practice we recommend you [authenticate using an API token](https://id.atlassian.com/manage/api-tokens).

However, you may also use [Personal Access Tokens](https://confluence.atlassian.com/enterprise/using-personal-access-tokens-1026032365.html),
which may help if your company uses SSO.

- CONFLUENCE_USERNAME - username for Confluence Cloud. When using API tokens set this to your full email.
- CONFLUENCE_PASSWORD - API token or password for Confluence Cloud
- CONFLUENCE_ENDPOINT - endpoint for Confluence Cloud, eg `https://mycompanyname.atlassian.net/wiki`
- CONFLUENCE_ACCESS_TOKEN - Bearer access token to use (instead of API token)

## Usage

```txt
Push markdown files to Confluence Cloud

Usage:
  markdown2confluence [flags]

Flags:
  -a, --access-token string   Confluence access-token. (Alternatively set CONFLUENCE_ACCESS_TOKEN environment variable)
  -c, --comment string        (Optional) Add comment to page
  -d, --debug                 Enable debug logging
  -e, --endpoint string       Confluence endpoint. (Alternatively set CONFLUENCE_ENDPOINT environment variable) (default "https://mydomain.atlassian.net/wiki")
  -x, --exclude strings       list of exclude file patterns (regex) for that will be applied on markdown file paths
  -w, --hardwraps             Render newlines as <br />
  -h, --help                  help for markdown2confluence
  -i, --insecuretls           Skip certificate validation. (e.g. for self-signed certificates)
  -m, --modified-since int    Only upload files that have modifed in the past n minutes
      --parent string         Optional parent page to next content under
  -p, --password string       Confluence password. (Alternatively set CONFLUENCE_PASSWORD environment variable)
  -s, --space string          Space in which page should be created
  -t, --title string          Set the page title on upload (defaults to filename without extension)
      --use-document-title    Will use the Markdown document title (# Title) if available
  -u, --username string       Confluence username. (Alternatively set CONFLUENCE_USERNAME environment variable)
      --version               version for markdown2confluence
```

## Examples

Upload a local directory of markdown files called `markdown-files` to Confluence.

```shell
markdown2confluence \
  --space 'MyTeamSpace' \
  markdown-files
```

Upload the same directory, but only those modified in the last 30 minutes. This is particurlarly useful for cron jobs/recurring one-way syncs.

```shell
markdown2confluence \
  --space 'MyTeamSpace' \
  --modified-since 30 \
  markdown-files
```

Upload a single file

```shell
markdown2confluence \
  --space 'MyTeamSpace' \
  markdown-files/test.md
```

Upload a directory of markdown files in space `MyTeamSpace` under the parent page `API Docs`

```shell
markdown2confluence \
  --space 'MyTeamSpace' \
  --parent 'API Docs' \
  markdown-files
```

Upload a directory of markdown files in space `MyTeamSpace` under a _nested_ parent page `Docs/API` and _exclude_ mardown files/directories that match `.*generated.*` or `.*temp.md`

```shell
markdown2confluence \
  --space 'MyTeamSpace' \
  --parent 'API/Docs' \
  --exclude '.*generated.*' \
  --exclude '.*temp.md' \
   markdown-files
```

Upload a directory of markdown files in space `MyTeamSpace` under the parent page `API Docs` and use the markdown _document-title_ instead of the filname as document title (if available) in Confluence.

```shell
markdown2confluence \
  --space 'MyTeamSpace' \
  --parent 'API Docs' \
  --use-document-title \
   markdown-files
```

## Enhancements

It is possible to insert Confluence macros using fenced code blocks.
The "language" for this is `CONFLUENCE-MACRO`, exactly like that in all-caps.
Here is an example for a ToC macro using all headlines starting at Level 2:

````markdown
    # Title

    ```CONFLUENCE-MACRO
    name:toc
    schema-version:1
      minLevel:2
    ```

    ## Section 1
````

In general almost all macros should be possible.
The general syntax is:

````markdown
    ```CONFLUENCE-MACRO
    name:Name of Macro
    schema-version:Schema Version (use `1`)
      attribute:Value of Attribute
      parameter-name:Value of Parameter
      next-parameter:Value of Parameter
    ```
````

So a fully fledged macro could look like:

````markdown
    ```CONFLUENCE-MACRO
    name:toc
    schema-version:1
      maxLevel:5
      minLevel:2
      exclude:Beispiel.*
      style:none
      type:flat
      separator:pipe
    ```
````

Which will translate to:

```XML
<ac:structured-macro ac:name="toc" ac:schema-version="1" >
  <ac:parameter ac:name="maxLevel">5</ac:parameter>
  <ac:parameter ac:name="minLevel">2</ac:parameter>
  <ac:parameter ac:name="exclude">Beispiel.*</ac:parameter>
  <ac:parameter ac:name="style">none</ac:parameter>
  <ac:parameter ac:name="type">flat</ac:parameter>
  <ac:parameter ac:name="separator">pipe</ac:parameter>
</ac:structured-macro>
```
