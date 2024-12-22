# Mangal Configuration Guide

This document describes all available configuration options for Mangal. Configuration can be set either through environment variables or a TOML configuration file.

## Configuration File

The default configuration file is named `mangal.toml` and should be placed in your config directory.

## Environment Variables

All configuration options can be set via environment variables. The environment variable names are derived from the TOML keys by:
1. Converting to uppercase
2. Replacing dots (.) with underscores (_)
3. Adding the `MANGAL_` prefix

For example, `downloader.path` becomes `MANGAL_DOWNLOADER_PATH`.

## Configuration Options

### Downloader Settings

| Option | Environment Variable | TOML Key | Description | Default |
|--------|-------------------|-----------|-------------|---------|
| Download Path | `MANGAL_DOWNLOADER_PATH` | `downloader.path` | Base directory for downloaded manga | `~/Downloads/manga` |
| Chapter Name Template | `MANGAL_DOWNLOADER_CHAPTER_NAME_TEMPLATE` | `downloader.chapter_name_template` | Template for chapter filenames | `Chapter {chapter_number}` |
| Async Download | `MANGAL_DOWNLOADER_ASYNC` | `downloader.async` | Enable asynchronous downloads | `true` |
| Create Manga Directory | `MANGAL_DOWNLOADER_CREATE_MANGA_DIR` | `downloader.create_manga_dir` | Create directory per manga | `true` |
| Create Volume Directory | `MANGAL_DOWNLOADER_CREATE_VOLUME_DIR` | `downloader.create_volume_dir` | Create directory per volume | `false` |
| Default Sources | `MANGAL_DOWNLOADER_DEFAULT_SOURCES` | `downloader.default_sources` | List of default manga sources | `[]` |
| Stop on Error | `MANGAL_DOWNLOADER_STOP_ON_ERROR` | `downloader.stop_on_error` | Stop downloading on error | `false` |
| Download Cover | `MANGAL_DOWNLOADER_DOWNLOAD_COVER` | `downloader.download_cover` | Download manga cover image | `true` |
| Redownload Existing | `MANGAL_DOWNLOADER_REDOWNLOAD_EXISTING` | `downloader.redownload_existing` | Redownload existing chapters | `false` |
| Read Downloaded | `MANGAL_DOWNLOADER_READ_DOWNLOADED` | `downloader.read_downloaded` | Open reader after download | `false` |

### Format Settings

| Option | Environment Variable | TOML Key | Description | Default |
|--------|-------------------|-----------|-------------|---------|
| Format to Use | `MANGAL_FORMATS_USE` | `formats.use` | Output format (pdf, cbz, zip, plain) | `cbz` |
| Skip Unsupported Images | `MANGAL_FORMATS_SKIP_UNSUPPORTED_IMAGES` | `formats.skip_unsupported_images` | Skip unsupported image formats | `false` |

### Metadata Settings

| Option | Environment Variable | TOML Key | Description | Default |
|--------|-------------------|-----------|-------------|---------|
| Fetch Anilist | `MANGAL_METADATA_FETCH_ANILIST` | `metadata.fetch_anilist` | Fetch metadata from Anilist | `true` |
| ComicInfo XML | `MANGAL_METADATA_COMIC_INFO_XML` | `metadata.comic_info_xml` | Generate ComicInfo.xml | `true` |
| Add Date to ComicInfo | `MANGAL_METADATA_COMIC_INFO_XML_ADD_DATE` | `metadata.comic_info_xml_add_date` | Add date to ComicInfo.xml | `true` |
| Alternative Date Format | `MANGAL_METADATA_COMIC_INFO_XML_ALTERNATIVE_DATE` | `metadata.comic_info_xml_alternative_date` | Use alternative date format | `false` |
| Tag Relevance Threshold | `MANGAL_METADATA_COMIC_INFO_XML_TAG_RELEVANCE_THRESHOLD` | `metadata.comic_info_xml_tag_relevance_threshold` | Minimum relevance for tags | `0.5` |
| Series JSON | `MANGAL_METADATA_SERIES_JSON` | `metadata.series_json` | Generate series.json | `true` |
| Debug Metadata | `MANGAL_METADATA_DEBUG` | `metadata.debug` | Enable metadata debug logging | `false` |

### Reader Settings

| Option | Environment Variable | TOML Key | Description | Default |
|--------|-------------------|-----------|-------------|---------|
| PDF Reader | `MANGAL_READER_PDF` | `reader.pdf` | PDF reader command | `""` |
| CBZ Reader | `MANGAL_READER_CBZ` | `reader.cbz` | CBZ reader command | `""` |
| ZIP Reader | `MANGAL_READER_ZIP` | `reader.zip` | ZIP reader command | `""` |
| Plain Reader | `MANGAL_READER_PLAIN` | `reader.plain` | Plain reader command | `""` |
| Browser Reader | `MANGAL_READER_BROWSER` | `reader.browser` | Browser command | `""` |
| Folder Reader | `MANGAL_READER_FOLDER` | `reader.folder` | Folder viewer command | `""` |
| Read in Browser | `MANGAL_READER_READ_IN_BROWSER` | `reader.read_in_browser` | Open in browser by default | `false` |

### Anilist Integration

| Option | Environment Variable | TOML Key | Description | Default |
|--------|-------------------|-----------|-------------|---------|
| Enable Anilist | `MANGAL_ANILIST_ENABLE` | `anilist.enable` | Enable Anilist integration | `false` |
| Anilist ID | `MANGAL_ANILIST_ID` | `anilist.id` | Anilist client ID | `""` |
| Anilist Secret | `MANGAL_ANILIST_SECRET` | `anilist.secret` | Anilist client secret | `""` |
| Anilist Code | `MANGAL_ANILIST_CODE` | `anilist.code` | Anilist auth code | `""` |
| Link on Manga Select | `MANGAL_ANILIST_LINK_ON_MANGA_SELECT` | `anilist.link_on_manga_select` | Link manga on selection | `false` |

### TUI Settings

| Option | Environment Variable | TOML Key | Description | Default |
|--------|-------------------|-----------|-------------|---------|
| Item Spacing | `MANGAL_TUI_ITEM_SPACING` | `tui.item_spacing` | Spacing between items | `1` |
| Read on Enter | `MANGAL_TUI_READ_ON_ENTER` | `tui.read_on_enter` | Start reading on Enter key | `false` |
| Search Prompt | `MANGAL_TUI_SEARCH_PROMPT` | `tui.search_prompt` | Search prompt text | `Search: ` |
| Show URLs | `MANGAL_TUI_SHOW_URLS` | `tui.show_urls` | Show manga URLs | `false` |
| Show Downloaded Path | `MANGAL_TUI_SHOW_DOWNLOADED_PATH` | `tui.show_downloaded_path` | Show download paths | `true` |
| Reverse Chapters | `MANGAL_TUI_REVERSE_CHAPTERS` | `tui.reverse_chapters` | Reverse chapter order | `false` |

### History Settings

| Option | Environment Variable | TOML Key | Description | Default |
|--------|-------------------|-----------|-------------|---------|
| Save on Read | `MANGAL_HISTORY_SAVE_ON_READ` | `history.save_on_read` | Save to history when reading | `true` |
| Save on Download | `MANGAL_HISTORY_SAVE_ON_DOWNLOAD` | `history.save_on_download` | Save to history when downloading | `true` |

### Installer Settings

| Option | Environment Variable | TOML Key | Description | Default |
|--------|-------------------|-----------|-------------|---------|
| Installer User | `MANGAL_INSTALLER_USER` | `installer.user` | GitHub username for updates | `""` |
| Installer Repo | `MANGAL_INSTALLER_REPO` | `installer.repo` | GitHub repository for updates | `""` |
| Installer Branch | `MANGAL_INSTALLER_BRANCH` | `installer.branch` | GitHub branch for updates | `""` |

## Example Configuration

Here's an example `mangal.toml` configuration file:

```toml
[downloader]
path = "~/manga"
chapter_name_template = "Chapter {chapter_number}"
async = true
create_manga_dir = true
download_cover = true

[formats]
use = "cbz"

[metadata]
fetch_anilist = true
comic_info_xml = true

[reader]
read_in_browser = false

[anilist]
enable = false

[tui]
item_spacing = 1
search_prompt = "Search: "
show_downloaded_path = true
```
