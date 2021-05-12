package cmd

import (
	"github.com/HamzaZo/helm-adopt/internal/discovery"
	"github.com/HamzaZo/helm-adopt/internal/generate"
	"github.com/HamzaZo/helm-adopt/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"io"
)

var (
	chartDir string
	releaseName string
	dryRun bool
	debug bool
)

type EnvSettings struct {
	KubeConfigFile string
	KubeContext    string
	Namespace      string
}

func NewResourcesCmd(_ io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use: "resources <pluralKind>:<name>",
		Short: "adopt k8s resources into a helm chart",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runResources(args)
			if err != nil{
				return err
			}
			return nil
		},

	}
	flags := cmd.Flags()

	flags.StringVarP(&chartDir, "output", "o", "", "Specify the chart directory of loaded yaml files")
	flags.StringVarP(&releaseName, "release", "r", "", "Specify the name for the generated release")
	flags.BoolVar(&dryRun, "dry-run", false, "print what resources will be adopted ")
	flags.BoolVar(&debug, "debug", false, "show the generated manifests on STDOUT")

	Settings.AddFlags(flags)

	return cmd
}

// AddFlags binds flags to the given flagset
func (e *EnvSettings) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&e.KubeConfigFile, "kubeconfig", "k", "", "path to the kubeconfig file")
	fs.StringVarP(&e.KubeContext, "kube-context", "c", e.KubeContext, "name of the kubeconfig context to use")
	fs.StringVarP(&e.Namespace, "namespace", "n", e.Namespace, "namespace scope for this request")

}


func runResources(args []string) error{
	err := utils.ChartValidator(chartDir,releaseName)
	if err != nil {
		return err
	}

	kubeconfig := discovery.KubConfigSetup{
		Context: Settings.KubeContext,
		KubeConfigFile: Settings.KubeConfigFile,
		Namespace: Settings.Namespace,
	}
	helmClient, err := discovery.NewHelmClient(kubeconfig, kubeconfig.Namespace)
	if err != nil {
		return err
	}
	input, err := utils.GetAllArgs(args)
	if err != nil {
		return err
	}
	content, err := fetchResources(helmClient, input)
	if err != nil {
		return err
	}
	chart := &generate.Chart{
		ChartName: chartDir,
		ReleaseName: releaseName,
		Content: content,
	}

	err = chart.Generate(helmClient)
	if err != nil {
		return err
	}

	return nil
}


func fetchResources(client *discovery.ApiClient, input map[string][]string) (map[string][]byte, error){
	var output map[string][]byte

	namespaceResource, clusterResource ,err := discovery.FetchedFilteredResources(client, input)
	if err != nil {
		return nil, err
	}

	ns , err := namespaceResource.Query(client, client.Namespace)
	if err != nil {
		return nil, err
	}

	cls , err := clusterResource.Query(client, "")
	if err != nil {
		return nil, err
	}
	output = utils.MergeMapsBytes(ns, cls)

	return output, err
}