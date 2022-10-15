package browsh

var configSample = `
# See; https://www.brow.sh/donate/
# By showing your support you can disable the app's branding and nags to donate.
browsh_supporter = "♥"

# The page to show at startup. Browsh will fail to boot if this URL is not accessible
startup-url = "http://www.brow.sh"

# The base query when a non-URL is entered into the URL bar
default_search_engine_base = "https://www.google.com/search?q="

# The mobile user agent for forcing web pages to use their mobile layout
mobile_user_agent = "Mozilla/5.0 (Android 7.0; Mobile; rv:54.0) Gecko/58.0 Firefox/58.0"

[browsh] # Browsh internals
websocket-port = 3334

# Possibly better handling of overlapping text in web pages. If a page seems to have
# text that shouldn't be visible, if it should be behind another element for example,
# then this experimental feature should help. It can also be toggled in-browser with F6.
use_experimental_text_visibility = false

# Custom CSS to apply to all loaded tabs, eg;
#   custom_css = """
#   body {
#     background-colour: black;
#   }
#   """
custom_css = ""

[firefox]
# The path to your Firefox binary
path = "firefox"
# Browsh has its own profile, separate from the normal user's. But you can change that.
profile = "browsh-default"
# Don't let Browsh launch Firefox, but make it try to connect to an existing one. Note
# it will need to have been launched with the '--marionette' flag.
use-existing = false
# Launch Firefox in with its visible GUI window. Useful for setting up the Browsh profile.
with-gui = false

# Config that you might usually set through Firefox's 'about:config' page
# Note that string must be wrapped in quotes
# preferences = [
#   "privacy.resistFingerprinting=true",
#   "network.proxy.http='localhost'",
#   "network.proxy.ssl='localhost'",
#   "network.proxy.http_port=8118",
#   "network.proxy.ssl_port=8118",
#   "network.proxy.type=1"
# ]

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
# parsed, too long and you wait unnecessarily.
render_delay = 100

# The length of time in seconds to wait before aborting the page load
timeout = 30

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
rate-limit = "100000000-M"

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
