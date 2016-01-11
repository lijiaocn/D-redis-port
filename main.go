//create: 2015/09/16 14:42:39 change: 2016/01/11 16:35:47 author:lijiao
package main

import(
	"log"

	"github.com/lijiaocn/GoPkgs/version"
	docopt "github.com/docopt/docopt-go"
)

func info(args map[string]interface{}){
	version.Show()
}

func init(){
	log.SetFlags(log.Llongfile)
}

func main() {
	  usage := `d-redis-port.
Usage:
	d-redis-port sync --config=CONFIGFILE 
	d-redis-port conf [--template=templatefile]
	d-redis-port info 
Options:
	-c, --config=CONFIGFILE       configfile, json
	-t, --template=templatefile   generate a template config file
	-v, --version                 show version
	`
	args, err := docopt.Parse(usage, nil ,true, "d-redis-port", false)
	if err != nil{
		log.Panic(err)
	}

	switch{
	case args["info"].(bool):
		info(args)
		return
	case args["conf"].(bool):
		conf(args)
		return
	case args["sync"].(bool):
		startSync(args)
		return
	}
}
