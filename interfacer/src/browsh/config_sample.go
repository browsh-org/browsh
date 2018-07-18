package browsh

var configSample = `
# See; https://www.brow.sh/donate/
# By showing your support you can disable the app's branding and nags to donate.
browsh_supporter = "♥"

# The base query when a non-URL is entered into the URL bar
default_search_engine_base = "https://www.google.com/search?q="

# The mobile user agent for forcing web pages to use their mobile layout
mobile_user_agent = "Mozilla/5.0 (Android 7.0; Mobile; rv:54.0) Gecko/58.0 Firefox/58.0"

[browsh] # Browsh internals
websocket-port = 3334

[firefox]
path = "firefox"
profile = "default"
use-existing = false
with-gui = false

[tty]
# The time in milliseconds between requesting a new TTY-sized pixel frame.
# This is essentially the frame rate for graphics. Lower values make for smoother
# animations and feedback, but also increases the CPU load.
small_pixel_frame_rate = 250

[http-server]
port = 4333
bind = "0.0.0.0"

# The time to wait in milliseconds after the DOM is ready before
# trying to parse and render the page's text. Too soon and text risks not being
# parsed, too long and you wait unecessarily.
render_delay = 400

# The dimensions of a char-based window onto a webpage.
# The columns are ultimately the width of the final text. Whereas the rows
# represent the height of the original web page made visible to the original
# browser window. So the number of rows can effect things like how far down a
# web page images are lazy-loaded.
columns = 100
rows = 30

# The amount of lossy JPG compression to apply to the background image of HTML
# pages.
jpeg_compression = 0.9

# Rate limit. For syntax, see: https://github.com/ulule/limiter
rate-limit = "10-M"

# Blocking is useful if the HTTP server is made public. All values are evaluated as
# regular expressions.
blocked-domains = [
]

blocked-user-agents = [
]

# HTML snippets to show at top and bottom of final page.
header = ""
footer = ""
`
