package vhost

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	gotemplate "text/template"

	"gopkg.in/yaml.v3"

	cp "github.com/otiai10/copy"
)

type TemplateInput map[string]string
type TemplateOutput map[string][]byte

type Template struct {
	Name        string
	Files       map[string][]byte
	Provisioner []byte
	InputSchema InputSchema
}

var (
	ErrTemplateNotFound       = errors.New("template not found")
	ErrTemplateStructure      = errors.New("template structure invalid")
	ErrTemplateFailedToWalk   = errors.New("failed to read template data")
	ErrSitesAvailableNotFound = errors.New(`template missing required sites-available file`)
)

func GetTemplates() []Template {
	templates := []Template{}
	templateDirs, errReadDir := ioutil.ReadDir(PATH_TEMPLATES_DIR)
	if errReadDir != nil {
		return templates
	}
	for _, templateDir := range templateDirs {
		template, errLoad := LoadTemplate(templateDir.Name())
		if errLoad == nil {
			templates = append(templates, template)
		}
	}
	return templates
}

func LoadTemplate(name string) (Template, error) {
	templateDir := path.Join(PATH_TEMPLATES_DIR, name)
	stat, errStat := os.Stat(templateDir)
	if errStat != nil {
		return Template{}, ErrTemplateNotFound
	}
	if !stat.IsDir() {
		return Template{}, ErrTemplateStructure
	}
	provisioner := []byte{}
	schema := []byte{}
	files := map[string][]byte{}
	errWalk := filepath.Walk(templateDir, func(fullFilePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && fullFilePath == filepath.Join(templateDir, "assets") {
			return filepath.SkipDir
		}
		if !info.IsDir() {
			fileContents, errRead := ioutil.ReadFile(fullFilePath)
			if errRead != nil {
				return errRead
			}
			switch fullFilePath {
			case filepath.Join(templateDir, "provisioner.sh"):
				provisioner = fileContents
			case filepath.Join(templateDir, "schema.yaml"):
				schema = fileContents
			default:
				files[strings.TrimPrefix(fullFilePath, templateDir+"/")] = fileContents
			}
		}
		return nil
	})
	if errWalk != nil {
		return Template{}, ErrTemplateFailedToWalk
	}
	availableSeen := false
	if _, ok := files["sites-available/[site_name].conf.tpl"]; !ok {
		availableSeen = true
	}
	if _, ok := files["sites-available/[site_name].conf"]; !ok {
		availableSeen = true
	}
	if !availableSeen {
		return Template{}, ErrSitesAvailableNotFound
	}
	parsedInputSchema, errParse := ParseInputSchema(schema)
	if errParse != nil {
		return Template{}, errParse
	}
	template := Template{
		Name:        name,
		Files:       files,
		Provisioner: provisioner,
		InputSchema: parsedInputSchema,
	}
	return template, nil
}

func RenderFilePath(content string, input TemplateInput) string {
	for fileName := range input {
		content = strings.ReplaceAll(content, "["+fileName+"]", input[fileName])
	}
	return content
}

func RenderTemplate(content []byte, input TemplateInput) ([]byte, error) {
	tmpl, errParse := gotemplate.New("").Parse(string(content))
	if errParse != nil {
		return []byte{}, errParse
	}
	rendered := bytes.Buffer{}
	errExecute := tmpl.Execute(&rendered, input)
	if errExecute != nil {
		return []byte{}, errExecute
	}
	return rendered.Bytes(), nil
}

func (t Template) Render(siteName string, input TemplateInput) (TemplateOutput, error) {
	input["site_name"] = siteName
	output := TemplateOutput{}
	for fileName, fileContents := range t.Files {
		renderedFileName := RenderFilePath(fileName, input)
		if strings.HasSuffix(fileName, ".tpl") {
			renderedContent, errRenderContent := RenderTemplate(fileContents, input)
			if errRenderContent != nil {
				return output, errRenderContent
			}
			output[strings.TrimSuffix(string(renderedFileName), ".tpl")] = renderedContent
		} else {
			output[string(renderedFileName)] = fileContents
		}
	}
	return output, nil
}

func (o TemplateOutput) SortedKeys() []string {
	keys := []string{}
	for key := range o {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (i TemplateInput) SortedKeys() []string {
	keys := []string{}
	for key := range i {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (t Template) Exists() bool {
	return t.Name != ""
}

func (t Template) Hash() string {
	hash := []byte(t.Name)
	hash = append(hash, t.Provisioner...)
	hash = append(hash, t.InputSchema.Hash()...)
	sortedKeys := []string{}
	for fileName := range t.Files {
		sortedKeys = append(sortedKeys, fileName)
	}
	sort.Strings(sortedKeys)
	for _, fileName := range sortedKeys {
		fileContents := t.Files[fileName]
		hash = append(hash, []byte(fileName)...)
		hash = append(hash, fileContents...)
	}
	return HashData(hash)
}

func (o TemplateOutput) Hash() string {
	hash := []byte{}
	for _, fileName := range o.SortedKeys() {
		fileContents := o[fileName]
		hash = append(hash, []byte(fileName)...)
		hash = append(hash, fileContents...)
	}
	return HashData(hash)
}

func (i TemplateInput) Hash() string {
	hash := []byte{}
	for _, inputName := range i.SortedKeys() {
		inputValue := i[inputName]
		hash = append(hash, []byte(inputName)...)
		hash = append(hash, inputValue...)
	}
	return HashData(hash)
}

func (i TemplateInput) Yaml() []byte {
	data, _ := yaml.Marshal(i)
	return data
}

func (o TemplateOutput) DeleteFiles(silent bool) error {
	for fileName := range o {
		filePath := path.Join(PATH_NGINX_DIR, fileName)
		err := os.Remove(filePath)
		if !silent && err != nil {
			return err
		}
	}
	return nil
}

func (o TemplateOutput) Save() error {
	for fileName, fileContents := range o {
		filePath := path.Join(PATH_NGINX_DIR, fileName)
		filePathDir := path.Dir(filePath)
		errDir := os.MkdirAll(filePathDir, 0755)
		if errDir != nil {
			return errDir
		}
		errFile := ioutil.WriteFile(filePath, fileContents, 0644)
		if errFile != nil {
			return errFile
		}
		if strings.TrimSpace(string(fileContents)) == "" {
			errRemove := os.Remove(filePath)
			if errRemove != nil {
				return errRemove
			}
		}
	}
	return nil
}

func (input TemplateInput) GetEnvironmentalVars() []string {
	vars := []string{}
	for key, value := range input {
		re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
		newKey := re.ReplaceAllString(strings.ToUpper(key), "_")
		vars = append(vars, newKey+"="+value)
	}
	return vars
}

func (t Template) RunProvisioner(siteName string, input TemplateInput, becomeUser string) error {
	validatedInput, errValidate := t.InputSchema.Validate(input)
	if errValidate != nil {
		return errValidate
	}
	validatedInput["site_name"] = siteName
	tmpDir, errMkdir := os.MkdirTemp("", "vhost-provisioner-")
	if errMkdir != nil {
		return errMkdir
	}
	defer os.RemoveAll(tmpDir)
	errChmod := os.Chmod(tmpDir, 0777)
	if errChmod != nil {
		return errChmod
	}
	assetsDir := path.Join(PATH_TEMPLATES_DIR, t.Name, "assets")
	if _, errStat := os.Stat(assetsDir); errStat == nil {
		errCopy := cp.Copy(assetsDir, path.Join(tmpDir, "assets"))
		if errCopy != nil {
			return errCopy
		}
	}
	provisionerPath := path.Join(tmpDir, "provisioner.sh")
	errWrite := ioutil.WriteFile(provisionerPath, t.Provisioner, 0755)
	if errWrite != nil {
		return errWrite
	}
	cmdMain := provisionerPath
	cmdArgs := []string{}
	if becomeUser != "" {
		cmdMain = "sudo"
		cmdArgs = []string{"-E", "-u", becomeUser, provisionerPath}
	}
	cmd := exec.Command(cmdMain, cmdArgs...)
	cmd.Env = []string{
		"HOME=" + os.Getenv("HOME"),
		"PATH=" + os.Getenv("PATH"),
	}
	cmd.Env = append(cmd.Env, validatedInput.GetEnvironmentalVars()...)
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
