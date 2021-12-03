package discovery

//KubConfigSetup holds user defined config
type KubConfigSetup struct {
	Context        string
	KubeConfigFile string
	Namespace      string
}
