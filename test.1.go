package main

import (
    "github.com/kataras/iris"
)
func mainw() {
    app := iris.New()
    // Load all templates from the "./views" folder
    // where extension is ".html" and parse them
    // using the standard `html/template` package.
    app.RegisterView(iris.HTML("./views", ".html"))

    // Method:    GET
    // Resource:  http://localhost:8080
    app.Get("/", func(ctx iris.Context) {

        ctx.WriteString("OK")
    })

    // Start the server using a network address.
    app.Run(iris.Addr(":8080"))
}