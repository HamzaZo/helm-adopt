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


func (c Chart) Generate(t *discovery.ApiClient) error{
	name, err := utils.CreateChartDirectory(c.ChartName)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Join(name, ChartsDir), utils.DefaultPermission); err != nil {
		return err
	}

	err = populateDefaults(c.ChartName, name)
	if err != nil {
		return err
	}

	for n, rs := range c.Content {
		templateDir := filepath.Join(name, TemplatesDir)
		if _, err := os.Stat(filepath.Join(templateDir, n + ".yaml")); err == nil {
			fmt.Fprintf(os.Stderr, "Warning: File %q already exist", name)
		}
		if err := utils.WriteToFile(rs, filepath.Join(templateDir, n + ".yaml")); err != nil {
			return err
		}
	}

	manifest, templates := c.addTemplates()
	rel, err := c.buildRelease(templates, manifest, t.Namespace)
	if err != nil {
		return err
	}
	sec := driver.NewSecrets(t.ClientSet.CoreV1().Secrets(t.Namespace))
	sec.Log = func(format string, v ...interface{}) {
		log.Debug(fmt.Sprintf(format,v))
	}
	releases := storage.Init(sec)
	err = releases.Create(rel)
	if err != nil {
		return err
	}

	return nil
}



func (c Chart) buildRelease(template []*chart.File, manifest, namespace string) (*release.Release, error){

	chartFile, err := chartutil.LoadChartfile(filepath.Join(c.ChartName, ChartFileName))
	if err != nil {
		return nil, err
	}

	fChart := &chart.Chart{
		Metadata: chartFile,
		Templates: template,
	}
	info := &release.Info{
		FirstDeployed: helmtime.Now(),
		LastDeployed: helmtime.Now(),
		Status: release.StatusDeployed,
		Description: "commander release",
	}

	rels := &release.Release{
		Name: c.ReleaseName,
		Info: info,
		Chart: fChart,
		Manifest: manifest,
		Version: 1,
		Namespace: namespace,

	}
	return rels, err

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

func populateDefaults(chartName, chartPath string) error {
	d := defaulter{
		{
			path: filepath.Join(chartPath, ChartFileName),
			content: utils.Replacer(defaultChartfile, chartName, "<CHARTNAME>"),
		},
		{
			path: filepath.Join(chartPath, ValuesFileName),
			content: utils.Replacer(defaultValues, chartName, "<CHARTNAME>"),
		},
		{
			path: filepath.Join(chartPath, HelpersFileName),
			content: utils.Replacer(defaultHelpers, chartName, "<CHARTNAME>"),
		},
		{
			path: filepath.Join(chartPath, IgnoreFileName),
			content: []byte(defaultIgnore),
		},
	}

	for _, f := range d {
		if _, err := os.Stat(f.path); err == nil {
			fmt.Fprintf(os.Stderr, "Warning: File %q already exist", f.path)
		}
		if err := utils.WriteToFile(f.content, f.path); err != nil {
			return err
		}
	}

	return nil

}