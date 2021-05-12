package discovery

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"os"
)


type ApiClient struct {
	DynClient dynamic.Interface
	ClientSet kubernetes.Interface
	Namespace string
}

var (
	settings = cli.New()
)

func NewHelmClient(kfcg KubConfigSetup, namespace string) (*ApiClient, error){
	actionConfig := new(action.Configuration)

	settings.KubeContext = kfcg.Context
	settings.KubeConfig  = kfcg.KubeConfigFile
	if namespace == "" {
		namespace = settings.Namespace()
	} else {
		kfcg.Namespace = settings.Namespace()
	}
	err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {
		log.Debug(fmt.Sprintf(format,v))
	})
	if err != nil {
		return nil, err
	}

	clset, err := actionConfig.KubernetesClientSet()
	if err != nil {
		return nil, err
	}
	restCfg, err := actionConfig.RESTClientGetter.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	dyn, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return nil, err
	}

	return &ApiClient{
		DynClient: dyn,
		ClientSet: clset,
		Namespace: namespace,
	}, err
}