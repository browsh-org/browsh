# Intercept all STDIN from the user's local machine and forward to the
# remote X framebuffer via xdotool. This is exposed via an SSH/MoSH session.
#
# Hat tip to Chris F.A. Johnson whereever you are.

ESC="" # A literal escape character (it's not blank!)
clear
mv=1000 # mv=1000 for press and release reporting; mv=9 for press only
_STTY=$(stty -g) # Save current terminal setup
stty -echo -icanon # Turn off line buffering
printf "${ESC}[?${mv}h        " # Turn on mouse reporting
printf "${ESC}[?25l" # Turn off cursor

# Ratios to convert mouse clicks in the terminal to mouse clicks on the virtual desktop
XFACTOR=$(($DESKTOP_WIDTH / $TERMINAL_WIDTH))
YFACTOR=$(($DESKTOP_HEIGHT / $TERMINAL_HEIGHT))

while :
do
  # TODO: parse actual keyboard presses
  # TODO: use the more modern mouse modes, eg; 1015

  x=$(dd bs=1 count=6 2>/dev/null) # Read six characters

  m1=${x#???}    # Remove the first 3 characters
  m2=${x#????}   # Remove the first 4 characters
  m3=${x#?????}  # Remove the first 5 characters

  # Convert from characters to decimal values
  eval "$(printf "mb=%d mx=%d my=%d" "'$m1" "'$m2" "'$m3")"

  # Values > 127 are signed
  [ $mx -lt 0 ] && MOUSEX=$(( 223 + $mx )) || MOUSEX=$(( $mx - 32 ))
  [ $my -lt 0 ] && MOUSEY=$(( 223 + $my )) || MOUSEY=$(( $my - 32 ))

  case $mb in
  32) # Left down
    BUTTON=1
    BPRESS='mousedown'
    ;;
  33) # Middle down
    BUTTON=2
    BPRESS='mousedown'
    ;;
  34) # Right down
    BUTTON=3
    BPRESS='mousedown'
    ;;
  35) # Last button up
    BUTTON=$BUTTON
    BPRESS='mouseup'
    ;;
  97)
    BUTTON=4
    BPRESS='click'
    ;;
  96)
    BUTTON=5
    BPRESS='click'
    ;;
  esac

  WINX=$(($MOUSEX * $XFACTOR))
  WINY=$(($MOUSEY * $YFACTOR))

  if [ "$DEBUG" = "true" ]; then
    echo "$MOUSEX"x"$MOUSEY" "$WINX"x"$WINY" "$BUTTON":"$BPRESS"
  fi

  xdotool mousemove $WINX $WINY
  xdotool keydown alt $BPRESS $BUTTON keyup alt

  echo "xdotool mousemove $WINX $WINY" >> mouse.out
  echo "xdotool keydown alt $BPRESS $BUTTON keyup alt" >> mouse.out
done

printf "${ESC}[?${mv}l" # Turn off mouse reporting
stty "$_STTY" # Restore terminal settings
printf "${ESC}[?12l${ESC}[?25h" # Turn cursor back on
printf "\n${ESC}[0J\n" # Clear from cursor to bottom of screen
