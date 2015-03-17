* goagen builds in temp folder then runs app with --bootstrap then builds results
* Reports build errors as is
* //go generate
* docker
* make resource stuff independent of handlers, create middleware for all go web frameworks.
* generate docs, client, middleware code, bootstrap code (framework specific "controllers") from resource.
* battery included - provide default web app framework.

```
goagen [--target=gin|negroni|martini|goji] [--bootstrap] [--middleware] [--docs] [--cli=NAME] [--gui]
```
