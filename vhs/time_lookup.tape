Output ../frontend/static/time_lookup.gif

Require curl

Set Shell "bash"
Set FontSize 32
Set Width 1800
Set Height 400

#Set Margin 20
#Set MarginFill "#dadadb"
#Set BorderRadius 20

Set Theme { "name": "Elmasy", "black": "#3d3d3d", "red": "#1DDDDD", "green": "#1DDDDD", "yellow": "#1DDDDD", "blue": "#1DDDDD", "magenta": "#1DDDDD", "cyan": "#1DDDDD", "white": "#DADADB", "brightBlack": "#3d3d3d", "brightRed": "#1DDDDD", "brightGreen": "#1DDDDD", "brightYellow": "#1DDDDD", "brightBlue": "#1DDDDD", "brightMagenta": "#1DDDDD", "brightCyan": "#1DDDDD", "brightWhite": "#DADADB", "background": "#3d3d3d", "foreground": "#DADADB", "selection": "#DADADB", "cursor": "#1DDDDD" }

Hide
# Set PS1
Type "PS1='\[\e[38;2;29;221;221m\]> \[\e[0m\]'"
Enter

## Add newline to curl
#Type 'echo -w "\n" > ~/.curlrc'
#Enter

# Remove outputs
Type "clear"
Enter
Show

Type 'time curl -s -o /dev/null "https://columbus.elmasy.com/api/lookup/tesla.com"'
Sleep 500ms
Enter 1

Sleep 5s

Hide
Type 'rm ~/.curlrc'
Enter

