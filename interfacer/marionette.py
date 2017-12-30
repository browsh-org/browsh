# The protocal is actually quite simple:
#   message_size:[command_type, message_id, command, params]
# Eg; the following starts a new session:
#   `o=JSON.dump([0,2,"newSession",{}]); puts "#{o.length}:#{o}"`
# So write this up in rust :D

from marionette_driver.marionette import Marionette
from marionette_driver.addons import Addons

client = Marionette('localhost', port=2828)
client.start_session()
print("session started")

# addons = Addons(client)
# addons.install("/home/tombh/Workspace/texttop/webext/web-ext-artifacts/browsh-0.1-an+fx.xpi")

client.navigate('http://media.giphy.com/media/3o6Zt4FQZaiDpAVu1O/giphy.gif')
print(client.get_url())
