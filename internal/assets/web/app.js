const $ = (id) => document.getElementById(id);
let config, runId = null, running = false, previewTimer, generated;
let rawOutput = '';
const controls = [$('scala'), $('example')];
const syntax = window.Prism?.languages.scala;
if (syntax) {
  // Supplement the Scala grammar's soft keywords and capitalized type names.
  syntax.keyword = new RegExp(`${syntax.keyword.source}|\\b(?:end|export|then)\\b`);
  Prism.languages.insertBefore('scala', 'keyword', {
    'type-name': { pattern: /\b[A-Z]\w*\b/, alias: 'class-name' }
  });
}
function highlightCode(text, target) {
  if (syntax) target.innerHTML = Prism.highlight(text, Prism.languages.scala, 'scala');
  else target.textContent = text;
}
function syncScroll() {
  $('lines').scrollTop = $('code').scrollTop;
  $('highlight').scrollTop = $('code').scrollTop;
  $('highlight').scrollLeft = $('code').scrollLeft;
}
function renderCode() {
  const text = $('code').value;
  highlightCode(text.endsWith('\n') ? `${text} ` : text, $('highlighted'));
  $('code-area').classList.toggle('syntax-ready', Boolean(syntax));
  syncScroll();
}
function requestBody() {
  return { code: $('code').value, scalaVersion: $('scala').value };
}
async function post(path, body) {
  const response = await fetch(path, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: `HTTP ${response.status}` }));
    throw new Error(error.error);
  }
  return response;
}
function setStatus(text, state = '') {
  $('status').textContent = text;
  $('status-dot').className = `dot ${state}`;
}
function syncEditor() {
  renderCode();
  $('scala-version').textContent = `Scala ${$('scala').value}`;
  const count = $('code').value.split('\n').length;
  $('lines').textContent = Array.from({ length: count }, (_, i) => i + 1).join('\n');
  $('line-count').textContent = `${count} lines`;
  try { localStorage.setItem('zio-learn', JSON.stringify(requestBody())); } catch { /* Storage can be unavailable. */ }
  clearTimeout(previewTimer);
  if ($('generated').open) previewTimer = setTimeout(updatePreview, 200);
}
async function updatePreview() {
  try {
    generated = await (await post('/api/script', requestBody())).json();
    highlightCode(generated.source, $('source'));
  } catch (error) { generated = null; $('source').textContent = error.message; }
}
function appendOutput(text) {
  rawOutput += text;
  $('output').textContent = rawOutput.replace(/\x1b\[[0-?]*[ -/]*[@-~]/g, '');
  $('output').scrollTop = $('output').scrollHeight;
}
function onEvent(event) {
  if (event.type === 'started') { runId = event.id; $('stop').disabled = false; }
  if (event.type === 'output') appendOutput(event.text);
  if (event.type === 'finished') {
    const labels = { success: '実行完了', error: `実行失敗 (exit ${event.exitCode})`, stopped: '停止しました', timeout: '実行時間の上限（120 秒）に達しました', output_limit: '出力の上限（1 MiB）に達しました' };
    setStatus(labels[event.status] || '実行が中断されました', event.status === 'success' ? '' : 'error');
    $('duration').textContent = `${(event.durationMs / 1000).toFixed(1)}s`;
    if (!$('output').textContent) appendOutput('出力はありません。');
    return true;
  }
  return false;
}
async function run() {
  if (running || !config) return;
  running = true;
  runId = null;
  $('run').disabled = true;
  controls.forEach(el => el.disabled = true);
  $('output').textContent = '';
  rawOutput = '';
  $('duration').textContent = '';
  setStatus('コンパイル・実行中…', 'running');
  let finished = false;
  try {
    const response = await post('/api/run', requestBody());
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    for (;;) {
      const { value, done } = await reader.read();
      buffer += decoder.decode(value, { stream: !done });
      let newline;
      while ((newline = buffer.indexOf('\n')) >= 0) {
        const line = buffer.slice(0, newline);
        buffer = buffer.slice(newline + 1);
        if (line) finished = onEvent(JSON.parse(line)) || finished;
      }
      if (done) break;
    }
    if (!finished) throw new Error('サーバーとの接続が終了しました。');
  } catch (error) { setStatus('実行できませんでした', 'error'); appendOutput(`\n${error.message}\n`); }
  finally {
    running = false; runId = null;
    $('run').disabled = false; $('stop').disabled = true;
    controls.forEach(el => el.disabled = false);
  }
}
$('run').addEventListener('click', run);
$('stop').addEventListener('click', async () => {
  if (!runId) return;
  $('stop').disabled = true;
  try { await post('/api/stop', { id: runId }); }
  catch (error) { setStatus(error.message, 'error'); $('stop').disabled = !running; }
});
$('clear').addEventListener('click', () => { rawOutput = ''; $('output').textContent = ''; });
$('code').addEventListener('input', syncEditor);
$('code').addEventListener('scroll', syncScroll);
$('scala').addEventListener('change', syncEditor);
$('example').addEventListener('change', () => {
  const name = $('example').value;
  if (!name) return;
  $('code').value = config.examples[name];
  $('example').value = '';
  syncEditor();
});
$('generated').addEventListener('toggle', () => { if ($('generated').open) updatePreview(); });
$('download').addEventListener('click', async () => {
  await updatePreview();
  if (!generated) return;
  const url = URL.createObjectURL(new Blob([generated.source], { type: 'text/plain;charset=utf-8' }));
  const link = document.createElement('a');
  link.href = url; link.download = generated.filename; link.click();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
});
document.addEventListener('keydown', (event) => {
  if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') { event.preventDefault(); run(); }
});
$('code').addEventListener('keydown', (event) => {
  if (event.key !== 'Tab') return;
  event.preventDefault();
  $('code').setRangeText('  ', $('code').selectionStart, $('code').selectionEnd, 'end');
  syncEditor();
});
async function init() {
  try {
    const response = await fetch('/api/config');
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    config = await response.json();
    let saved;
    try { saved = JSON.parse(localStorage.getItem('zio-learn')); } catch { /* Use defaults. */ }
    $('code').value = typeof saved?.code === 'string' ? saved.code : config.examples['hello.scala'];
    for (const version of config.scalaVersions) {
      const option = document.createElement('option');
      option.value = version; option.textContent = version;
      $('scala').append(option);
    }
    $('scala').value = config.scalaVersions.includes(saved?.scalaVersion) ? saved.scalaVersion : config.scalaVersion;
    $('scala').disabled = false;
    syncEditor();
    setStatus('実行できます');
    $('run').disabled = false;
  } catch (error) { setStatus(`読み込み失敗: ${error.message}`, 'error'); }
}
init();
