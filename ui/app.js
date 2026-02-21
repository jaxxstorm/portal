const state = {
  requests: [],
  stats: null,
  health: null,
  filter: "",
  selectedId: null,
  requestTab: "summary",
  responseTab: "summary",
  bootedAt: Date.now(),
  lastUpdatedAt: null
}

const apiBasePath = resolveAPIBasePath()

document.addEventListener("DOMContentLoaded", init)

function init() {
  wireNavigation()
  wireInspectControls()
  wireTabs("request-tabs", (tab) => {
    state.requestTab = tab
    renderDetail()
  })
  wireTabs("response-tabs", (tab) => {
    state.responseTab = tab
    renderDetail()
  })

  poll()
  setInterval(poll, 1000)
}

function wireNavigation() {
  const buttons = document.querySelectorAll(".nav-btn")
  buttons.forEach((button) => {
    button.addEventListener("click", () => {
      buttons.forEach((b) => b.classList.remove("active"))
      button.classList.add("active")
      const view = button.dataset.view
      document.querySelectorAll(".view").forEach((node) => node.classList.remove("active"))
      document.getElementById(view).classList.add("active")
    })
  })
}

function wireInspectControls() {
  document.getElementById("request-filter").addEventListener("input", (event) => {
    state.filter = event.target.value || ""
    renderRequestList()
  })

  document.getElementById("clear-requests").addEventListener("click", async () => {
    try {
      const response = await fetch(apiURL("requests"), { method: "DELETE" })
      if (!response.ok && response.status !== 204) {
        throw new Error("clear failed")
      }
      state.requests = []
      state.selectedId = null
      state.lastUpdatedAt = Date.now()
      render()
    } catch (_error) {
      // Keep the UI responsive even when clear fails.
    }
  })
}

function wireTabs(containerId, onSelect) {
  const container = document.getElementById(containerId)
  if (!container) {
    return
  }
  container.querySelectorAll(".tab-btn").forEach((btn) => {
    btn.addEventListener("click", () => {
      container.querySelectorAll(".tab-btn").forEach((node) => node.classList.remove("active"))
      btn.classList.add("active")
      onSelect(btn.dataset.tab)
    })
  })
}

async function poll() {
  try {
    const [requests, stats, health] = await Promise.all([
      fetchJSON(apiURL("requests")),
      fetchJSON(apiURL("stats")),
      fetchJSON(apiURL("health"))
    ])

    state.requests = (Array.isArray(requests) ? requests : []).slice().reverse()
    state.stats = stats || {}
    state.health = health || {}
    state.lastUpdatedAt = Date.now()

    const hasCurrentSelection = state.requests.some((request) => request.id === state.selectedId)
    if (!hasCurrentSelection) {
      state.selectedId = state.requests.length > 0 ? state.requests[0].id : null
    }

    setOnline(true)
    render()
  } catch (_error) {
    setOnline(false)
  }
}

async function fetchJSON(url) {
  const response = await fetch(url)
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }
  return response.json()
}

function render() {
  renderTopMeta()
  renderKpis()
  renderRequestList()
  renderDetail()
  renderStatusView()
}

function renderTopMeta() {
  const target = document.getElementById("last-updated")
  if (!state.lastUpdatedAt) {
    target.textContent = "waiting for traffic..."
    return
  }
  target.textContent = `updated ${timeAgo(state.lastUpdatedAt)}`
}

function renderKpis() {
  const stats = state.stats || {}
  const derived = deriveMetrics(state.requests)

  document.getElementById("kpi-total").textContent = String(derived.totalRequests)
  document.getElementById("kpi-errors").textContent = `${formatPercent(derived.errorRate)}%`
  document.getElementById("kpi-p50").textContent = `${formatMs(stats.p50_response_time)} ms`
  document.getElementById("kpi-p90").textContent = `${formatMs(stats.p90_response_time)} ms`
}

function renderRequestList() {
  const container = document.getElementById("request-list")
  const requests = filteredRequests()
  if (requests.length === 0) {
    container.innerHTML = `<div class="empty-state">No requests match the current filter.</div>`
    return
  }

  container.innerHTML = requests.map((request) => {
    const isActive = request.id === state.selectedId ? "active" : ""
    const statusCode = Number(request.status_code || request.response?.status_code || 0)
    const statusClass = statusCode >= 400 ? "status-err" : "status-ok"
    const durationMs = nsToMs(request.duration)
    return `
      <article class="request-row ${isActive}" data-id="${escapeHtml(request.id)}">
        <span class="method-badge">${escapeHtml(request.method || "-")}</span>
        <div class="request-path">${escapeHtml(request.url || "/")}</div>
        <div class="status-pill ${statusClass}">${escapeHtml(String(statusCode || "-"))}</div>
        <div class="request-meta">${formatMs(durationMs)} ms</div>
      </article>
    `
  }).join("")

  container.querySelectorAll(".request-row").forEach((row) => {
    row.addEventListener("click", () => {
      state.selectedId = row.dataset.id
      renderRequestList()
      renderDetail()
    })
  })
}

function renderDetail() {
  const selected = currentSelectedRequest()
  const emptyNode = document.getElementById("empty-detail")
  const detailNode = document.getElementById("detail-content")

  if (!selected) {
    document.getElementById("selected-title").textContent = "Select a request"
    document.getElementById("selected-meta").textContent = ""
    emptyNode.classList.remove("hidden")
    detailNode.classList.add("hidden")
    return
  }

  emptyNode.classList.add("hidden")
  detailNode.classList.remove("hidden")

  const statusCode = Number(selected.status_code || selected.response?.status_code || 0)
  document.getElementById("selected-title").textContent = `${selected.method || "-"} ${selected.url || "/"}`
  document.getElementById("selected-meta").textContent = [
    statusCode > 0 ? `status ${statusCode}` : "status n/a",
    `${formatMs(nsToMs(selected.duration))} ms`,
    selected.remote_addr || "remote n/a",
    formatAbsoluteTime(selected.timestamp)
  ].join(" • ")

  document.getElementById("request-tab-content").innerHTML = renderRequestTab(selected, state.requestTab)
  document.getElementById("response-tab-content").innerHTML = renderResponseTab(selected, state.responseTab)
}

function renderRequestTab(request, tab) {
  const requestHeaders = request.headers || {}
  switch (tab) {
    case "headers":
      return renderHeadersBlock(requestHeaders)
    case "raw":
      return `<pre class="mono-block">${escapeHtml(renderRawRequest(request))}</pre>`
    case "body":
      return `<pre class="mono-block">${escapeHtml(renderRequestBody(request))}</pre>`
    default:
      return renderSummaryGrid([
        ["ID", request.id || "-"],
        ["Method", request.method || "-"],
        ["URL", request.url || "-"],
        ["Remote", request.remote_addr || "-"],
        ["User-Agent", request.user_agent || "-"],
        ["Content-Type", request.content_type || "-"],
        ["Body Size", `${request.size || 0} bytes`]
      ])
  }
}

function renderResponseTab(request, tab) {
  const response = request.response || {}
  switch (tab) {
    case "headers":
      return renderHeadersBlock(response.headers || {})
    case "raw":
      return `<pre class="mono-block">${escapeHtml(renderRawResponse(request))}</pre>`
    case "body":
      return `<pre class="mono-block">${escapeHtml(renderResponseBody(response))}</pre>`
    default:
      return renderSummaryGrid([
        ["Status", String(response.status_code || request.status_code || "-")],
        ["Duration", `${formatMs(nsToMs(request.duration))} ms`],
        ["Response Size", `${response.size || 0} bytes`],
        ["Content-Type", response.headers?.["Content-Type"] || "-"],
        ["Body Captured", response.body ? "yes" : "no"],
        ["Body Truncated", response.body_truncated ? "yes" : "no"]
      ])
  }
}

function renderSummaryGrid(entries) {
  return `
    <dl class="summary-grid">
      ${entries.map(([k, v]) => `<div><dt>${escapeHtml(k)}</dt><dd>${escapeHtml(v)}</dd></div>`).join("")}
    </dl>
  `
}

function renderHeadersBlock(headers) {
  const entries = Object.entries(headers || {})
  if (entries.length === 0) {
    return `<pre class="mono-block">(no headers captured)</pre>`
  }
  const lines = entries
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([key, value]) => `${key}: ${value}`)
    .join("\n")
  return `<pre class="mono-block">${escapeHtml(lines)}</pre>`
}

function renderRawRequest(request) {
  const headerLines = Object.entries(request.headers || {})
    .map(([key, value]) => `${key}: ${value}`)
    .join("\n")
  const body = renderRequestBody(request)
  return `${request.method || "GET"} ${request.url || "/"} HTTP/1.1\n${headerLines}\n\n${body}`
}

function renderRawResponse(request) {
  const response = request.response || {}
  const statusCode = response.status_code || request.status_code || 0
  const headerLines = Object.entries(response.headers || {})
    .map(([key, value]) => `${key}: ${value}`)
    .join("\n")
  return `HTTP/1.1 ${statusCode}\n${headerLines}\n\n${renderResponseBody(response)}`
}

function renderRequestBody(request) {
  const body = typeof request.body === "string" ? request.body : ""
  return body === "" ? "(empty request body)" : body
}

function renderResponseBody(response) {
  const body = typeof response.body === "string" ? response.body : ""
  if (body === "") {
    return "(empty or non-captured response body)"
  }
  if (response.body_truncated) {
    return `${body}\n\n[response body truncated]`
  }
  return body
}

function renderStatusView() {
  const metrics = deriveMetrics(state.requests)
  const stats = state.stats || {}
  const health = state.health || {}

  const runtimeTable = document.getElementById("runtime-table")
  runtimeTable.innerHTML = [
    ["Health", health.status || "unknown"],
    ["Log Provider", String(Boolean(health.log_provider))],
    ["Request Count", String(metrics.totalRequests)],
    ["Last Request", metrics.lastRequestAt ? formatAbsoluteTime(metrics.lastRequestAt) : "n/a"],
    ["Uptime", formatUptime(Date.now() - state.bootedAt)]
  ].map(([k, v]) => `<tr><td>${escapeHtml(k)}</td><td>${escapeHtml(v)}</td></tr>`).join("")

  const metricsTable = document.getElementById("metrics-table")
  metricsTable.innerHTML = [
    ["Open Connections", String(stats.open_connections || 0)],
    ["Avg Latency 1m", `${formatMs(stats.avg_response_time_1m)} ms`],
    ["Avg Latency 5m", `${formatMs(stats.avg_response_time_5m)} ms`],
    ["P50 Latency", `${formatMs(stats.p50_response_time)} ms`],
    ["P90 Latency", `${formatMs(stats.p90_response_time)} ms`],
    ["Requests / 1m", String(metrics.requests1m)],
    ["Requests / 5m", String(metrics.requests5m)],
    ["Requests / 15m", String(metrics.requests15m)],
    ["Unique Clients", String(metrics.uniqueClients)],
    ["Error Rate", `${formatPercent(metrics.errorRate)}%`]
  ].map(([k, v]) => `<tr><td>${escapeHtml(k)}</td><td>${escapeHtml(v)}</td></tr>`).join("")

  document.getElementById("method-breakdown").innerHTML = renderBreakdown(metrics.methodCounts)
  document.getElementById("status-breakdown").innerHTML = renderBreakdown(metrics.statusCounts)
}

function renderBreakdown(counts) {
  const entries = Object.entries(counts || {}).sort((a, b) => b[1] - a[1])
  if (entries.length === 0) {
    return `<div class="empty-state">No data yet.</div>`
  }
  return entries.map(([name, value]) => {
    return `<div class="breakdown-row"><span>${escapeHtml(name)}</span><strong>${value}</strong></div>`
  }).join("")
}

function filteredRequests() {
  const query = state.filter.trim().toLowerCase()
  if (!query) {
    return state.requests
  }

  return state.requests.filter((request) => {
    const statusCode = String(request.status_code || request.response?.status_code || "")
    const haystack = [
      request.method || "",
      request.url || "",
      request.remote_addr || "",
      request.user_agent || "",
      statusCode
    ].join(" ").toLowerCase()
    return haystack.includes(query)
  })
}

function currentSelectedRequest() {
  if (!state.selectedId) {
    return null
  }
  return state.requests.find((request) => request.id === state.selectedId) || null
}

function deriveMetrics(requests) {
  const now = Date.now()
  let errorCount = 0
  let requests1m = 0
  let requests5m = 0
  let requests15m = 0
  const methodCounts = {}
  const statusCounts = {}
  const clientSet = new Set()

  requests.forEach((request) => {
    const statusCode = Number(request.status_code || request.response?.status_code || 0)
    if (statusCode >= 400) {
      errorCount++
    }

    const method = request.method || "UNKNOWN"
    const statusLabel = statusCode > 0 ? String(statusCode) : "unknown"
    methodCounts[method] = (methodCounts[method] || 0) + 1
    statusCounts[statusLabel] = (statusCounts[statusLabel] || 0) + 1

    if (request.remote_addr) {
      clientSet.add(request.remote_addr)
    }

    const timestamp = toMs(request.timestamp)
    if (!Number.isNaN(timestamp)) {
      const age = now - timestamp
      if (age <= 60_000) requests1m++
      if (age <= 300_000) requests5m++
      if (age <= 900_000) requests15m++
    }
  })

  const totalRequests = requests.length
  const errorRate = totalRequests === 0 ? 0 : (errorCount / totalRequests) * 100
  const lastRequestAt = totalRequests > 0 ? requests[0].timestamp : null

  return {
    totalRequests,
    errorCount,
    errorRate,
    requests1m,
    requests5m,
    requests15m,
    uniqueClients: clientSet.size,
    methodCounts,
    statusCounts,
    lastRequestAt
  }
}

function setOnline(isOnline) {
  const pill = document.getElementById("status-pill")
  pill.textContent = isOnline ? "online" : "degraded"
  pill.classList.toggle("online", isOnline)
}

function nsToMs(durationNs) {
  return Number(durationNs || 0) / 1_000_000
}

function formatMs(value) {
  const parsed = Number(value || 0)
  return Number.isFinite(parsed) ? parsed.toFixed(1) : "0.0"
}

function formatPercent(value) {
  const parsed = Number(value || 0)
  return Number.isFinite(parsed) ? parsed.toFixed(1) : "0.0"
}

function formatAbsoluteTime(input) {
  const timestamp = toMs(input)
  if (Number.isNaN(timestamp)) {
    return "n/a"
  }
  return new Date(timestamp).toLocaleTimeString()
}

function toMs(value) {
  if (typeof value === "number") {
    return value
  }
  if (!value) {
    return Number.NaN
  }
  const parsed = Date.parse(value)
  return Number.isNaN(parsed) ? Number.NaN : parsed
}

function timeAgo(timestampMs) {
  const delta = Math.max(0, Date.now() - Number(timestampMs))
  const seconds = Math.floor(delta / 1000)
  if (seconds < 2) return "just now"
  if (seconds < 60) return `${seconds}s ago`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  return `${hours}h ago`
}

function formatUptime(ms) {
  const totalSeconds = Math.max(0, Math.floor(ms / 1000))
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60
  if (hours > 0) {
    return `${hours}h ${minutes}m ${seconds}s`
  }
  if (minutes > 0) {
    return `${minutes}m ${seconds}s`
  }
  return `${seconds}s`
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;")
}

function resolveAPIBasePath() {
  const path = window.location.pathname || "/"
  if (path.startsWith("/ui/")) {
    return "/ui/api/"
  }
  if (path === "/ui") {
    return "/ui/api/"
  }
  return "/api/"
}

function apiURL(endpoint) {
  return `${apiBasePath}${endpoint}`
}
