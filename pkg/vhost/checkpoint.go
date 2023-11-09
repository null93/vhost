package vhost

import (
	"fmt"
	"os"
	"path"

	"archive/tar"
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
)

const PATTERN_CHECKPOINT_FILE = `^(.+?)_([0-9_-]+).state$`

type CheckPoint struct {
	Revision    int
	SiteName    string
	Timestamp   time.Time
	Template    Template
	Input       TemplateInput
	Output      TemplateOutput
	Description string
}

type TarFile struct {
	Header *tar.Header
	Body   []byte
}

func GetCheckPoints(name string) ([]CheckPoint, error) {
	checkPoints := []CheckPoint{}
	matches, errGlob := filepath.Glob(path.Join(PATH_CHECKPOINTS_DIR, fmt.Sprintf("%s_*.state", name)))
	if errGlob != nil {
		return checkPoints, errGlob
	}
	for _, match := range matches {
		filename := path.Base(match)
		checkPoint, errLoad := LoadCheckPoint(filename)
		if errLoad != nil {
			return checkPoints, errLoad
		}
		checkPoints = append(checkPoints, checkPoint)
	}
	return checkPoints, nil
}

func GetLatestCheckPoint(name string) (CheckPoint, error) {
	checkPoints, errGetCheckPoints := GetCheckPoints(name)
	if errGetCheckPoints != nil {
		return CheckPoint{}, errGetCheckPoints
	}
	latest := CheckPoint{Revision: 0}
	for _, checkPoint := range checkPoints {
		if checkPoint.Revision > latest.Revision {
			latest = checkPoint
		}
	}
	return latest, nil
}

func GetCheckPoint(name string, revision int) (CheckPoint, error) {
	checkPointFileName := fmt.Sprintf("%s_%d.state", name, revision)
	return LoadCheckPoint(checkPointFileName)
}

func NewCheckPoint(siteName string, template Template, rawInput TemplateInput) (CheckPoint, error) {
	input, err := template.InputSchema.Validate(rawInput)
	if err != nil {
		return CheckPoint{}, err
	}
	output, err := template.Render(siteName, input)
	if err != nil {
		return CheckPoint{}, err
	}
	for key, schema := range template.InputSchema {
		if schema.ProvisionerOnly {
			delete(input, key)
		}
	}
	checkpoint := CheckPoint{
		Revision:    1,
		SiteName:    siteName,
		Timestamp:   time.Now(),
		Template:    template,
		Input:       input,
		Output:      output,
		Description: "-",
	}
	return checkpoint, nil
}

func (t CheckPoint) Serialize() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(t)
	if err != nil {
		return nil, err
	}
	base64Encoded := make([]byte, base64.StdEncoding.EncodedLen(len(buffer.Bytes())))
	base64.StdEncoding.Encode(base64Encoded, buffer.Bytes())
	blockedResult := []byte{}
	blockSize := 76
	for len(base64Encoded) > blockSize {
		blockedResult = append(blockedResult, base64Encoded[0:blockSize]...)
		base64Encoded = base64Encoded[blockSize:]
		blockedResult = append(blockedResult, []byte("\n")...)
	}
	blockedResult = append(blockedResult, base64Encoded...)
	return blockedResult, nil
}

func DeserializeCheckPoint(serialized []byte) (CheckPoint, error) {
	base64Decoded := make([]byte, base64.StdEncoding.DecodedLen(len(serialized)))
	_, err := base64.StdEncoding.Decode(base64Decoded, serialized)
	if err != nil {
		return CheckPoint{}, err
	}
	var checkpoint CheckPoint
	decoder := gob.NewDecoder(strings.NewReader(string(base64Decoded)))
	err = decoder.Decode(&checkpoint)
	if err != nil {
		return CheckPoint{}, err
	}
	return checkpoint, nil
}

func (t CheckPoint) GetFileName() string {
	return fmt.Sprintf("%s_%d.state", t.SiteName, t.Revision)
}

func (t CheckPoint) Save() error {
	serialized, err := t.Serialize()
	if err != nil {
		return err
	}
	filename := t.GetFileName()
	checkpointPath := path.Join(PATH_CHECKPOINTS_DIR, filename)
	return ioutil.WriteFile(checkpointPath, serialized, 0600)
}

func (t CheckPoint) GetTarFiles() ([]TarFile, error) {
	tarFiles := []TarFile{}
	for filePath, fileContents := range t.Template.Files {
		header := &tar.Header{
			Name: path.Join("templates", t.Template.Name, filePath),
			Mode: 0644,
			Size: int64(len(fileContents)),
		}
		tarFiles = append(tarFiles, TarFile{Header: header, Body: fileContents})
	}
	provisionerHeader := &tar.Header{
		Name: path.Join("templates", t.Template.Name, "provisioner.sh"),
		Mode: 0755,
		Size: int64(len(t.Template.Provisioner)),
	}
	tarFiles = append(tarFiles, TarFile{Header: provisionerHeader, Body: t.Template.Provisioner})
	schemaHeader := &tar.Header{
		Name: path.Join("templates", t.Template.Name, "schema.yaml"),
		Mode: 0755,
		Size: int64(len(t.Template.InputSchema.Yaml())),
	}
	tarFiles = append(tarFiles, TarFile{Header: schemaHeader, Body: t.Template.InputSchema.Yaml()})
	for filePath, fileContents := range t.Output {
		header := &tar.Header{
			Name: path.Join("sites", t.SiteName, filePath),
			Mode: 0644,
			Size: int64(len(fileContents)),
		}
		tarFiles = append(tarFiles, TarFile{Header: header, Body: fileContents})
	}
	inputHeader := &tar.Header{
		Name: path.Join("input.yaml"),
		Mode: 0644,
		Size: int64(len(t.Input.Yaml())),
	}
	tarFiles = append(tarFiles, TarFile{Header: inputHeader, Body: t.Input.Yaml()})
	return tarFiles, nil
}

func LoadCheckPoint(checkPointFileName string) (CheckPoint, error) {
	checkpointPath := path.Join(PATH_CHECKPOINTS_DIR, checkPointFileName)
	serialized, err := ioutil.ReadFile(checkpointPath)
	if err != nil {
		return CheckPoint{}, err
	}
	return DeserializeCheckPoint(serialized)
}

func PurgeCheckPoints(siteName string, force bool) error {
	matches, err := filepath.Glob(path.Join(PATH_CHECKPOINTS_DIR, fmt.Sprintf("%s_*.state", siteName)))
	if err != nil {
		return err
	}
	for _, match := range matches {
		if err := os.Remove(match); err != nil {
			if !force {
				return err
			}
		}
	}
	return nil
}
