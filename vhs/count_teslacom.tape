Output ../frontend/static/count_teslacom.gif

Require curl

Set Shell "bash"
Set FontSize 26
Set Width 2000
Set Height 400

#Set Margin 20
#Set MarginFill "#dadadb"
#Set BorderRadius 20

Set Theme { "name": "Elmasy", "black": "#3d3d3d", "red": "#1DDDDD", "green": "#1DDDDD", "yellow": "#1DDDDD", "blue": "#1DDDDD", "magenta": "#1DDDDD", "cyan": "#1DDDDD", "white": "#DADADB", "brightBlack": "#3d3d3d", "brightRed": "#1DDDDD", "brightGreen": "#1DDDDD", "brightYellow": "#1DDDDD", "brightBlue": "#1DDDDD", "brightMagenta": "#1DDDDD", "brightCyan": "#1DDDDD", "brightWhite": "#DADADB", "background": "#3d3d3d", "foreground": "#DADADB", "selection": "#DADADB", "cursor": "#1DDDDD" }

Hide
# Set PS1
Type "PS1='\[\e[38;2;29;221;221m\]> \[\e[0m\]'"
Enter

# Remove outputs
Type "clear"
Enter
Show

Type 'time curl -s -H "Accept: text/plain" "https://columbus.elmasy.com/api/lookup/tesla.com" | wc -l'
Sleep 500ms
Enter 1

Sleep 5s

