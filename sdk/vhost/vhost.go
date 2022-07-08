package vhost

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const (
	PATH_PROVISIONER_DIR = "/etc/nginx/conf.d/provisioners"
	PATH_TEMPLATE_DIR    = "/etc/nginx/conf.d/templates"
	PATH_AVAILABLE_DIR   = "/etc/nginx/sites-available"
	PATH_ENABLED_DIR     = "/etc/nginx/sites-enabled"
	PATH_DATA_DIR        = "/home/jetrails"
)

func GetTemplatePath(templateName string) string {
	return path.Join(PATH_TEMPLATE_DIR, strings.TrimSuffix(path.Base(templateName), ".tpl")+".tpl")
}

func GetAvailablePath(domainName string) string {
	return path.Join(PATH_AVAILABLE_DIR, strings.TrimSuffix(path.Base(domainName), ".conf")+".conf")
}

func GetEnabledPath(domainName string) string {
	return path.Join(PATH_ENABLED_DIR, strings.TrimSuffix(path.Base(domainName), ".conf")+".conf")
}

func GetProvisionerPath(templateName string) string {
	return path.Join(PATH_PROVISIONER_DIR, strings.TrimSuffix(path.Base(templateName), ".sh")+".sh")
}

func BlockComment(input string, width int) string {
	lines := []string{}
	for len(input) > 0 {
		max := int(math.Min(float64(len(input)), float64(width-2)))
		lines = append(lines, "# "+input[:max])
		input = input[max:]
	}
	return strings.Join(lines, "\n")
}

func Create(templateName, domainName string, parameters map[string]string) error {
	if template, err := LoadTemplate(templateName); err == nil {
		input := TemplateInput(parameters)
		if site, err := template.Render(domainName, input); err == nil {
			if err := site.Save(false); err == nil {
				if err := ExecuteProvisioner(templateName, site.Input); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			return err
		}
	} else {
		return err
	}
	return nil
}

func Upgrade(domainName string, parameters map[string]string) error {
	availablePath := GetAvailablePath(domainName)
	if _, err := os.Stat(availablePath); os.IsNotExist(err) {
		return fmt.Errorf("could not find available site with given domain name %q", domainName)
	}
	if vh, err := LoadVirtualHost(domainName); err == nil {
		combined := TemplateInput{}
		for key, value := range vh.Input {
			combined[key] = value
		}
		for key, value := range parameters {
			combined[key] = value
		}
		if template, err := LoadTemplate(vh.Template.Name); err == nil {
			if template.Hash() == vh.Template.Hash() {
				return fmt.Errorf("most up-to-date template was used to generate this config")
			}
			if site, err := template.Render(domainName, combined); err == nil {
				if err := site.Save(true); err == nil {
					if err := ExecuteProvisioner(vh.Template.Name, combined); err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				return err
			}
		} else {
			return err
		}
	} else {
		return err
	}
	return nil
}

func Modify(domainName string, parameters map[string]string) error {
	availablePath := GetAvailablePath(domainName)
	if _, err := os.Stat(availablePath); os.IsNotExist(err) {
		return fmt.Errorf("could not find available site with given domain name %q", domainName)
	}
	if vh, err := LoadVirtualHost(domainName); err == nil {
		combined := TemplateInput{}
		for key, value := range vh.Input {
			combined[key] = value
		}
		for key, value := range parameters {
			combined[key] = value
		}
		if site, err := vh.Template.Render(domainName, combined); err == nil {
			if err := site.Save(true); err == nil {
				if err := ExecuteProvisioner(vh.Template.Name, combined); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			return err
		}
	} else {
		return err
	}
	return nil
}

func Delete(domainName string) error {
	availablePath := GetAvailablePath(domainName)
	enabledPath := GetEnabledPath(domainName)
	if _, err := os.Stat(availablePath); os.IsNotExist(err) {
		return fmt.Errorf("could not find available site with given domain name %q", domainName)
	}
	if _, err := os.Stat(enabledPath); !os.IsNotExist(err) {
		if err := os.Remove(enabledPath); err != nil {
			return fmt.Errorf("could not remove enabled site with domain name %q", domainName)
		}
	}
	if err := os.Remove(availablePath); err != nil {
		return fmt.Errorf("could not remove available site with domain name %q", domainName)
	}
	return nil
}

func Disable(domainName string) error {
	enabledPath := GetEnabledPath(domainName)
	if _, err := os.Stat(enabledPath); os.IsNotExist(err) {
		return fmt.Errorf("could not find site with name %q", domainName)
	}
	if err := os.Remove(enabledPath); err != nil {
		return fmt.Errorf("failed to disable site with name %q", domainName)
	}
	return nil
}

func Enable(domainName string) error {
	availablePath := GetAvailablePath(domainName)
	enabledPath := GetEnabledPath(domainName)
	if _, err := os.Stat(availablePath); os.IsNotExist(err) {
		return fmt.Errorf("could not find available site with name %q", domainName)
	}
	if err := os.Symlink(availablePath, enabledPath); err != nil {
		return fmt.Errorf("failed enabling site with name %q", domainName)
	}
	return nil
}

type VirtualHostStatus struct {
	VirtualHost VirtualHost
	Enabled     bool
}

func List() []VirtualHostStatus {
	hosts := []VirtualHostStatus{}
	files, _ := filepath.Glob(GetAvailablePath("*"))
	for _, availablePath := range files {
		if status := Info(path.Base(availablePath)); status != nil {
			hosts = append(hosts, *status)
		}
	}
	return hosts
}

func ListTemplates() []template {
	templates := []template{}
	files, _ := filepath.Glob(GetTemplatePath("*"))
	for _, templatePath := range files {
		if template, err := LoadTemplate(path.Base(templatePath)); err == nil {
			templates = append(templates, *template)
		}
	}
	return templates
}

func Info(domainName string) *VirtualHostStatus {
	enabledPath := GetEnabledPath(domainName)
	_, errStat := os.Stat(enabledPath)
	if vh, err := LoadVirtualHost(domainName); err == nil {
		vhs := VirtualHostStatus{
			VirtualHost: *vh,
			Enabled:     !os.IsNotExist(errStat),
		}
		return &vhs
	}
	return nil
}

func ExecuteProvisioner(templateName string, input TemplateInput) error {
	provisionerPath := GetProvisionerPath(templateName)
	if _, err := os.Stat(provisionerPath); os.IsNotExist(err) {
		return nil
	}
	cmd := exec.Command(provisionerPath)
	cmd.Env = input.GetEnvironmentalVars()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to execute provisioner for %s", templateName)
	}
	defer cmd.Wait()
	return nil
}
