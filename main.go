package main

import (
	"github.com/HamzaZo/helm-adopt/cmd"
	_ "k8s.io/client-go/plugin/pkg/client/auth" //required for auth
	"os"
)

func main(){
	v := cmd.NewRootCmd(os.Stdout, os.Args[1:])
	if err := v.Execute(); err != nil {
		os.Exit(1)
	}
}