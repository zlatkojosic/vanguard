const go = () => window.go.main.App;
const runtime = () => window.runtime;

let ui = { kind: "tabs", title: "N POS Z v16.2", items: [], fields: [], theme: "theme3.css", accent: "#ff0000" };
let kbTarget = null;

function call(name, ...args) {
  const api = go();
  if (!api || typeof api[name] !== "function") return Promise.resolve();
  return api[name](...args);
}

function collectForm() {
  const out = {};
  document.querySelectorAll("[data-field]").forEach((el) => {
    const name = el.getAttribute("data-field");
    if (el.type === "checkbox") out[name] = el.checked ? "true" : "false";
    else out[name] = el.value || "";
  });
  return out;
}

function grid(items, columns) {
  const cols = columns === 2 ? "cols-2" : "cols-4";
  return `<div class="grid ${cols}">${(items || []).map((it, i) => {
    const n = i < 9 ? `<span class="n">${i + 1}</span>` : "";
    const color = it.color ? `style="background:${it.color}"` : "";
    return `<button type="button" class="tile" data-id="${esc(it.id)}" ${color}>${n}${esc(it.title)}</button>`;
  }).join("")}</div>`;
}

function fieldsHTML(fields, keyboard) {
  return `<div class="form">${(fields || []).map((f) => {
    if (f.type === "check") {
      const on = f.value === "true" ? "checked" : "";
      return `<label class="check"><input type="checkbox" data-field="${esc(f.name)}" ${on} /> ${esc(f.label)}</label>`;
    }
    if (f.type === "select") {
      const opts = (f.options || []).map((o) => `<option ${o === f.value ? "selected" : ""}>${esc(o)}</option>`).join("");
      return `<label>${esc(f.label)}</label><select data-field="${esc(f.name)}">${opts}</select>`;
    }
    const t = f.type === "password" ? "password" : "text";
    const kb = keyboard ? `<button type="button" class="kb-open" data-for="${esc(f.name)}">⌨</button>` : "";
    return `<label>${esc(f.label)}</label><div class="row" style="display:flex;gap:6px"><input type="${t}" data-field="${esc(f.name)}" value="${esc(f.value || "")}" placeholder="${esc(f.placeholder || "")}" />${kb}</div>`;
  }).join("")}</div>`;
}

function esc(s) {
  return String(s || "").replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
}

function render(next) {
  if (!next) return;
  ui = next;
  const L = ui.labels || ui.Labels || {};
  document.documentElement.lang = ui.lang || "sr";
  document.documentElement.style.setProperty("--accent", ui.accent || "#ff0000");
  document.getElementById("stripe").style.background = ui.accent || "#ff0000";
  document.getElementById("title").textContent = ui.title || "N POS Z v16.2";
  document.getElementById("message").textContent = ui.message || "";
  const toast = document.getElementById("toast");
  toast.textContent = ui.toast || "";
  toast.classList.toggle("hidden", !ui.toast);
  document.getElementById("busy").classList.toggle("hidden", !ui.busy);
  document.getElementById("busy").textContent = L.busy || "Radim...";
  document.getElementById("exit").textContent = L.exit || "Izlaz";
  document.getElementById("home").textContent = L.home || "Početna";
  document.getElementById("back").textContent = L.back || "Nazad";
  document.querySelector("#modal h3").textContent = L.confirm || "Potvrda";
  document.getElementById("modal-yes").textContent = L.yes || "Da";
  document.getElementById("modal-no").textContent = L.no || "Ne";
  document.getElementById("kb-ok").textContent = L.ok || "OK";
  document.getElementById("kb-back").textContent = L.back || "Nazad";

  const modal = document.getElementById("modal");
  if (ui.confirm) {
    document.getElementById("modal-text").textContent = ui.confirm;
    modal.classList.remove("hidden");
  } else {
    modal.classList.add("hidden");
  }

  const main = document.getElementById("main");
  let html = "";
  if (ui.kind === "tabs") {
    const onLogin = !!ui.login;
    const left = onLogin ? "login" : "apps";
    const leftLabel = onLogin ? (L.login || "Prijava") : (L.apps || "Aplikacije");
    const activeLeft = ui.tab === "settings" ? "" : "active";
    const activeRight = ui.tab === "settings" ? "active" : "";
    html += `<div class="tabs"><button type="button" class="${activeLeft}" data-tab="${left}">${esc(leftLabel)}</button><button type="button" class="${activeRight}" data-tab="settings">${esc(L.settings || "Podešavanja")}</button></div>`;
    if (ui.tab === "settings") {
      html += grid(ui.items, 4);
    } else if (onLogin) {
      html += `<div class="card"><h2>${esc(L.login || "Prijava")}</h2><hr/>${fieldsHTML(ui.fields, ui.keyboard)}<div class="okbar"><button type="button" class="primary" id="ok">${esc(L.ok || "OK")}</button></div></div>`;
    } else {
      html += grid(ui.items, ui.columns || 4);
    }
  } else if (ui.kind === "menu") {
    html += (ui.info ? `<p>${esc(ui.info)}</p>` : "") + grid(ui.items, ui.columns || 4);
  } else if (ui.kind === "theme") {
    html += `<p>${esc(ui.info || "")}</p>` + (ui.items || []).map((it) =>
      `<button type="button" class="theme-btn" data-id="${esc(it.id)}" style="background:${it.color || "var(--accent)"}">${esc(it.title)}</button>`
    ).join("");
  } else if (ui.kind === "form") {
    html += (ui.info ? `<p>${esc(ui.info)}</p>` : "") + fieldsHTML(ui.fields, ui.keyboard);
    html += `<div class="okbar"><button type="button" class="primary" id="ok">${esc(L.ok || "OK")}</button></div>`;
  } else if (ui.kind === "report") {
    html += fieldsHTML(ui.fields, ui.keyboard);
    html += `<div class="report-actions"><button type="button" class="primary" data-id="reportShow">${esc(L.show || "Prikaži")}</button></div>`;
    html += tableHTML(ui.table, false);
    html += `<div class="summary">${esc(ui.summary || "")}</div>`;
    html += `<div class="report-actions"><button type="button" class="primary" data-id="reportPrint">${esc(L.print || "Štampaj")}</button></div>`;
  } else if (ui.kind === "table") {
    html += tableHTML(ui.table, true);
  } else if (ui.kind === "info") {
    html += `<div class="info">${esc(ui.info || "")}</div>`;
  }
  main.innerHTML = html;
}

function tableHTML(table, storno) {
  if (!table) return "";
  const heads = (table.heads || []).map((h) => `<th>${esc(h)}</th>`).join("");
  const rows = (table.rows || []).map((r, ri) => {
    const action = (table.actions && table.actions[ri]) || (storno && r[2] ? `xbonStorno:${r[2]}` : "");
    const cells = r.map((c, i) => {
      if (action && i === r.length - 1) {
        return `<td class="action" data-id="${esc(action)}">${esc(c)}</td>`;
      }
      return `<td>${esc(c)}</td>`;
    }).join("");
    return `<tr>${cells}</tr>`;
  }).join("");
  return `<table><thead><tr>${heads}</tr></thead><tbody>${rows}</tbody></table>`;
}

function submit() {
  call("Submit", collectForm());
}

document.getElementById("exit").onclick = () => runtime().Quit();
document.getElementById("home").onclick = () => call("Home");
document.getElementById("back").onclick = () => call("Back");
document.getElementById("modal-yes").onclick = () => call("Confirm", true);
document.getElementById("modal-no").onclick = () => call("Confirm", false);

document.getElementById("main").addEventListener("click", (e) => {
  const tab = e.target.closest("[data-tab]");
  if (tab) {
    call("SetTab", tab.getAttribute("data-tab"));
    return;
  }
  if (e.target.id === "ok") {
    submit();
    return;
  }
  const tile = e.target.closest("[data-id]");
  if (tile) {
    const id = tile.getAttribute("data-id");
    if (id === "reportShow") {
      call("Submit", collectForm());
      return;
    }
    call("Tap", id);
  }
  const kb = e.target.closest(".kb-open");
  if (kb) openKb(kb.getAttribute("data-for"));
});

document.addEventListener("keydown", (e) => {
  if (e.target && (e.target.tagName === "INPUT" || e.target.tagName === "SELECT")) {
    if (e.key === "Enter") {
      e.preventDefault();
      submit();
    }
    return;
  }
  if (e.key === "Enter") {
    e.preventDefault();
    call("KeyEnter");
    return;
  }
  if (e.key === "Escape") {
    call("Back");
    return;
  }
  if (e.key.length === 1) call("Key", e.key);
});

function openKb(field) {
  const el = document.querySelector(`[data-field="${field}"]`);
  if (!el) return;
  kbTarget = el;
  document.getElementById("kb-display").value = el.value || "";
  setKbLayout("123");
  document.getElementById("kb").classList.remove("hidden");
}

function setKbLayout(which) {
  let rows;
  if (which === "ABC") rows = [["Q","W","E","R","T","Y","U","I","O","P"],["A","S","D","F","G","H","J","K","L"],["Z","X","C","V","B","N","M",".","123","abc"]];
  else if (which === "abc") rows = [["q","w","e","r","t","y","u","i","o","p"],["a","s","d","f","g","h","j","k","l"],["z","x","c","v","b","n","m",".","123","ABC"]];
  else rows = [["7","8","9"],["4","5","6"],["1","2","3"],["0",".","ABC","abc"]];
  document.getElementById("kb-keys").innerHTML = rows.map((r) =>
    `<div class="kb-row">${r.map((ch) => `<button type="button" data-k="${ch}">${ch}</button>`).join("")}</div>`
  ).join("");
}

document.getElementById("kb-keys").addEventListener("click", (e) => {
  const b = e.target.closest("[data-k]");
  if (!b) return;
  const ch = b.getAttribute("data-k");
  if (ch === "ABC" || ch === "abc" || ch === "123") {
    setKbLayout(ch);
    return;
  }
  const d = document.getElementById("kb-display");
  if (d.value.length < 19) d.value += ch;
});
document.getElementById("kb-ok").onclick = () => {
  if (kbTarget) kbTarget.value = document.getElementById("kb-display").value;
  document.getElementById("kb").classList.add("hidden");
};
document.getElementById("kb-back").onclick = () => document.getElementById("kb").classList.add("hidden");
document.getElementById("kb-one").onclick = () => {
  const d = document.getElementById("kb-display");
  d.value = d.value.slice(0, -1);
};
document.getElementById("kb-all").onclick = () => { document.getElementById("kb-display").value = ""; };

function boot() {
  if (!window.runtime || !window.go) {
    setTimeout(boot, 50);
    return;
  }
  runtime().EventsOn("ui", render);
  call("State").then(render);
}

boot();
