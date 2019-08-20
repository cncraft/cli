package manifestparser

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-cli/director/template"
	"gopkg.in/yaml.v2"
)

type ParsedManifest struct {
	Applications []Application

	pathToManifest string
	rawManifest    []byte
	validators     []validatorFunc
	hasParsed      bool
}

func NewParser() *ParsedManifest {
	return new(ParsedManifest)
}

func (m ParsedManifest) AppNames() []string {
	var names []string
	for _, app := range m.Applications {
		names = append(names, app.Name)
	}
	return names
}

func (m ParsedManifest) Apps() []Application {
	return m.Applications
}

func (m ParsedManifest) ContainsManifest() bool {
	return m.hasParsed
}

func (m ParsedManifest) ContainsMultipleApps() bool {
	return len(m.Applications) > 1
}

func (m ParsedManifest) ContainsPrivateDockerImages() bool {
	for _, app := range m.Applications {
		if app.Docker != nil && app.Docker.Username != "" {
			return true
		}
	}
	return false
}

func (m ParsedManifest) FullRawManifest() []byte {
	return m.rawManifest
}

func (m ParsedManifest) GetPathToManifest() string {
	return m.pathToManifest
}

func (m ParsedManifest) GetParsedManifest() ParsedManifest {
	return m
}

func (m ParsedManifest) HasAppWithNoName() bool {
	for _, app := range m.Applications {
		if app.Name == "" {
			return true
		}
	}
	return false
}
// InterpolateAndParse reads the manifest at the provided paths, interpolates
// variables if a vars file is provided, and sets the current manifest to the
// resulting manifest.
// For manifests with only 1 application, appName will override the name of the
// single app defined.
// For manifests with multiple applications, appName will filter the
// applications and leave only a single application in the resulting parsed
// manifest structure.
func (m *ParsedManifest) InterpolateAndParse(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV, appName string) error {
	rawManifest, err := ioutil.ReadFile(pathToManifest)
	if err != nil {
		return err
	}

	tpl := template.NewTemplate(rawManifest)
	fileVars := template.StaticVariables{}

	for _, path := range pathsToVarsFiles {
		rawVarsFile, ioerr := ioutil.ReadFile(path)
		if ioerr != nil {
			return ioerr
		}

		var sv template.StaticVariables

		err = yaml.Unmarshal(rawVarsFile, &sv)
		if err != nil {
			return InvalidYAMLError{Err: err}
		}

		for k, v := range sv {
			fileVars[k] = v
		}
	}

	for _, kv := range vars {
		fileVars[kv.Name] = kv.Value
	}

	rawManifest, err = tpl.Evaluate(fileVars, nil, template.EvaluateOpts{ExpectAllKeys: true})
	if err != nil {
		return InterpolationError{Err: err}
	}

	m.pathToManifest = pathToManifest
	return m.parse(rawManifest, appName)
}

func (m ParsedManifest) RawAppManifest(appName string) ([]byte, error) {
	var appManifest manifest
	for _, app := range m.Applications {
		if app.Name == appName {
			appManifest.Applications = []Application{app}
			return yaml.Marshal(appManifest)
		}
	}
	return nil, AppNotInManifestError{Name: appName}
}

func (m ParsedManifest) RawManifest() ([]byte, error) {
	return yaml.Marshal(m)
}

func (m *ParsedManifest) parse(manifestBytes []byte, appName string) error {
	m.rawManifest = manifestBytes
	pathToManifest := m.GetPathToManifest()
	var raw manifest

	err := yaml.Unmarshal(manifestBytes, &raw)
	if err != nil {
		return err
	}

	if len(raw.Applications) == 0 {
		return errors.New("must have at least one application")
	}

	fmt.Printf("RAW: %v", raw.Applications)

	for i := range raw.Applications {
		if raw.Applications[i].Path == "" {
			continue
		}

		var finalPath = raw.Applications[i].Path
		if !filepath.IsAbs(finalPath) {
			finalPath = filepath.Join(filepath.Dir(pathToManifest), finalPath)
		}
		finalPath, err = filepath.EvalSymlinks(finalPath)
		if err != nil {
			if os.IsNotExist(err) {
				return InvalidManifestApplicationPathError{
					Path: raw.Applications[i].Path,
				}
			}
			return err
		}
		raw.Applications[i].Path = finalPath
	}

	m.Applications = raw.Applications
	m.rawManifest, err = yaml.Marshal(raw)
	if err != nil {
		return err
	}

	m.hasParsed = true
	return nil
}

func (m *ParsedManifest) GetFirstAppWebProcess() *ProcessModel {
	for i, process := range m.Applications[0].Processes {
		if process.Type == "web" {
			return &m.Applications[0].Processes[i]
		}
	}

	return nil
}

func (m *ParsedManifest) GetFirstApp() *Application {
	return &m.Applications[0]
}

func (m *ParsedManifest) UpdateFirstAppWebProcess(updateFunc func(process *ProcessModel)) {
	for i, process := range m.Applications[0].Processes {
		if process.Type == "web" {
			updateFunc(&m.Applications[0].Processes[i])
			break
		}
	}
}
