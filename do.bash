#
# This is AWESOME.
#
# Original code by: Chris F.A. Johnson
#
ESC="" ##  A literal escape character
but_row=1

clear
mv=1000  ## mv=1000 for press and release reporting; mv=9 for press only

_STTY=$(stty -g)      ## Save current terminal setup
stty -echo -icanon    ## Turn off line buffering
printf "${ESC}[?${mv}h        "   ## Turn on mouse reporting
printf "${ESC}[?25l"  ## Turn off cursor

while :
do
  x=$(dd bs=1 count=6 2>/dev/null) ## Read six characters

  m1=${x#???}    ## Remove the first 3 characters
  m2=${x#????}   ## Remove the first 4 characters
  m3=${x#?????}  ## Remove the first 5 characters

  ## Convert to characters to decimal values
  eval "$(printf "mb=%d mx=%d my=%d" "'$m1" "'$m2" "'$m3")"

  ## Values > 127 are signed
  [ $mx -lt 0 ] && MOUSEX=$(( 223 + $mx )) || MOUSEX=$(( $mx - 32 ))
  [ $my -lt 0 ] && MOUSEY=$(( 223 + $my )) || MOUSEY=$(( $my - 32 ))

  ## Button pressed is in first 2 bits; use bitwise AND
  BUTTON=$(( ($mb & 3) + 1 ))
  echo $MOUSEX $MOUSEY $BUTTON > mouse.out

done

printf "${ESC}[?${mv}l"  ## Turn off mouse reporting
stty "$_STTY"            ## Restore terminal settings
printf "${ESC}[?12l${ESC}[?25h" ## Turn cursor back on
printf "\n${ESC}[0J\n"   ## Clear from cursor to bottom of screen
