import fs from 'fs';
import pty from 'pty.js';
import child from 'child_process';
import stripAnsi from 'strip-ansi';

before((done) => {
  helper.boot(done);
});

after(() => {
  helper.shutdown();
});

class Helper {
  constructor () {
    this.frame = '';
    this.tty_width = 70;
    this.tty_height = 30;
    this.is_last_startup_message_consumed = false;
    this.browserFingerprint = ' ← | x | ';
    this.project_root = child.execSync('git rev-parse --show-toplevel').toString().trim();
  }

  log(message) {
    const log_file = this.project_root + '/interfacer/spec.log';
    message = stripAnsi(message);
    fs.appendFileSync(log_file, message);
  }

  boot(callback) {
    // Race condition is avoided because Firefox "should" consistently take longer
    // to start than the CLI
    this.startBrowshPTY();
    this.startFirefox();
    this.consumeStartupOutput(callback);
  }

  shutdown() {
    this.stopWatching = true;
    this.browshPTY.destroy();
    this.firefoxPTY.destroy();
  }

  startBrowshPTY() {
    const dir = this.project_root + '/interfacer';
    this.browshPTY = pty.spawn('bash', [], {
      cols: this.tty_width,
      rows: this.tty_height,
      env: process.env
    });
    this.browshPTY.write(`cd ${dir} \r`);
    this.browshPTY.write(`go run *.go -use-existing-ff \r`);
    this.broadcastOutput();
  }

  broadcastOutput() {
    let buffer = '';
    this.browshPTY.on('data', (data) => {
      if (this.is_last_startup_message_consumed) {
        buffer += data;
        buffer = this.broadcastBrowserOutput(buffer);
      } else {
        this.log(data);
        this.frame = this.cleanFrame(data);
      }
    });
  }

  // pty.js sends chunks, so we need to wait for a particular signature before
  // we have a whole browser frame with which we can work with.
  broadcastBrowserOutput(buffer) {
    const cursor_reset_sig = '\u001b[1;1H';
    if (buffer.includes(cursor_reset_sig)) {
      buffer = this.cleanFrame(buffer);
      this.frame = this.insertTTYLines(buffer);
      buffer = '';
    }
    return buffer;
  }

  // TODO: Handle wide UTF8 chars in the same way the app does
  insertTTYLines(buffer) {
    let split = '';
    for (var i = 0; i < buffer.length; i++) {
      if (((i + 1) % this.tty_width) === 0) {
        split += buffer[i] + '\n';
      } else {
        split += buffer[i];
      }
    }
    return split;
  }

  // Currently we're just converting the browser output into pure alphanumerical
  // text, ie no colour information.
  cleanFrame(buffer) {
    buffer = stripAnsi(buffer);
    buffer = buffer.replace(/▄/g, ' ');
    buffer = buffer.trim();
    return buffer;
  }

  // Wait for the given string to appear anywhere in the entire frame, whether in
  // the UI or the webpage.
  watchOutputFor(match, callback) {
    const interval = setInterval(() => {
      const regex = new RegExp(match, 'g');
      if (this.stopWatching) clearInterval(interval);
      if (regex.test(this.frame)) {
        clearInterval(interval);
        callback(this.frame);
      }
    }, 5);
  }

  getPage(url, done) {
    const signature = this.browserFingerprint + url;
    this.watchOutputFor(signature, (frame) => {
      done(this.buildPageObject(frame));
    });
  }

  buildPageObject(frame) {
    let frame_lines = [];
    for(let line of frame.split(/\r?\n/)) {
      line = line.replace(/\s+$/, ''); // Right trim
      frame_lines.push(line);
    }
    return {
      title: frame_lines[0],
      url: frame_lines[1].replace(this.browserFingerprint, ''),
      body: frame_lines.slice(2)
    }
  }

  // Wait until the penultimate message before actual webpage content is shown
  consumeStartupOutput(done) {
    this.watchOutputFor('Waiting for a Firefox instance to connect', (_) => {
      this.is_last_startup_message_consumed = true;
      this.frame = '';
      done();
    });
  }

  // Firefox doesn't actually need a PTY, but seeing as we're using pty.js already
  // then may as well keep consistent.
  startFirefox() {
    const dir = this.project_root + '/webext/dist';
    this.firefoxPTY = pty.spawn('bash', [], {
      env: process.env
    });
    this.firefoxPTY.write(`cd ${dir} \r`);
    let command = `../node_modules/.bin/web-ext run ` +
      `--firefox="${this.project_root}/webext/contrib/firefoxheadless.sh" ` +
      `--url https://google.com` +
      `\r`;
    this.firefoxPTY.write(command);
    this.firefoxPTY.on('data', (data) => {
      this.log(data);
    });
  }
}

let helper = new Helper();
export default helper;
