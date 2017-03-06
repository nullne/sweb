# sweb
Swagger Editor Backend

sweb opens swagger editor and save the content every n seconds. Except this, @nullne add feature static file server so
that people can dive into their local changes immediately by exploring url: `127.0.0.1:8765/static/swagger.yaml` in the swagger UI.
Don't forget to remove your leading test url from your object definition because there is no more need once they are deployed
online.

## Installation

```
go get -u github.com/nullne/sweb
```

## usage

```
Usage of sweb:
  -f string
    	the full path to the document being edited (default "api-spec.yaml")
  -p string
    	port for editor's http backend (default "8765")
  -se string
    	the full path to swagger-editor installation (default "builtin")
```

### Important stuff

The backend can only run in one instance, and you'll be able to work on one file at a time. If there will be demand, it is possible that it will support multiple files.
There is no save button. Whatever you do is saved. There is undo in the editor, but it's still easy to mess up. The idea is that you use this tool inside a git repository
so you can revert changes etc. 

### Appreciation
Thank you @zgiber for you effort
