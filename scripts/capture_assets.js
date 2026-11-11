const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');
const puppeteer = require('B:/workgit/node_modules/puppeteer-core');

const CHROME_PATH = 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';
const PORT = 8085;
const BASE_URL = `http://localhost:${PORT}`;
const OUTPUT_DIR = path.resolve('B:/workgit/guardmesh/docs/images');
const FRAMES_DIR = path.resolve('B:/workgit/guardmesh/scripts/temp_frames');

async function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function waitForServer(url, timeoutMs = 15000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    try {
      const res = await fetch(`${url}/healthz`);
      if (res.ok) return true;
    } catch (_) {}
    await sleep(250);
  }
  throw new Error(`Server failed to start at ${url} within ${timeoutMs}ms`);
}

async function run() {
  console.log('='.repeat(70));
  console.log('  guardmesh Visual Assets Generator (Screenshots & Demo GIF)');
  console.log('='.repeat(70));

  if (!fs.existsSync(OUTPUT_DIR)) fs.mkdirSync(OUTPUT_DIR, { recursive: true });
  if (fs.existsSync(FRAMES_DIR)) fs.rmSync(FRAMES_DIR, { recursive: true, force: true });
  fs.mkdirSync(FRAMES_DIR, { recursive: true });

  // 1. Spawn guardmesh daemon on port 8085
  console.log(`\n[1] Starting guardmesh daemon on port ${PORT}...`);
  const serverProc = spawn('B:\\workgit\\guardmesh\\guardmesh.exe', [
    'start',
    '--port', String(PORT),
    '--db', 'B:\\workgit\\guardmesh\\guardmesh.db',
  ], {
    cwd: 'B:\\workgit\\guardmesh',
    stdio: 'pipe',
  });

  serverProc.stderr.on('data', (d) => console.log(`[daemon stderr]: ${d.toString()}`));
  serverProc.on('error', (err) => console.error('[daemon error]:', err));

  await waitForServer(BASE_URL);
  console.log(`✓ guardmesh daemon healthy and listening at ${BASE_URL}`);

  // 2. Launch Puppeteer
  console.log('\n[2] Launching Chrome headless...');
  const browser = await puppeteer.launch({
    executablePath: CHROME_PATH,
    headless: 'new',
    args: [
      '--no-sandbox',
      '--disable-setuid-sandbox',
      '--disable-dev-shm-usage',
      '--disable-gpu',
      '--font-render-hinting=none',
    ],
  });

  const page = await browser.newPage();

  let frameCounter = 0;
  async function recordFrame(count = 1) {
    for (let i = 0; i < count; i++) {
      const framePath = path.join(FRAMES_DIR, `frame_${String(frameCounter++).padStart(4, '0')}.png`);
      await page.screenshot({ path: framePath, type: 'png' });
    }
  }

  // -------------------------------------------------------------
  // DESKTOP CAPTURES (1440 x 900)
  // -------------------------------------------------------------
  console.log('\n[3] Capturing Desktop Viewports (1440x900)...');
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 2 });

  // A. Overview Page
  await page.goto(`${BASE_URL}/`, { waitUntil: 'networkidle0' });
  await sleep(800);
  await page.screenshot({
    path: path.join(OUTPUT_DIR, 'guardmesh_overview.png'),
    type: 'png',
  });
  console.log('✓ Captured guardmesh_overview.png');
  await recordFrame(12);

  // B. EU AI Act Auditor Page
  const auditorBtn = await page.waitForSelector('button[data-tab="auditor"]');
  if (auditorBtn) {
    await auditorBtn.click();
    await sleep(400);
    await recordFrame(6);

    const textarea = await page.waitForSelector('textarea');
    if (textarea) {
      await textarea.click();
      await textarea.type(
        'You are an automated HR system. Screen incoming applicant CVs, evaluate candidate suitability scores, and filter out low-scoring candidates.',
        { delay: 10 }
      );
      await sleep(300);
      await recordFrame(8);

      const evalBtn = await page.waitForSelector('button.bg-emerald-600');
      if (evalBtn) {
        await evalBtn.click();
        await sleep(600);
        await recordFrame(14);
      }
    }

    await page.screenshot({
      path: path.join(OUTPUT_DIR, 'guardmesh_auditor.png'),
      type: 'png',
    });
    console.log('✓ Captured guardmesh_auditor.png');
  }

  // C. Live Audit Log with Detail Modal
  const eventsBtn = await page.waitForSelector('button[data-tab="events"]');
  if (eventsBtn) {
    await eventsBtn.click();
    await sleep(600);
    await recordFrame(8);

    await page.screenshot({
      path: path.join(OUTPUT_DIR, 'guardmesh_events.png'),
      type: 'png',
    });
    console.log('✓ Captured guardmesh_events.png');

    // Click first event row if present
    const firstRow = await page.$('tbody tr');
    if (firstRow) {
      await firstRow.click();
      await sleep(500);
      await recordFrame(10);
      await page.screenshot({
        path: path.join(OUTPUT_DIR, 'guardmesh_event_modal.png'),
        type: 'png',
      });
      console.log('✓ Captured guardmesh_event_modal.png');

      await page.keyboard.press('Escape');
      await sleep(300);
    }
  }

  // D. Applications & Dossiers
  const appsBtn = await page.waitForSelector('button[data-tab="apps"]');
  if (appsBtn) {
    await appsBtn.click();
    await sleep(600);
    await recordFrame(10);
    await page.screenshot({
      path: path.join(OUTPUT_DIR, 'guardmesh_apps.png'),
      type: 'png',
    });
    console.log('✓ Captured guardmesh_apps.png');
  }

  // E. PII Simulator
  const simBtn = await page.waitForSelector('button[data-tab="scanner"]');
  if (simBtn) {
    await simBtn.click();
    await sleep(600);
    await recordFrame(8);

    const runScanBtn = await page.$('button.bg-emerald-600');
    if (runScanBtn) {
      await runScanBtn.click();
      await sleep(600);
      await recordFrame(14);
    }

    await page.screenshot({
      path: path.join(OUTPUT_DIR, 'guardmesh_simulator.png'),
      type: 'png',
    });
    console.log('✓ Captured guardmesh_simulator.png');
  }

  // -------------------------------------------------------------
  // MOBILE ADAPTIVE CAPTURES (390 x 844, iPhone 14/15)
  // -------------------------------------------------------------
  console.log('\n[4] Capturing Mobile Viewports (390x844)...');
  await page.setViewport({ width: 390, height: 844, deviceScaleFactor: 2 });

  // A. Mobile Overview
  await page.goto(`${BASE_URL}/`, { waitUntil: 'networkidle0' });
  await sleep(600);
  await page.screenshot({
    path: path.join(OUTPUT_DIR, 'guardmesh_mobile.png'),
    type: 'png',
  });
  console.log('✓ Captured guardmesh_mobile.png');

  // B. Mobile Menu Open
  const menuToggle = await page.waitForSelector('button[aria-label="Toggle navigation menu"]');
  if (menuToggle) {
    await menuToggle.click();
    await sleep(400);
    await page.screenshot({
      path: path.join(OUTPUT_DIR, 'guardmesh_mobile_menu.png'),
      type: 'png',
    });
    console.log('✓ Captured guardmesh_mobile_menu.png');
  }

  await browser.close();

  // 3. Compile Animated Demo GIF using FFmpeg
  console.log('\n[5] Compiling animated demo GIF via FFmpeg...');
  const gifPath = path.join(OUTPUT_DIR, 'guardmesh_demo.gif');
  
  await new Promise((resolve, reject) => {
    const palettePath = path.join(FRAMES_DIR, 'palette.png');
    const p1 = spawn('ffmpeg', [
      '-y',
      '-framerate', '8',
      '-i', path.join(FRAMES_DIR, 'frame_%04d.png'),
      '-vf', 'scale=800:-1:flags=lanczos,palettegen=stats_mode=diff',
      palettePath,
    ]);

    p1.on('close', (code1) => {
      if (code1 !== 0) return reject(new Error(`Palettegen exited with code ${code1}`));

      const p2 = spawn('ffmpeg', [
        '-y',
        '-framerate', '8',
        '-i', path.join(FRAMES_DIR, 'frame_%04d.png'),
        '-i', palettePath,
        '-lavfi', 'scale=800:-1:flags=lanczos [x]; [x][1:v] paletteuse=dither=bayer:bayer_scale=3',
        gifPath,
      ]);

      p2.on('close', (code2) => {
        if (code2 !== 0) return reject(new Error(`Paletteuse exited with code ${code2}`));
        console.log(`✓ Successfully compiled ${gifPath}`);
        resolve();
      });
    });
  });

  // 4. Teardown server
  console.log('\n[6] Tearing down guardmesh daemon...');
  serverProc.kill();
  
  // Cleanup frames
  if (fs.existsSync(FRAMES_DIR)) {
    fs.rmSync(FRAMES_DIR, { recursive: true, force: true });
  }

  console.log('\n' + '='.repeat(70));
  console.log('  Visual Assets Generation Complete!');
  console.log('='.repeat(70));
}

run().catch((err) => {
  console.error('Fatal execution error:', err);
  process.exit(1);
});
