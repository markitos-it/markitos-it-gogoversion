//#:[.'.]:>-==================================================================================
//#:[.'.]:>- Marco Antonio - markitos devsecops kulture
//#:[.'.]:>- The Way of the Artisan
//#:[.'.]:>- markitos.es.info@gmail.com
//#:[.'.]:>- 🌍 https://github.com/orgs/markitos-it/repositories
//#:[.'.]:>- 🌍 https://github.com/orgs/markitos-public/repositories
//#:[.'.]:>- 📺 https://www.youtube.com/@markitos_devsecops
//#:[.'.]:>- =================================================================================

package main

import "flag"

type Config struct {
	RepoPath    string
	DryRun      bool
	NoTag       bool
	NoChangelog bool
}

func newConfig() Config {
	cfg := Config{}
	flag.StringVar(&cfg.RepoPath,    "path",          ".",   "Ruta al repositorio")
	flag.BoolVar(&cfg.DryRun,        "dry-run",       false, "Muestra la versión sin escribir nada")
	flag.BoolVar(&cfg.NoTag,         "no-tag",        false, "No crea el tag git")
	flag.BoolVar(&cfg.NoChangelog,   "no-changelog",  false, "No escribe CHANGELOG.md")
	flag.Parse()
	return cfg
}