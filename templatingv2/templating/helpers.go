package templating

import "strings"

func PathToFSPath(p string) string {
	if p == "/" {
		return "."
	}

	p = strings.TrimPrefix(p, "/")
	p = strings.TrimSuffix(p, "/")

	return p
}

func FSPathToPath(p string) string {
	if p == "." {
		return "/"
	}

	p = strings.TrimPrefix(p, ".")

	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}

	if !strings.HasSuffix(p, "/") {
		p = p + "/"
	}

	return p
}
