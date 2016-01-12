package env

import (
	"fmt"
	"os"
)

var envvars []*envvar
var prefix string

type envvar struct {
	name string
	dest *string
	def  string
	desc string
}

func (e *envvar) read() {
	value := os.Getenv(prefix + e.name)
	if value == "" {
		value = e.def
	}
	*e.dest = value
}

func Var(dest *string, name string, def string, desc string) {
	v := &envvar{
		name: name,
		dest: dest,
		def:  def,
		desc: desc,
	}

	envvars = append(envvars, v)
}

func Parse(p string, allowArgs bool) {
	if prefix = p; prefix != "" {
		prefix = prefix + "_"
	}
	if len(os.Args) > 1 && allowArgs == false {
		printHelpAndExit()
	}
	for _, e := range envvars {
		e.read()
	}
}

func printHelpAndExit() {
	fmt.Printf("'%s' needs to be configured via environment variables and does not take any arguments.\n", os.Args[0])
	fmt.Println("The following environment variables are read ('_nil_' implies an empty string):\n")
	for _, e := range envvars {
		def := e.def
		if def == "" {
			def = "_nil_"
		} else {
			def = "'" + def + "'"
		}
		fmt.Printf("\t%s%s=%s: %s\n", prefix, e.name, def, e.desc)
	}
	os.Exit(1)
}
