# grep-go
## Description:
This project is a custom 'grep' implementation using regular expressions.

### What the program does:
This CLI program runs a script which compares user input with a regex pattern. Exits with 0 on successful match and 1 if unsuccessful.

### How to use:
```sh
echo "input text" | ./grep_go.sh -E "pattern to match"
```

## Patterns:
- Literal character:
"c" will match "c"

- Digits '\d':
"\d" will match any single digit

- Alphanumeric characters '\w'
"\w" will match and single alphanumeric character or '_'

- Positive character group: [...]
"[abc]" will match any character inside the brackets
ex: "advance" is fine, but "event" is not

- Negative character group [^...]:
"[^abc]" will match any character NOT inside the brackets
ex: "event" is fine, but "advance" is not

- Start of line '^'
"^" matches the start of a line
ex: "^night" will match "night", but not "knight"

- End of line '$'
"$" matches the end of a line
ex: "bug$" will match "bug", but not "bugs"

- One or more quantifier '+'
"o+" will match "owl" and "booster", but not "event"

- Zero or one quantifier '?'
"owls?" will match "owl" and "owls", but not "fish"

- Wildcard '.'
"c.t" will match "cat" and "cut", but not "owl"

- Alternation (either/or) '|'
"bug | bean" will match "bug" and "bean", but not "fish"

- Backreferences (...) then \1, \2, etc
- Each group (...) captured will be the \x group depending on the number
"(ducks) are \1" will match "ducks and ducks", but not "ducks and owls" 
"(\d) (\w+) are \1 \2" will match "3 bugs are 3 bugs", but not "3 bugs are 3 cats"
