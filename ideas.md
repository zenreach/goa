* revise how validations are done, follow http://json-schema.org/latest/json-schema-validation.html
* goagen builds in temp folder then runs app with --bootstrap then builds results
* Reports build errors as is
* //go generate
* docker
* make resource stuff independent of handlers, create middleware for all go web frameworks.
* generate docs, client, middleware code, bootstrap code (framework specific "controllers") from resource.
* battery included - provide default web app framework.
* just goagen does everything by default (update semantic tbd)

```
goagen [--target=gin|negroni|martini|goji] [--bootstrap] [--middleware] [--docs] [--cli=NAME] [--gui]
```
