package http

import (
	"fmt"
	"html/template"
	"task-service/internal/storage"
)

func renderIndex(tasks []storage.Task, draftCount int) string {
	draftHTML := renderDraftPanel(tasks)
	progressHTML := renderInProgressPanel(tasks)
	doneHTML := renderDonePanel(tasks)

	tmpl := template.Must(template.New("index").Parse(indexHTML))

	data := struct {
		DraftCount    int
		DraftPanel    template.HTML
		ProgressPanel template.HTML
		DonePanel     template.HTML
	}{
		DraftCount:    draftCount,
		DraftPanel:    template.HTML(draftHTML),
		ProgressPanel: template.HTML(progressHTML),
		DonePanel:     template.HTML(doneHTML),
	}

	var buf stringWriter
	tmpl.Execute(&buf, data)
	return buf.String()
}

func renderDraftPanel(tasks []storage.Task) string {
	var drafts []storage.Task
	for _, t := range tasks {
		if t.Status == storage.Draft {
			drafts = append(drafts, t)
		}
	}

	html := fmt.Sprintf(`<h3 class="panel-title">Drafts <span class="count-badge">%d</span></h3>`, len(drafts))
	for _, t := range drafts {
		html += renderTaskCard(t, true)
	}
	if len(drafts) == 0 {
		html += `<div class="empty-state">No drafts yet</div>`
	}
	return html
}

func renderInProgressPanel(tasks []storage.Task) string {
	var inProgress []storage.Task
	for _, t := range tasks {
		if t.Status == storage.InProgress {
			inProgress = append(inProgress, t)
		}
	}

	html := fmt.Sprintf(`<h3 class="panel-title">In Progress <span class="count-badge">%d</span></h3>`, len(inProgress))
	for _, t := range inProgress {
		html += renderTaskCard(t, false)
	}
	if len(inProgress) == 0 {
		html += `<div class="empty-state">Nothing in progress</div>`
	}
	return html
}

func renderDonePanel(tasks []storage.Task) string {
	var done []storage.Task
	for _, t := range tasks {
		if t.Status == storage.Done {
			done = append(done, t)
		}
	}

	html := fmt.Sprintf(`<h3 class="panel-title">Done <span class="count-badge">%d</span></h3>`, len(done))
	for _, t := range done {
		html += renderTaskCard(t, false)
	}
	if len(done) == 0 {
		html += `<div class="empty-state">No completed tasks</div>`
	}
	return html
}

func renderTaskCard(t storage.Task, isDraft bool) string {
	statusClass := "status-draft"
	switch t.Status {
	case storage.InProgress:
		statusClass = "status-progress"
	case storage.Done:
		statusClass = "status-done"
	}

	nextStatus := ""
	nextLabel := ""
	switch t.Status {
	case storage.Draft:
		nextStatus = "In Progress"
		nextLabel = "▶ Start"
	case storage.InProgress:
		nextStatus = "Done"
		nextLabel = "✓ Done"
	}

	html := fmt.Sprintf(`
	<div class="task-card %s">
		<div class="task-header">
			<span class="task-status-dot %s"></span>
			<span class="task-title">%s</span>
		</div>`,
		statusClass, statusClass, t.Title)

	if t.Description != "" {
		html += fmt.Sprintf(`<div class="task-desc">%s</div>`, t.Description)
	}

	html += `<div class="task-actions">`

	if nextStatus != "" {
		html += fmt.Sprintf(`
			<button class="btn btn-action btn-next"
				hx-post="/tasks/status?id=%s"
				hx-vals='{"status": "%s"}'
				hx-target="#main-area"
				hx-swap="outerHTML">%s</button>`,
			t.ID, nextStatus, nextLabel)
	}

	if isDraft || t.Status == storage.Done {
		html += fmt.Sprintf(`
			<button class="btn btn-action btn-delete"
				hx-delete="/tasks/delete?id=%s"
				hx-target="#main-area"
				hx-swap="outerHTML"
				hx-confirm="Delete this task?">✕</button>`,
			t.ID)
	}

	html += `
		</div>
	</div>`
	return html
}

// renderMainArea returns the full interactive area (form + panels) wrapped in #main-area.
// All HTMX operations target #main-area with outerHTML swap.
func renderMainArea(tasks []storage.Task, draftCount int) string {
	draftHTML := renderDraftPanel(tasks)
	progressHTML := renderInProgressPanel(tasks)
	doneHTML := renderDonePanel(tasks)

	return fmt.Sprintf(`<span id="draft-count" class="draft-count" hx-swap-oob="innerHTML:#draft-count">%d in draft</span>
		<div id="main-area">
			<form class="add-task-row"
				  hx-post="/tasks/create"
				  hx-target="#main-area"
				  hx-swap="outerHTML"
				  hx-on::after-request="this.querySelector('input[name=title]').value='';this.querySelector('textarea').value=''">
				<input type="text" name="title" placeholder="Task title" autocomplete="off" required>
				<textarea name="description" placeholder="Description (optional)"></textarea>
				<div class="form-actions">
					<button type="submit" class="btn btn-primary">+ Add Task</button>
				</div>
			</form>
			<div class="panels">
				<div class="panel panel-draft">
					%s
				</div>
				<div class="right-stack">
					<div class="panel panel-progress">
						%s
					</div>
					<div class="panel panel-done">
						%s
					</div>
				</div>
			</div>
		</div>`, draftCount, draftHTML, progressHTML, doneHTML)
}

type stringWriter struct {
	data []byte
}

func (w *stringWriter) Write(p []byte) (int, error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

func (w *stringWriter) String() string {
	return string(w.data)
}

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Task Service</title>
	<script src="https://unpkg.com/htmx.org@2.0.4"></script>
	<style>
		/* ===== Reset & Base ===== */
		*, *::before, *::after {
			box-sizing: border-box;
			margin: 0;
			padding: 0;
		}

		html, body {
			height: 100%;
		}

		body {
			font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			background: linear-gradient(135deg, #0f0a1a 0%, #1a1030 25%, #2b1e3e 50%, #1a1030 75%, #0f0a1a 100%);
			background-attachment: fixed;
			color: #e6e6fa;
			position: relative;
			overflow-x: hidden;
		}

		/* Noise overlay */
		body::before {
			content: '';
			position: fixed;
			top: 0; left: 0; right: 0; bottom: 0;
			background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)' opacity='0.03'/%3E%3C/svg%3E");
			pointer-events: none;
			z-index: 0;
		}

		/* Ambient glow */
		body::after {
			content: '';
			position: fixed;
			top: -50%; left: -50%;
			width: 200%; height: 200%;
			background: radial-gradient(ellipse at 30% 20%, rgba(75, 70, 143, 0.12) 0%, transparent 50%),
						radial-gradient(ellipse at 70% 80%, rgba(164, 144, 194, 0.08) 0%, transparent 50%);
			pointer-events: none;
			z-index: 0;
		}

		/* ===== Layout ===== */
		.container {
			position: relative;
			z-index: 1;
			max-width: 960px;
			margin: 0 auto;
			padding: 24px 16px 16px;
			display: flex;
			flex-direction: column;
			height: 100vh;
		}

		/* ===== Header ===== */
		.page-header {
			text-align: center;
			flex-shrink: 0;
		}

		.page-header h1 {
			font-size: 1.75rem;
			font-weight: 700;
			background: linear-gradient(135deg, #e6e6fa 0%, #a490c2 50%, #e6e6fa 100%);
			-webkit-background-clip: text;
			-webkit-text-fill-color: transparent;
			background-clip: text;
			letter-spacing: 0.02em;
		}

		.page-header .draft-count {
			color: #a490c2;
			-webkit-text-fill-color: #a490c2;
		}

		.divider {
			height: 1px;
			background: linear-gradient(90deg, transparent 0%, #4a4e8f 50%, transparent 100%);
			margin: 12px 0;
			flex-shrink: 0;
		}

		/* ===== Add Task Form ===== */
		.add-task-row {
			display: flex;
			flex-direction: column;
			gap: 8px;
			flex-shrink: 0;
		}

		.add-task-row input[type="text"],
		.add-task-row textarea {
			width: 100%;
			padding: 10px 14px;
			border: 1px solid rgba(164, 144, 194, 0.25);
			border-radius: 10px;
			background: rgba(43, 30, 62, 0.6);
			color: #e6e6fa;
			font-size: 0.88rem;
			font-family: inherit;
			outline: none;
			transition: border-color 0.2s, box-shadow 0.2s;
			backdrop-filter: blur(8px);
			resize: none;
		}

		.add-task-row input[type="text"] {
			padding: 12px 16px;
			font-size: 0.95rem;
		}

		.add-task-row textarea {
			min-height: 64px;
			max-height: 160px;
			margin-top: 4px;
		}

		.add-task-row input::placeholder,
		.add-task-row textarea::placeholder {
			color: rgba(164, 144, 194, 0.5);
		}

		.add-task-row input:focus,
		.add-task-row textarea:focus {
			border-color: #a490c2;
			box-shadow: 0 0 0 3px rgba(164, 144, 194, 0.15);
		}

		.add-task-row .form-actions {
			display: flex;
			justify-content: flex-end;
		}

		/* ===== Buttons ===== */
		.btn {
			border: none;
			border-radius: 10px;
			cursor: pointer;
			font-size: 0.9rem;
			font-weight: 600;
			transition: all 0.2s;
		}

		.btn-primary {
			padding: 10px 20px;
			background: #3d3266;
			color: #c4b8d9;
			border: 1px solid rgba(164, 144, 194, 0.18);
		}

		.btn-primary:hover {
			background: #4a3f75;
			border-color: rgba(164, 144, 194, 0.3);
		}

		.btn-primary:active {
			background: #352b59;
		}

		.btn-action {
			padding: 5px 12px;
			font-size: 0.78rem;
			border-radius: 6px;
		}

		.btn-next {
			background: rgba(74, 78, 143, 0.5);
			color: #a490c2;
			border: 1px solid rgba(164, 144, 194, 0.2);
		}

		.btn-next:hover {
			background: rgba(74, 78, 143, 0.7);
			color: #e6e6fa;
			border-color: rgba(164, 144, 194, 0.4);
		}

		.btn-delete {
			background: rgba(180, 60, 60, 0.3);
			color: #e08888;
			border: 1px solid rgba(180, 60, 60, 0.25);
		}

		.btn-delete:hover {
			background: rgba(180, 60, 60, 0.5);
			color: #ffaaaa;
		}

		/* ===== Panels ===== */
		#main-area {
			flex: 1;
			min-height: 0;
			display: flex;
			flex-direction: column;
			gap: 8px;
		}

		.panels {
			display: flex;
			gap: 8px;
			flex: 1;
			min-height: 0;
		}

		.right-stack {
			flex: 1;
			display: flex;
			flex-direction: column;
			gap: 8px;
			min-height: 0;
		}

		.panel {
			background: rgba(43, 30, 62, 0.45);
			border: 1px solid rgba(164, 144, 194, 0.12);
			border-radius: 14px;
			padding: 16px;
			backdrop-filter: blur(12px);
			overflow-y: auto;
		}

		.panel-draft {
			flex: 1;
			min-height: 0;
		}

		.panel-progress,
		.panel-done {
			flex: 1;
			min-height: 0;
		}

		/* ===== Panel Title ===== */
		.panel-title {
			font-size: 0.75rem;
			text-transform: uppercase;
			letter-spacing: 0.08em;
			color: #a490c2;
			margin-bottom: 10px;
			display: flex;
			align-items: center;
			gap: 8px;
		}

		.count-badge {
			background: rgba(164, 144, 194, 0.15);
			color: #a490c2;
			font-size: 0.7rem;
			padding: 1px 8px;
			border-radius: 10px;
			font-weight: 600;
		}

		/* ===== Task Card ===== */
		.task-card {
			background: rgba(74, 78, 143, 0.12);
			border: 1px solid rgba(164, 144, 194, 0.08);
			border-radius: 10px;
			padding: 10px 12px;
			margin-bottom: 8px;
			transition: all 0.2s;
		}

		.task-card:last-child {
			margin-bottom: 0;
		}

		.task-card:hover {
			background: rgba(74, 78, 143, 0.2);
			border-color: rgba(164, 144, 194, 0.18);
		}

		.task-card.status-done {
			opacity: 0.6;
		}

		.task-card.status-done .task-title {
			text-decoration: line-through;
			color: #a490c2;
		}

		.task-header {
			display: flex;
			align-items: center;
			gap: 8px;
			margin-bottom: 4px;
		}

		.task-status-dot {
			width: 8px;
			height: 8px;
			border-radius: 50%;
			flex-shrink: 0;
		}

		.task-status-dot.status-draft {
			background: #a490c2;
			box-shadow: 0 0 6px rgba(164, 144, 194, 0.5);
		}

		.task-status-dot.status-progress {
			background: #7b8ef5;
			box-shadow: 0 0 6px rgba(123, 142, 245, 0.5);
		}

		.task-status-dot.status-done {
			background: #6bcf8f;
			box-shadow: 0 0 6px rgba(107, 207, 143, 0.5);
		}

		.task-title {
			font-size: 0.9rem;
			font-weight: 500;
			color: #e6e6fa;
			flex: 1;
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
		}

		.task-desc {
			font-size: 0.78rem;
			color: rgba(164, 144, 194, 0.6);
			margin-bottom: 6px;
			padding-left: 16px;
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
		}

		.task-actions {
			display: flex;
			gap: 6px;
			justify-content: flex-end;
		}

		/* ===== Empty State ===== */
		.empty-state {
			text-align: center;
			color: rgba(164, 144, 194, 0.35);
			font-size: 0.82rem;
			padding: 16px 0;
			font-style: italic;
		}

		/* ===== Scrollbar ===== */
		::-webkit-scrollbar {
			width: 6px;
		}
		::-webkit-scrollbar-track {
			background: transparent;
		}
		::-webkit-scrollbar-thumb {
			background: rgba(164, 144, 194, 0.2);
			border-radius: 3px;
		}
		::-webkit-scrollbar-thumb:hover {
			background: rgba(164, 144, 194, 0.35);
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="page-header">
			<h1>Task Service: <span id="draft-count" class="draft-count">{{.DraftCount}} in draft</span></h1>
		</div>
		<div class="divider"></div>

		<div id="main-area">
			<form class="add-task-row"
				  hx-post="/tasks/create"
				  hx-target="#main-area"
				  hx-swap="outerHTML"
				  hx-on::after-request="this.querySelector('input[name=title]').value='';this.querySelector('textarea').value=''">
				<input type="text" name="title" placeholder="Task title" autocomplete="off" autofocus required>
				<textarea name="description" placeholder="Description (optional)"></textarea>
				<div class="form-actions">
					<button type="submit" class="btn btn-primary">+ Add Task</button>
				</div>
			</form>
			<div class="panels">
				<div class="panel panel-draft">
					{{.DraftPanel}}
				</div>
				<div class="right-stack">
					<div class="panel panel-progress">
						{{.ProgressPanel}}
					</div>
					<div class="panel panel-done">
						{{.DonePanel}}
					</div>
				</div>
			</div>
		</div>
	</div>
</body>
</html>`
