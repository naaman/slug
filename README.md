# slug
Makes slugs

## Install

CLI:

```
go get github.com/naaman/slug/cmd/slug
```

Code:

```
go get github.com/naaman/slug
```

## Usage

CLI:

```term
$ slug -dir /path/to/src -app infinite-mesa-8755 -release
Initializing slug for /path/to/src...done
Archiving /path/to/src...done
Pushing /tmp/slug548040943.tgz....done 
Releasing...done (v148)
```

In code:

```term
myslug := slug.NewSlug(apiKey, appName, sourceDir)
myslug.Archive()
myslug.Push()
release := s.Release()
fmt.Printf("done (v%d)", release.Version)
```
