package discovery

//KubConfigSetup hold user defined config
type KubConfigSetup struct {
	Context        string
	KubeConfigFile string
	Namespace      string
}