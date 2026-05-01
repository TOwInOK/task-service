package http

import (
	"fmt"
	"html/template"
	"time"

	"task-service/internal/storage"
)

func renderIndex(tasks []storage.Task, draftCount int) string {
	draftHTML := renderDraftPanel(tasks)
	progressHTML := renderInProgressPanel(tasks)
	doneHTML := renderDonePanel(tasks)

	var totalCount, progressCount, doneCount int
	for _, t := range tasks {
		totalCount++
		switch t.Status {
		case storage.Draft:
			// counted in draftCount
		case storage.InProgress:
			progressCount++
		case storage.Done:
			doneCount++
		}
	}

	tmpl := template.Must(template.New("index").Parse(indexHTML))

	data := struct {
		DraftCount    int
		TotalCount    int
		ProgressCount int
		DoneCount     int
		DraftPanel    template.HTML
		ProgressPanel template.HTML
		DonePanel     template.HTML
	}{
		DraftCount:    draftCount,
		TotalCount:    totalCount,
		ProgressCount: progressCount,
		DoneCount:     doneCount,
		DraftPanel:    template.HTML(draftHTML),
		ProgressPanel: template.HTML(progressHTML),
		DonePanel:     template.HTML(doneHTML),
	}

	var buf stringWriter
	if err := tmpl.Execute(&buf, data); err != nil {
		// Template execution error — return minimal error page
		return "<html><body><h1>Render Error</h1></body></html>"
	}
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
		html += renderTaskCard(t)
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
		html += renderTaskCard(t)
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
		html += renderTaskCard(t)
	}
	if len(done) == 0 {
		html += `<div class="empty-state">No completed tasks</div>`
	}
	return html
}

func renderTaskCard(t storage.Task) string {
	statusClass := "status-draft"
	statusLabel := "Draft"
	switch t.Status {
	case storage.InProgress:
		statusClass = "status-progress"
		statusLabel = "In Progress"
	case storage.Done:
		statusClass = "status-done"
		statusLabel = "Done"
	}

	escapedTitle := template.HTMLEscapeString(t.Title)

	html := fmt.Sprintf(`
	<div class="task-card %s">
		<div class="task-header">
			<span class="task-status-dot %s"></span>
			<span class="task-title">%s</span>
			<span class="task-status-label status-label-%s">%s</span>
		</div>`,
		statusClass, statusClass, escapedTitle, statusClass, statusLabel)

	if t.Description != "" {
		escapedDesc := template.HTMLEscapeString(t.Description)
		html += fmt.Sprintf(`<div class="task-desc">%s</div>`, escapedDesc)
	}

	// Timer display
	if t.Status == storage.InProgress && t.StartedAt != nil {
		// Live timer: JS will update this every second
		totalSecs := t.TimeSpent + int64(time.Since(*t.StartedAt).Seconds())
		html += fmt.Sprintf(`<div class="task-timer" data-started="%s" data-accumulated="%d">⏱ %s</div>`,
			t.StartedAt.Format(time.RFC3339),
			t.TimeSpent,
			storage.FormatDuration(totalSecs))
	} else if t.Status == storage.Done && t.TimeSpent > 0 {
		html += fmt.Sprintf(`<div class="task-time">⏱ %s</div>`,
			storage.FormatDuration(t.TimeSpent))
	}

	html += `<div class="task-actions">`

	// Status transition buttons based on current status
	switch t.Status {
	case storage.Draft:
		html += fmt.Sprintf(`
			<button class="btn btn-action btn-next"
				hx-post="/tasks/status?id=%s"
				hx-vals='{"status": "In Progress"}'
				hx-target="#main-area"
				hx-swap="outerHTML">▶ Start</button>`,
			t.ID)
	case storage.InProgress:
		html += fmt.Sprintf(`
			<button class="btn btn-action btn-next"
				hx-post="/tasks/status?id=%s"
				hx-vals='{"status": "Done"}'
				hx-target="#main-area"
				hx-swap="outerHTML">✓ Complete</button>`,
			t.ID)
		html += fmt.Sprintf(`
			<button class="btn btn-action btn-next"
				hx-post="/tasks/status?id=%s"
				hx-vals='{"status": "Draft"}'
				hx-target="#main-area"
				hx-swap="outerHTML">↩ Draft</button>`,
			t.ID)
	case storage.Done:
		html += fmt.Sprintf(`
			<button class="btn btn-action btn-next"
				hx-post="/tasks/status?id=%s"
				hx-vals='{"status": "In Progress"}'
				hx-target="#main-area"
				hx-swap="outerHTML">↺ Reopen</button>`,
			t.ID)
	}

	// Delete button on ALL tasks
	html += fmt.Sprintf(`
			<button class="btn btn-action btn-delete"
				hx-delete="/tasks/delete?id=%s"
				hx-target="#main-area"
				hx-swap="outerHTML"
				hx-confirm="Delete this task?">✕</button>`,
		t.ID)

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

	var total, active, done int
	for _, t := range tasks {
		total++
		switch t.Status {
		case storage.InProgress:
			active++
		case storage.Done:
			done++
		}
	}

	return fmt.Sprintf(`<span id="draft-count" class="draft-count" hx-swap-oob="innerHTML:#draft-count">%d tasks · %d draft · %d active · %d done</span>
		<div id="main-area">
			<form class="add-task-row"
				  hx-post="/tasks/create"
				  hx-target="#main-area"
				  hx-swap="outerHTML"
				  hx-indicator="#htmx-indicator"
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
		</div>`, total, draftCount, active, done, draftHTML, progressHTML, doneHTML)
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
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Playwrite+DE+SAS:wght@100..400&display=swap" rel="stylesheet">
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
    <style>
        /* ===== RESET & BASE ===== */
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

        html {
            font-size: 15px;
            -webkit-font-smoothing: antialiased;
            -moz-osx-font-smoothing: grayscale;
        }

        body {
            font-family: 'Inter', Arial, system-ui, -apple-system, sans-serif;
            background: linear-gradient(135deg, #0f0a1a 0%, #1a1030 25%, #2b1e3e 50%, #1a1030 75%, #0f0a1a 100%);
            background-attachment: fixed;
            color: #e6e6fa;
            min-height: 100vh;
            line-height: 1.5;
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

        /* ===== SCROLLBAR ===== */
        ::-webkit-scrollbar { width: 8px; height: 8px; }
        ::-webkit-scrollbar-track { background: transparent; }
        ::-webkit-scrollbar-thumb { background: rgba(164, 144, 194, 0.15); border-radius: 4px; }
        ::-webkit-scrollbar-thumb:hover { background: rgba(164, 144, 194, 0.25); }

        /* ===== CONTAINER ===== */
        .container {
            max-width: 1100px;
            margin: 0 auto;
            padding: 2rem 1.5rem;
            display: flex;
            flex-direction: column;
            min-height: 100vh;
            position: relative;
            z-index: 1;
        }

        /* ===== PAGE HEADER ===== */
        .page-header {
            text-align: center;
            padding-bottom: 1rem;
            position: relative;
        }

        .page-header h1 {
            font-family: 'Playwrite DE SAS', Georgia, 'Times New Roman', serif;
            font-size: 1.75rem;
            font-weight: 500;
            background: linear-gradient(135deg, #e6e6fa, #a490c2, #e6e6fa);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            letter-spacing: -0.02em;
        }

        .page-header .draft-count {
            display: block;
            font-size: 0.85rem;
            font-weight: 400;
            color: #a490c2;
            margin-top: 0.25rem;
        }

        /* ===== HTMX INDICATOR ===== */
        .htmx-indicator {
            display: none;
            position: absolute;
            top: 0;
            right: 0;
            width: 22px;
            height: 22px;
            border: 2.5px solid rgba(164, 144, 194, 0.25);
            border-top-color: rgba(164, 144, 194, 0.6);
            border-radius: 50%;
            animation: spin 0.6s linear infinite;
        }

        .htmx-indicator.htmx-request {
            display: inline-block;
        }

        @keyframes spin {
            to { transform: rotate(360deg); }
        }

        /* ===== DIVIDER ===== */
        .divider {
            border: none;
            height: 1px;
            background: linear-gradient(90deg, transparent, #4a4e8f, transparent);
            margin: 1rem 0;
        }

        /* ===== MAIN AREA ===== */
        #main-area {
            display: flex;
            flex-direction: column;
            flex: 1;
            gap: 1.25rem;
        }

        /* ===== ADD TASK FORM ===== */
        .add-task-row {
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
            background: rgba(43, 30, 62, 0.45);
            backdrop-filter: blur(12px);
            border: 1px solid rgba(164, 144, 194, 0.12);
            border-radius: 10px;
            padding: 1rem;
        }

        .add-task-row input[type="text"] {
            width: 100%;
            padding: 0.6rem 0.85rem;
            font-size: 0.95rem;
            font-family: inherit;
            color: #e6e6fa;
            background: rgba(43, 30, 62, 0.6);
            backdrop-filter: blur(8px);
            border: 1px solid rgba(164, 144, 194, 0.20);
            border-radius: 7px;
            outline: none;
            transition: border-color 0.15s, box-shadow 0.15s;
        }

        .add-task-row textarea {
            width: 100%;
            padding: 0.5rem 0.85rem;
            font-size: 0.9rem;
            font-family: inherit;
            color: #e6e6fa;
            background: rgba(43, 30, 62, 0.6);
            backdrop-filter: blur(8px);
            border: 1px solid rgba(164, 144, 194, 0.20);
            border-radius: 7px;
            outline: none;
            resize: vertical;
            min-height: 48px;
            max-height: 120px;
            transition: border-color 0.15s, box-shadow 0.15s;
        }

        .add-task-row input::placeholder,
        .add-task-row textarea::placeholder {
            color: rgba(164, 144, 194, 0.5);
        }

        .add-task-row input:focus,
        .add-task-row textarea:focus {
            border-color: #3898ec;
            box-shadow: 0 0 0 3px rgba(56, 152, 236, 0.15);
        }

        .form-actions {
            display: flex;
            gap: 0.5rem;
        }

        /* ===== PANELS LAYOUT ===== */
        .panels {
            display: flex;
            gap: 1rem;
            flex: 1;
        }

        .right-stack {
            display: flex;
            flex-direction: column;
            gap: 1rem;
            flex: 1;
            min-width: 0;
        }

        /* ===== PANEL ===== */
        .panel {
            background: rgba(43, 30, 62, 0.45);
            backdrop-filter: blur(12px);
            border: 1px solid rgba(164, 144, 194, 0.12);
            border-radius: 10px;
            padding: 1rem;
            display: flex;
            flex-direction: column;
            gap: 0.75rem;
            min-width: 0;
            max-height: calc(100vh - 240px);
            overflow-y: auto;
        }

        .panel-draft {
            flex: 0 0 320px;
            border-top: 3px solid #a490c2;
        }

        .panel-progress {
            flex: 1;
            border-top: 3px solid #7b8ef5;
        }

        .panel-done {
            flex: 1;
            border-top: 3px solid #6bcf8f;
        }

        .panel-title {
            font-family: 'Inter', Arial, system-ui, sans-serif;
            font-size: 0.75rem;
            font-weight: 500;
            text-transform: uppercase;
            letter-spacing: 0.08em;
            color: #a490c2;
            display: flex;
            align-items: center;
            gap: 0.5rem;
            padding-bottom: 0.25rem;
        }

        .count-badge {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            min-width: 20px;
            height: 20px;
            padding: 0 6px;
            font-size: 0.7rem;
            font-weight: 600;
            border-radius: 10px;
            background: rgba(164, 144, 194, 0.15);
            color: #a490c2;
            line-height: 1;
        }

        /* ===== TASK CARDS ===== */
        .task-card {
            background: rgba(74, 78, 143, 0.12);
            border: 1px solid rgba(164, 144, 194, 0.12);
            border-radius: 8px;
            padding: 0.75rem;
            display: flex;
            flex-direction: column;
            gap: 0.4rem;
            transition: background 0.15s, box-shadow 0.15s, border-color 0.15s;
        }

        .task-card:hover {
            background: rgba(74, 78, 143, 0.2);
            box-shadow: rgba(74, 78, 143, 0.2) 0px 0px 0px 0px, rgba(164, 144, 194, 0.18) 0px 0px 0px 1px;
            border-color: transparent;
        }

        .task-card.status-done {
            opacity: 0.65;
        }

        /* ===== TASK HEADER ===== */
        .task-header {
            display: flex;
            align-items: center;
            gap: 0.45rem;
            min-width: 0;
        }

        .task-status-dot {
            width: 9px;
            height: 9px;
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
            font-family: 'Playwrite DE SAS', Georgia, 'Times New Roman', serif;
            font-size: 1.06rem;
            font-weight: 500;
            color: #e6e6fa;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            flex: 1;
            min-width: 0;
        }

        /* ===== STATUS LABEL (pill) ===== */
        .task-status-label {
            display: inline-flex;
            align-items: center;
            font-size: 0.65rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.04em;
            padding: 1px 7px;
            border-radius: 10px;
            flex-shrink: 0;
            line-height: 1.6;
        }

        .status-label-draft {
            background: rgba(164, 144, 194, 0.15);
            color: #a490c2;
            border: 1px solid rgba(164, 144, 194, 0.25);
        }

        .status-label-progress {
            background: rgba(123, 142, 245, 0.15);
            color: #7b8ef5;
            border: 1px solid rgba(123, 142, 245, 0.25);
        }

        .status-label-done {
            background: rgba(107, 207, 143, 0.15);
            color: #6bcf8f;
            border: 1px solid rgba(107, 207, 143, 0.25);
        }

        /* ===== TASK DESCRIPTION ===== */
        .task-desc {
            font-size: 0.94rem;
            color: rgba(164, 144, 194, 0.6);
            line-height: 1.6;
            padding-left: 1.2rem;
            margin-bottom: 8px;
            word-break: break-word;
            overflow-wrap: break-word;
        }

        /* ===== TASK TIMER ===== */
        .task-timer,
        .task-time {
            font-size: 0.88rem;
            font-family: 'Inter', Arial, system-ui, sans-serif;
            font-variant-numeric: tabular-nums;
            color: #7b8ef5;
            padding-left: 1.2rem;
        }

        .task-time {
            color: rgba(107, 207, 143, 0.7);
        }

        /* ===== TASK ACTIONS ===== */
        .task-actions {
            display: flex;
            gap: 0.35rem;
            padding-top: 0.25rem;
            flex-wrap: wrap;
        }

        /* ===== BUTTONS ===== */
        .btn {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            font-family: inherit;
            font-size: 0.78rem;
            font-weight: 550;
            line-height: 1;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            padding: 0.4rem 0.7rem;
            transition: background 0.15s, color 0.15s, box-shadow 0.15s, opacity 0.15s;
            white-space: nowrap;
        }

        .btn-primary {
            background: #3d3266;
            color: #c4b8d9;
            padding: 0.55rem 1.1rem;
            font-size: 0.85rem;
            font-weight: 600;
            border-radius: 12px;
            box-shadow: #3d3266 0px 0px 0px 0px, rgba(164, 144, 194, 0.25) 0px 0px 0px 1px;
        }

        .btn-primary:hover {
            background: #4a3f75;
            box-shadow: #4a3f75 0px 0px 0px 0px, rgba(164, 144, 194, 0.35) 0px 0px 0px 1px;
        }

        .btn-primary:active {
            background: #352b59;
            box-shadow: inset 0px 0px 0px 1px rgba(164, 144, 194, 0.15);
        }

        .btn-action {
            background: transparent;
            border: 1px solid rgba(164, 144, 194, 0.20);
            color: #a490c2;
            padding: 0.3rem 0.6rem;
            font-size: 0.72rem;
        }

        .btn-action:hover {
            background: rgba(74, 78, 143, 0.2);
            border-color: rgba(164, 144, 194, 0.18);
        }

        .btn-next {
            color: #a490c2;
            border-color: rgba(164, 144, 194, 0.20);
        }

        .btn-next:hover {
            background: rgba(74, 78, 143, 0.7);
            color: #e6e6fa;
        }

        .btn-delete {
            color: #e08888;
            border-color: rgba(180, 60, 60, 0.25);
        }

        .btn-delete:hover {
            background: rgba(180, 60, 60, 0.5);
            color: #ffaaaa;
            border-color: rgba(180, 60, 60, 0.35);
        }

        /* ===== EMPTY STATE ===== */
        .empty-state {
            text-align: center;
            font-family: 'Playwrite DE SAS', Georgia, 'Times New Roman', serif;
            font-style: italic;
            color: rgba(164, 144, 194, 0.35);
            font-size: 1rem;
            font-weight: 400;
            padding: 1.5rem 0;
        }

        /* ===== RESPONSIVE ===== */
        @media (max-width: 680px) {
            .container {
                padding: 1rem;
            }

            .panels {
                flex-direction: column;
            }

            .panel-draft {
                flex: none;
            }

            .page-header h1 {
                font-size: 1.6rem;
                line-height: 1.10;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="page-header">
            <h1>Task Service <span id="draft-count" class="draft-count">{{.TotalCount}} tasks · {{.DraftCount}} draft · {{.ProgressCount}} active · {{.DoneCount}} done</span></h1>
            <div id="htmx-indicator" class="htmx-indicator"></div>
        </div>
        <div class="divider"></div>

        <div id="main-area">
            <form class="add-task-row"
                  hx-post="/tasks/create"
                  hx-target="#main-area"
                  hx-swap="outerHTML"
                  hx-indicator="#htmx-indicator"
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
    <script>
        function fmtDur(s) {
            if (s < 60) return s + 's';
            if (s < 3600) return Math.floor(s/60) + 'm ' + (s%60) + 's';
            if (s < 86400) return Math.floor(s/3600) + 'h ' + Math.floor((s%3600)/60) + 'm';
            return Math.floor(s/86400) + 'd ' + Math.floor((s%86400)/3600) + 'h';
        }
        function updateTimers() {
            document.querySelectorAll('.task-timer[data-started]').forEach(function(el) {
                var started = new Date(el.dataset.started);
                var acc = parseInt(el.dataset.accumulated) || 0;
                var elapsed = acc + Math.floor((Date.now() - started.getTime()) / 1000);
                el.textContent = '\u23F1 ' + fmtDur(elapsed);
            });
        }
        updateTimers();
        setInterval(updateTimers, 1000);
    </script>
</body>
</html>`
