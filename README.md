# binfiles
Provides compatibility with embed.FS for Macaron-Go

____

Binfile allows you to use embed.fs to store the template files and public folder files in macaron-go and thereby generate a single executable file with this information embedded.

Project under development, but functional at the moment.

Usage: 

```GO
//go:embed public
var staticFS embed.FS

//go:embed templates
var teplatesFS embed.FS

func main() {
    // Create macaron instance
    m := macaron.New()
	
	// Public files
	m.Use(macaron.Static("public",
		macaron.StaticOptions{
			FileSystem:  binfiles.New(&staticFS, "public"),
			SkipLogging: true,
			ETag:        true,
			Expires:     func() string { return time.Now().Add(time.Minute * time.Duration(60)).Format(http.TimeFormat) },
		},
	))
	m.Use(macaron.Recovery())

	// Template files (Pongo2)
	m.Use(pongo22.Pongoer(pongo22.Options{
		TemplateFileSystem: binfiles.New(&teplatesFS, "templates"),
	}))

	m.Get("/", controllers.Home)
	
	// Run server in http port
	go m.Run(80)
}
```