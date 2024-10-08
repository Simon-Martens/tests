To me the whole point of templating is that you have a textfile OUTSIDE of the program code that can determine the application look and feel. Like you can give a bunch of .tmpl files and data availablity specs to the frontend team. TEMPL templates get compiled into go code, then get compiled by the compiler. Whats the point of that? 
1) It adds not one but two compilation steps, since simple go tmpl files are just a bunch of tmpl files in a folder read in by go at runtime, not neccessary at compile time.
2) It intertwines go code, which is procedural, with template code, which is descriptive. But to be able to just differentiate between those two I always found to be the good kind of separation of concerns (reaching a state of kubernetes being the bad kind).

```go
    text := `
    <html><head>
    {{ block "head" . }}Default{{ end }}
        </head><body>
        {{ block "body" . }}Default Body{{ end }}
        </body></html>
    `
    layout, err := template.New(DEFAULT_LAYOUT_NAME).Parse(text)
```
This creates a template with a template.namespace.set length of 3: 
- DEFAULT_VALUE_NAME
- "head"
- "body"

STATE BEFORE ADD PARSE TREE:

1. "/" is in the name space as an empty template. 
2. escapeError is wrapped in the Path "/"
3. tc has no tree but a set of two "/" & "body"
4. layout has the ROOT one, but the ns len is 1 (should there be 3, for body and head?)

AFTER ADD PARSE TREE:

ITS ALL WEIRD