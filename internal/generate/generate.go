package generate

import (
	"fmt"
	"github.com/HamzaZo/helm-adopt/internal/discovery"
	"github.com/HamzaZo/helm-adopt/internal/utils"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	helmtime "helm.sh/helm/v3/pkg/time"
	"io"
	"os"
	"path/filepath"
)

const (
	// ChartFileName is the default Chart file name.
	ChartFileName = "Chart.yaml"
	// ValuesFileName is the default values file name.
	ValuesFileName = "values.yaml"
	// TemplatesDir is the relative directory name for templates.
	TemplatesDir = "templates"
	// ChartsDir is the relative directory name for charts dependencies.
	ChartsDir = "charts"
	//HelpersFileName is the name of the helper file
	HelpersFileName = TemplatesDir + "/" + "_helpers.tpl"
	// IgnoreFileName is the name of the Helm ignore file.
	IgnoreFileName = ".helmignore"
)

type Chart struct {
	ChartName string
	ReleaseName string
	Content map[string][]byte
}

type defaulter []struct {
	path string
	content []byte
}

//Generate adopt k8s resource and generate helm chart
func (c Chart) Generate(client *discovery.ApiClient, out io.Writer, dryRun, debug bool) error{
	log.Infof("Generating chart %s\n", c.ChartName)

	if dryRun {
		log.Info("Adopting Resources on Dry-run mode..")
		for n := range c.Content {
			log.Infof("Added resouce as file %s into %s chart", n, c.ChartName)
		}
		log.Infof("Chart %s will be released as %v.1", c.ChartName, c.ReleaseName )
		return nil
	}
	name, err := utils.CreateChartDirectory(c.ChartName)
	if err != nil {
		return err
	}
	utils.DebugPrinter("CHART PATH: %s", debug, out ,name)

	if err = os.MkdirAll(filepath.Join(name, ChartsDir), utils.DefaultPermission); err != nil {
		return err
	}

	err = populateDefaults(c.ChartName, name, out)
	if err != nil {
		return err
	}

	templateDir := filepath.Join(name, TemplatesDir)

	for n, ct := range c.Content {
		if _, err = os.Stat(filepath.Join(templateDir, n + ".yaml")); err == nil {
			fmt.Fprintf(out, "WARNING: File %q already exists Overwriting.\n", name)
		}
		if err := utils.WriteToFile(ct, filepath.Join(templateDir, n + ".yaml")); err != nil {
			return err
		}
		log.Infof("Added resource as file %s into %s chart", n, c.ChartName)
	}

	err = c.createRelease(client, debug, out)
	if err != nil {
		return err
	}

	return nil
}

//createRelease create a helm release using secret driver
func (c Chart) createRelease(client *discovery.ApiClient, debug bool, out io.Writer) error{
	manifest, templates := c.addTemplates()

	rel, err := c.buildRelease(templates, manifest, client.Namespace)
	if err != nil {
		return err
	}
	sec := driver.NewSecrets(client.ClientSet.CoreV1().Secrets(client.Namespace))
	sec.Log = func(format string, v ...interface{}) {
		log.Debug(fmt.Sprintf(format,v))
	}

	releases := storage.Init(sec)

	err = releases.Create(rel)
	if err != nil {
		return err
	}
	log.Infof("Chart %s is released as %v.1", c.ChartName, c.ReleaseName )

	if debug {
		utils.DebugPrinter("Adopting %v resource(s) into %s chart:", debug, out, len(templates), c.ChartName)
		utils.DebugPrinter("MANIFEST:", debug, out)
		fmt.Fprint(out, manifest)
	}

	return nil
}

func (c Chart) buildRelease(template []*chart.File, manifest, namespace string) (*release.Release, error){

	chartFile, err := chartutil.LoadChartfile(filepath.Join(c.ChartName, ChartFileName))
	if err != nil {
		return nil, err
	}

	ch := &chart.Chart{
		Metadata: chartFile,
		Templates: template,
	}
	info := &release.Info{
		FirstDeployed: helmtime.Now(),
		LastDeployed: helmtime.Now(),
		Status: release.StatusDeployed,
		Description: "adopt k8s resources into helm chart",
	}

	return &release.Release{
		Name: c.ReleaseName,
		Info: info,
		Chart: ch,
		Manifest: manifest,
		Version: 1,
		Namespace: namespace,
	}, nil

}


func (c Chart) addTemplates() (string, []*chart.File) {
	var templates []*chart.File
	manifest := ""
	for name, content := range c.Content {
		filename := fmt.Sprintf("templates/%v.yaml", name)

		templates = append(templates, &chart.File{
			Name: filename,
			Data: content,
		})
		manifest += fmt.Sprintf("\n---\n# Source: %v\n%v", filename, string(content))
	}

	return manifest, templates
}

func populateDefaults(chartName, chartPath string, out io.Writer) error {
	d := defaulter{
		{
			path: filepath.Join(chartPath, ChartFileName),
			content: utils.ReplaceStr(defaultChartfile, chartName, "<CHARTNAME>"),
		},
		{
			path: filepath.Join(chartPath, ValuesFileName),
			content: utils.ReplaceStr(defaultValues, chartName, "<CHARTNAME>"),
		},
		{
			path: filepath.Join(chartPath, HelpersFileName),
			content: utils.ReplaceStr(defaultHelpers, chartName, "<CHARTNAME>"),
		},
		{
			path: filepath.Join(chartPath, IgnoreFileName),
			content: []byte(defaultIgnore),
		},
	}

	for _, f := range d {
		if _, err := os.Stat(f.path); err == nil {
			fmt.Fprintf(out, "Warning: File %q already exist Overwriting \n", f.path)
		}
		if err := utils.WriteToFile(f.content, f.path); err != nil {
			return err
		}
	}

	return nil

}