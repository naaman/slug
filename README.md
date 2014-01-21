# slug
Makes slugs

## Usage

```term
myslug := slug.NewSlug(apiKey, appName, sourceDir)
myslug.Archive()
myslug.Push()
release := s.Release()
fmt.Printf("done (v%d)", release.Version)
```