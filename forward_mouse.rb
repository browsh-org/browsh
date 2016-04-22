require 'curses'
include Curses

# These 2 allow mouse input
noecho # Prevent input being echoed to the screen
stdscr.keypad(true) # Not sure what this does exactly, but mouse input doesn't work without it

# Block whilst waiting for input
stdscr.timeout = -1

mousemask(ALL_MOUSE_EVENTS)

loop do
  case getch
  when KEY_MOUSE
    m = getmouse
    File.open('/home/tombh/Workspace/lowbandwidth/mouse.out', 'w') do |file|
      file.write("#{m.x}, #{m.y}, #{m.z}, #{m.bstate}")
    end
  end
end
