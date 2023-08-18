Output ../frontend/static/starts.gif

Require curl

Set Shell "bash"
Set FontSize 32
Set Width 1950
Set Height 240

Set Padding 0
#Set Margin 20
#Set MarginFill "#dadadb"
#Set BorderRadius 20

Set Theme { "name": "Elmasy", "black": "#1e1f21", "red": "#1DDDDD", "green": "#1DDDDD", "yellow": "#1DDDDD", "blue": "#1DDDDD", "magenta": "#1DDDDD", "cyan": "#1DDDDD", "white": "#DADADB", "brightBlack": "#1e1f21", "brightRed": "#1DDDDD", "brightGreen": "#1DDDDD", "brightYellow": "#1DDDDD", "brightBlue": "#1DDDDD", "brightMagenta": "#1DDDDD", "brightCyan": "#1DDDDD", "brightWhite": "#DADADB", "background": "#1e1f21", "foreground": "#DADADB", "selection": "#DADADB", "cursor": "#1DDDDD" }

Hide
# Set PS1
Type "PS1='\[\e[38;2;29;221;221m\]> \[\e[0m\]'"
Enter

# Add newline to curl
Type 'echo -w "\n" > ~/.curlrc'
Enter

# Remove outputs
Type "clear"
Enter

Show

Type 'curl -s "https://columbus.elmasy.com/api/starts/reddit" | head -c 76'
Sleep 1s
Enter 1
Sleep 5s

Type 'clear'
Sleep 500ms
Enter 1

Type 'curl -s -H "Accept: text/plain" "https://columbus.elmasy.com/api/starts/reddit" | head -n 4'
Sleep 1s
Enter 1
Sleep 5s

Hide
Type 'rm ~/.curlrc'
Enter

