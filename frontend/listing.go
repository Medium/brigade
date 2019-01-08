package frontend

import (
	"fmt"
	"html/template"
	"path"
	"time"

	"github.com/Medium/brigade/backend"
)

type Listing struct {
	Host    string
	Path    string
	Entries []*backend.Entry
}

var listingTemplate = template.Must(
	template.New("test").
		Funcs(template.FuncMap{
			"size":   FormatSize,
			"time":   FormatTime,
			"parent": path.Dir,
		}).
		Parse(listingHTML),
)

func FormatSize(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func FormatTime(t *time.Time) string {
	if t == nil {
		return ""
	}

	const format = "2006-01-02 15:04:05 MST"
	return t.Format(format)
}

var listingHTML = `
<!doctype html>
<html>
	<head>
		<title>/{{ .Path }} • {{ .Host }}</title>
		<link href="https://fonts.googleapis.com/css?family=IBM+Plex+Mono:200,300,400|Material+Icons" rel="stylesheet">
		<style type="text/css">
			html, body {
				background: white;
				color: #333;
				margin: 0;
				padding: 0;
				font-family: "IBM Plex Mono", monospace;
			}

			.container {
				max-width: 60rem;
				margin: 0 auto;
				padding: 0 2rem;
			}

			header {
				background: #D0EEFE;
				color: #00A6FB;
				padding: 1em;
			}

			header h1 {
				font-size: 1.3rem;
				margin: 0;
				font-weight: 400;
			}

			header .up {
				display: inline-block;
				margin-left: -28px;
				width: 24px;
			}

			main {
				margin: 2rem 0;
			}

			a:link, a:visited {
				color: #00A6FB;
			}

			table.listing {
				box-sizing: border-box;
				width: 100%;
				margin: 0 0 0 -28px;
				padding: 0;
				border: none;
				border-collapse: collapse;
				border-spacing: 0;
				empty-cells: show;
			}

			td.icon {
				width: 28px;
				color: #5cc6fc;
				vertical-align: middle;
			}

			td.text {
				padding: 4px 0;	
				vertical-align: middle;
			}

			td.name {
				font-weight: 300;
			}

			td.detail {
				color: #666;
				font-weight: 200;
				font-size: 0.9em;
			}

			td.size {
				width: 8em;
				padding-right: 2em;
				text-align: right;
			}

			td.date {
				width: 16em;
			}

			footer {
				color: #666;
				font-weight: 300;
				font-size: 0.85rem;
				margin: 2em 0;
				text-align: center;
			}

			footer a:link, footer a:visited {
				color: #444;
			}
		</style>
	</head>
	<body>
		<header>
			<div class="container">
				<h1>
					{{ if ne .Path "" }}
						<a class="up" href="/{{ .Path | parent | parent }}/">
							<i class="material-icons">keyboard_arrow_up</i>
						</a>
					{{ end }}
					{{ if ne .Path "" }}
						/{{ .Path }}
					{{ else }}
						{{ .Host }}/
					{{ end }}
				</h1>
			</div>
		</header>
		<main>
			<div class="container">
				<table class="listing">
						{{ $path := .Path }}
						{{ range .Entries }}
							<tr>
								<td class="icon">
									{{ if eq .Type "directory" }}
										<i class="material-icons">folder_open</i>
									{{ end }}
								</td>
								<td class="name text">
									<a href="/{{ $path }}{{ .Name }}">{{ .Name }}</a>
								</td>
								<td class="size detail text">
									{{ if eq .Type "file" }}
										{{ .Size | size }}
									{{ end }}
								</td>
								<td class="date detail text">
									{{ if eq .Type "file" }}
										{{ .LastModified | time }}
									{{ end }}
								</td>
							</tr>
						{{ end }}
				</table>
			</div>
		</main>
		<footer>
			Powered by <a href="https://github.com/Medium/brigade">Brigade</a>.
		</footer>
	</body>
</html>
`
