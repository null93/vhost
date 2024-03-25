package vhost

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type State string

const (
	SITE_ENABLED  State = "enabled"
	SITE_DISABLED State = "disabled"
)

var (
	ErrSiteNotFound = fmt.Errorf("failed to find site")
)

type Site struct {
	Name             string
	State            State
	LatestCheckPoint CheckPoint
}

func GetSites() ([]Site, error) {
	sites := []Site{}
	globPattern := path.Join(PATH_NGINX_AVAILABLE_DIR, "*.conf")
	matches, errGlob := filepath.Glob(globPattern)
	if errGlob != nil {
		return sites, errGlob
	}
	for _, match := range matches {
		siteName := strings.TrimSuffix(filepath.Base(match), filepath.Ext(match))
		site := Site{
			Name:  siteName,
			State: SITE_DISABLED,
		}
		if _, err := os.Stat(path.Join(PATH_NGINX_ENABLED_DIR, siteName+".conf")); err == nil {
			site.State = SITE_ENABLED
		}
		latestCheckPoint, errLatest := GetLatestCheckPoint(siteName)
		if errLatest == nil {
			site.LatestCheckPoint = latestCheckPoint
		}
		sites = append(sites, site)
	}
	return sites, nil
}

func GetSite(siteName string) (Site, error) {
	globPattern := path.Join(PATH_NGINX_AVAILABLE_DIR, siteName+".conf")
	matches, errGlob := filepath.Glob(globPattern)
	if errGlob != nil {
		return Site{}, errGlob
	}
	if len(matches) == 0 {
		return Site{}, ErrSiteNotFound
	}
	site := Site{
		Name:  siteName,
		State: SITE_DISABLED,
	}
	if _, err := os.Stat(path.Join(PATH_NGINX_ENABLED_DIR, siteName+".conf")); err == nil {
		site.State = SITE_ENABLED
	}
	latestCheckPoint, errLatest := GetLatestCheckPoint(siteName)
	if errLatest == nil {
		site.LatestCheckPoint = latestCheckPoint
	}
	return site, nil
}

func SiteExists(siteName string) bool {
	globPattern := path.Join(PATH_CHECKPOINTS_DIR, fmt.Sprintf("%s_*.state", siteName))
	matches, err := filepath.Glob(globPattern)
	if err != nil {
		return false
	}
	if len(matches) > 0 {
		return true
	}
	if _, err := os.Stat(path.Join(PATH_NGINX_AVAILABLE_DIR, siteName+".conf")); err == nil {
		return true
	}
	return false
}

func DeleteSite(siteName string, silent bool) error {
	latestCheckPoint, errLatest := GetLatestCheckPoint(siteName)
	if errLatest != nil {
		return errLatest
	}
	if err := latestCheckPoint.Output.DeleteFiles(silent); !silent && err != nil {
		return err
	}
	if err := os.Remove(path.Join(PATH_NGINX_ENABLED_DIR, siteName+".conf")); err != nil {
		if !silent {
			return err
		}
	}
	return nil
}

func DisableSite(siteName string) error {
	if err := os.Remove(path.Join(PATH_NGINX_ENABLED_DIR, siteName+".conf")); err != nil {
		return err
	}
	return nil
}

func EnableSite(siteName string) error {
	sourcePath := path.Join(PATH_NGINX_AVAILABLE_DIR, siteName+".conf")
	targetPath := path.Join(PATH_NGINX_ENABLED_DIR, siteName+".conf")
	if err := os.Symlink(sourcePath, targetPath); err != nil {
		return err
	}
	return nil
}
