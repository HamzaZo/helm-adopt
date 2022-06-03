package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/chartutil"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sigs.k8s.io/yaml"
	"strings"
)

var chartName = regexp.MustCompile("^[a-zA-Z0-9._-]+$")

const (
	maxChartNameLength = 250
	DefaultPermission  = 0755
)

//GetPrettyYaml generate a well formatted yaml file
func GetPrettyYaml(obj interface{}) ([]byte, error) {
	var prettyJSON bytes.Buffer
	output, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	err = json.Indent(&prettyJSON, output, "", "  ")
	if err != nil {
		return nil, err
	}
	value, err := yaml.JSONToYAML(prettyJSON.Bytes())
	if err != nil {
		return nil, err
	}

	return value, err
}

func MergeMapsBytes(m1, m2 map[string][]byte) map[string][]byte {
	output := make(map[string][]byte, len(m1))
	for k, v := range m1 {
		output[k] = v
	}
	for i, j := range m2 {
		output[i] = j
	}
	return output
}

func CreateChartDirectory(name string) (string, error) {
	dir, err := os.Stat(name)
	if os.IsNotExist(err) {
		err := os.Mkdir(name, DefaultPermission)
		if err != nil {
			log.Errorf("unable to create chart directory %v", err)
			return "", err
		}

	} else if !dir.IsDir() {
		log.Errorf("%s is not a directory", name)
		return "", err
	}
	path, err := filepath.Abs(name)
	if err != nil {
		return "", err
	}

	return path, err
}

func ChartValidator(chart, release string) error {
	if chart == "" || len(chart) > maxChartNameLength {
		return fmt.Errorf("chart name must be between 1 and %d characters", maxChartNameLength)
	}
	if !chartName.MatchString(chart) {
		return fmt.Errorf("chart name must match the regular expression %q", chartName.String())
	}
	err := chartutil.ValidateReleaseName(release)
	if err != nil {
		return err
	}
	return nil
}

func ReplaceStr(src, newStr, old string) []byte {
	return []byte(strings.ReplaceAll(src, old, newStr))
}

func WriteToFile(content []byte, name string) error {
	if err := os.MkdirAll(filepath.Dir(name), DefaultPermission); err != nil {
		return err
	}
	if err := ioutil.WriteFile(name, content, 0644); err != nil {
		return err
	}
	return nil
}

func Contains(m map[string][]string, val string) ([]string, bool) {
	if value, ok := m[val]; ok {
		return value, true
	}
	return nil, false
}

func GetAllArgs(set []string) (map[string][]string, error) {
	m := make(map[string][]string)
	return func() (map[string][]string, error) {
		o, err := processingArgs(set, m)
		if err != nil {
			return nil, err
		}
		return o, nil
	}()
}

func processingArgs(set []string, output map[string][]string) (map[string][]string, error) {
	var list []string

	for _, k := range set {

		switch p := strings.SplitN(k, ":", 1); {
		case !strings.HasSuffix(p[0], "s"):
			return nil, fmt.Errorf("resource %v must be plural", p[0])
		case strings.Contains(p[1], ","):
			e := strings.Split(p[1], ",")
			for _, i := range e {
				list = append(list, i)
				output[p[0]] = list
			}
			list = nil
		case !strings.Contains(p[1], ","):
			output[p[0]] = []string{p[1]}

		}
	}

	return output, nil
}

func DebugPrinter(format string, debug bool, out io.Writer, v ...interface{}) {
	if debug {
		format = fmt.Sprintf("[debug] %s\n", format)
		fmt.Fprintf(out, fmt.Sprintf(format, v...))
	}
}
