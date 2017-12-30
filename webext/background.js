let tty_width, tty_height;

// Create WebSocket connection.
const socket = new WebSocket('ws://localhost:2794');

// Connection opened
socket.addEventListener('open', function (_event) {
  console.log('Connected to Browsh client');
});

// Listen for messages
socket.addEventListener('message', (event) => {
  console.log('Message from server ', event.data);
  const parts = event.data.split(',');
  const command = parts[0];
  if (command === '/tty_size') {
    tty_width = parts[1];
    tty_height = parts[2];
  }
});

let active_tab_id;
function handleActivated(active_info) {
  console.log(active_info);
  active_tab_id = active_info.id
}
browser.tabs.onActivated.addListener(handleActivated);

function handleRegistration(_request, sender, sendResponse) {
  console.log('tab registered');
  sendResponse(sender.tab);
  if (sender.tab.active) active_tab_id = sender.tab.id;
}
browser.runtime.onMessage.addListener(handleRegistration);

let tabs = {};
function connected(channel) {
  console.log('tab connected');
  tabs[channel.name] = channel;
  resizeBrowserWindow();
  // Currently tabs will only ever send screen output over their channel
  channel.onMessage.addListener(function(screen) {
    socket.send(screen);
  });
}
browser.runtime.onConnect.addListener(connected);

setInterval(() => {
  if (tabs[active_tab_id] === undefined) return;
  tabs[active_tab_id].postMessage(`/send_frame,${tty_width},${tty_height}`);
}, 1000);

function resizeBrowserWindow() {
  const width = tty_width * 9;
  const height = tty_height * 19.5; // this is actually line-height
  var getting = browser.windows.getCurrent();
  getting.then(
    (windowInfo) => {
      console.log('resizing window', windowInfo, width, height);
      const updating = browser.windows.update(
        windowInfo.id,
        {
          width: width,
          height: height
        }
      );

      updating(
        (info) => {
          console.log(info);
        },
        (error) => {
          console.error(error);
        }
      );
    },
    (error) => {
      console.error(error);
    }
  );
}
