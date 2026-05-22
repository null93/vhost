package vhost

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTempVhost(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()

	origNginx := PATH_NGINX_DIR
	origAvail := PATH_NGINX_AVAILABLE_DIR
	origEnabled := PATH_NGINX_ENABLED_DIR
	origTpl := PATH_TEMPLATES_DIR
	origCp := PATH_CHECKPOINTS_DIR

	PATH_NGINX_DIR = filepath.Join(tmpDir, "etc/nginx")
	PATH_NGINX_AVAILABLE_DIR = filepath.Join(PATH_NGINX_DIR, "sites-available")
	PATH_NGINX_ENABLED_DIR = filepath.Join(PATH_NGINX_DIR, "sites-enabled")
	PATH_TEMPLATES_DIR = filepath.Join(tmpDir, "var/lib/vhost/templates")
	PATH_CHECKPOINTS_DIR = filepath.Join(tmpDir, "var/lib/vhost/checkpoints")

	for _, d := range []string{PATH_NGINX_AVAILABLE_DIR, PATH_NGINX_ENABLED_DIR, PATH_TEMPLATES_DIR, PATH_CHECKPOINTS_DIR} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	t.Cleanup(func() {
		PATH_NGINX_DIR = origNginx
		PATH_NGINX_AVAILABLE_DIR = origAvail
		PATH_NGINX_ENABLED_DIR = origEnabled
		PATH_TEMPLATES_DIR = origTpl
		PATH_CHECKPOINTS_DIR = origCp
	})
}

func writeTestTemplate(t *testing.T, name string) Template {
	t.Helper()
	tplDir := filepath.Join(PATH_TEMPLATES_DIR, name)
	files := map[string][]byte{
		"schema.yaml": []byte(`domain:
  pattern: domain
  description: domain name
  value: example.com
`),
		"provisioner.sh":                       []byte("#!/bin/sh\nexit 0\n"),
		"sites-available/[site_name].conf.tpl": []byte("server { server_name {{ .domain }}; root /var/www/{{ .site_name }}; }\n"),
	}
	for rel, body := range files {
		full := filepath.Join(tplDir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, body, 0644); err != nil {
			t.Fatal(err)
		}
	}
	tpl, err := LoadTemplate(name)
	if err != nil {
		t.Fatal(err)
	}
	return tpl
}

// State files are addressed by filename prefix, but their content carries an
// independent SiteName field. Cross-rename rollback relies on this: after
// rename, pre-rename state files are moved to the new prefix but must still
// load with their original SiteName intact so rollback can detect the rename.
func TestStateFileRenamePreservesContent(t *testing.T) {
	setupTempVhost(t)
	tpl := writeTestTemplate(t, "blog")

	cp, err := NewCheckPoint("alpha", tpl, TemplateInput{"domain": "alpha.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	cp.Description = "initial"
	if err := cp.Save(); err != nil {
		t.Fatal(err)
	}

	if err := os.Rename(
		filepath.Join(PATH_CHECKPOINTS_DIR, "alpha_1.state"),
		filepath.Join(PATH_CHECKPOINTS_DIR, "beta_1.state"),
	); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadCheckPoint("beta_1.state")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.SiteName != "alpha" {
		t.Errorf("SiteName = %q, want %q (renaming the file must not touch content)", loaded.SiteName, "alpha")
	}
	if loaded.Input["site_name"] != "alpha" {
		t.Errorf("Input[site_name] = %q, want %q", loaded.Input["site_name"], "alpha")
	}
	if _, ok := loaded.Output["sites-available/alpha.conf"]; !ok {
		t.Errorf("Output keys = %v, want to contain sites-available/alpha.conf", loaded.Output.SortedKeys())
	}
}

// Re-rendering the latest checkpoint's template with a new site name is what
// rename writes to disk. Verify that the render substitutes the new name into
// both file paths and file bodies.
func TestRenderUsesProvidedSiteName(t *testing.T) {
	setupTempVhost(t)
	tpl := writeTestTemplate(t, "blog")

	input := TemplateInput{"domain": "shared.example.com"}
	out, err := tpl.Render("beta", input)
	if err != nil {
		t.Fatal(err)
	}
	body, ok := out["sites-available/beta.conf"]
	if !ok {
		t.Fatalf("Output keys = %v, want sites-available/beta.conf", out.SortedKeys())
	}
	if want := "/var/www/beta"; !contains(string(body), want) {
		t.Errorf("rendered body should reference %q, got: %s", want, body)
	}
	if _, ok := out["sites-available/alpha.conf"]; ok {
		t.Errorf("Output should not contain old-name path")
	}
}

// End-to-end rename simulation using the same primitives the cobra command
// uses. Asserts the post-rename state: pre-rename state files moved under the
// new prefix with content untouched, a rename checkpoint appended at
// latest+1, and on-disk files moved to the new name.
func TestRenameFlow(t *testing.T) {
	setupTempVhost(t)
	tpl := writeTestTemplate(t, "blog")

	cp1, err := NewCheckPoint("alpha", tpl, TemplateInput{"domain": "alpha.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	cp1.Description = "initial"
	if err := cp1.Output.Save(); err != nil {
		t.Fatal(err)
	}
	if err := cp1.Save(); err != nil {
		t.Fatal(err)
	}
	cp2 := cp1
	cp2.Revision = 2
	cp2.Description = "modify"
	cp2.Timestamp = time.Now()
	if err := cp2.Save(); err != nil {
		t.Fatal(err)
	}

	newInput := TemplateInput{}
	for k, v := range cp2.Input {
		newInput[k] = v
	}
	newInput["site_name"] = "beta"
	newOutput, err := cp2.Template.Render("beta", newInput)
	if err != nil {
		t.Fatal(err)
	}
	if err := cp2.Output.DeleteFiles(true); err != nil {
		t.Fatal(err)
	}
	for _, rev := range []int{1, 2} {
		if err := os.Rename(
			filepath.Join(PATH_CHECKPOINTS_DIR, fmt.Sprintf("alpha_%d.state", rev)),
			filepath.Join(PATH_CHECKPOINTS_DIR, fmt.Sprintf("beta_%d.state", rev)),
		); err != nil {
			t.Fatal(err)
		}
	}
	renameCp := CheckPoint{
		Revision:    3,
		SiteName:    "beta",
		Timestamp:   time.Now(),
		Template:    cp2.Template,
		Input:       newInput,
		Output:      newOutput,
		Description: `renamed from "alpha" to "beta"`,
	}
	if err := renameCp.Save(); err != nil {
		t.Fatal(err)
	}
	if err := renameCp.Output.Save(); err != nil {
		t.Fatal(err)
	}

	for _, rev := range []int{1, 2, 3} {
		if _, err := os.Stat(filepath.Join(PATH_CHECKPOINTS_DIR, fmt.Sprintf("beta_%d.state", rev))); err != nil {
			t.Errorf("beta_%d.state missing: %v", rev, err)
		}
	}
	for _, rev := range []int{1, 2} {
		if _, err := os.Stat(filepath.Join(PATH_CHECKPOINTS_DIR, fmt.Sprintf("alpha_%d.state", rev))); !os.IsNotExist(err) {
			t.Errorf("alpha_%d.state should be removed", rev)
		}
	}
	if _, err := os.Stat(filepath.Join(PATH_NGINX_AVAILABLE_DIR, "alpha.conf")); !os.IsNotExist(err) {
		t.Errorf("sites-available/alpha.conf should be removed")
	}
	if _, err := os.Stat(filepath.Join(PATH_NGINX_AVAILABLE_DIR, "beta.conf")); err != nil {
		t.Errorf("sites-available/beta.conf should exist: %v", err)
	}

	rev1, err := GetCheckPoint("beta", 1)
	if err != nil {
		t.Fatal(err)
	}
	if rev1.SiteName != "alpha" {
		t.Errorf("pre-rename revision lost original SiteName: got %q, want %q", rev1.SiteName, "alpha")
	}
	rev3, err := GetCheckPoint("beta", 3)
	if err != nil {
		t.Fatal(err)
	}
	if rev3.SiteName != "beta" {
		t.Errorf("rename checkpoint SiteName: got %q, want %q", rev3.SiteName, "beta")
	}

	history, err := GetCheckPoints("beta")
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 3 {
		t.Errorf("history length: got %d, want 3", len(history))
	}
	if SiteExists("alpha") {
		t.Errorf("old site %q should no longer exist", "alpha")
	}
}

// End-to-end cross-rename rollback simulation. After rename alpha→beta, a
// rollback to a pre-rename revision should move every state file back under
// the alpha prefix, restore alpha's on-disk files, and append a rollback
// revision at alpha_N.state.
func TestCrossRenameRollbackFlow(t *testing.T) {
	setupTempVhost(t)
	tpl := writeTestTemplate(t, "blog")

	// Post-rename pre-state: rev 1 has SiteName=alpha but is saved at beta_1.state,
	// rev 2 is the rename checkpoint with SiteName=beta.
	cp1, err := NewCheckPoint("alpha", tpl, TemplateInput{"domain": "alpha.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	cp1.Description = "initial"
	raw, err := cp1.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(PATH_CHECKPOINTS_DIR, "beta_1.state"), raw, 0600); err != nil {
		t.Fatal(err)
	}

	newInput := TemplateInput{}
	for k, v := range cp1.Input {
		newInput[k] = v
	}
	newInput["site_name"] = "beta"
	newOutput, err := tpl.Render("beta", newInput)
	if err != nil {
		t.Fatal(err)
	}
	renameCp := CheckPoint{
		Revision:    2,
		SiteName:    "beta",
		Template:    cp1.Template,
		Input:       newInput,
		Output:      newOutput,
		Description: `renamed from "alpha" to "beta"`,
		Timestamp:   time.Now(),
	}
	if err := renameCp.Save(); err != nil {
		t.Fatal(err)
	}
	if err := renameCp.Output.Save(); err != nil {
		t.Fatal(err)
	}

	latest, err := GetLatestCheckPoint("beta")
	if err != nil {
		t.Fatal(err)
	}
	target, err := GetCheckPoint("beta", 1)
	if err != nil {
		t.Fatal(err)
	}
	if target.SiteName == latest.SiteName {
		t.Fatalf("expected cross-rename: target=%q latest=%q", target.SiteName, latest.SiteName)
	}

	if err := latest.Output.DeleteFiles(true); err != nil {
		t.Fatal(err)
	}
	allCps, err := GetCheckPoints("beta")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range allCps {
		if err := os.Rename(
			filepath.Join(PATH_CHECKPOINTS_DIR, fmt.Sprintf("beta_%d.state", c.Revision)),
			filepath.Join(PATH_CHECKPOINTS_DIR, fmt.Sprintf("alpha_%d.state", c.Revision)),
		); err != nil {
			t.Fatal(err)
		}
	}
	target.Revision = latest.Revision + 1
	target.Description = `rollback to revision 1 (renamed from "beta" to "alpha")`
	target.Timestamp = time.Now()
	if err := target.Output.Save(); err != nil {
		t.Fatal(err)
	}
	if err := target.Save(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(PATH_NGINX_AVAILABLE_DIR, "alpha.conf")); err != nil {
		t.Errorf("alpha.conf should be restored: %v", err)
	}
	if _, err := os.Stat(filepath.Join(PATH_NGINX_AVAILABLE_DIR, "beta.conf")); !os.IsNotExist(err) {
		t.Errorf("beta.conf should be removed after rollback across rename")
	}
	for _, rev := range []int{1, 2, 3} {
		if _, err := os.Stat(filepath.Join(PATH_CHECKPOINTS_DIR, fmt.Sprintf("alpha_%d.state", rev))); err != nil {
			t.Errorf("alpha_%d.state missing: %v", rev, err)
		}
		if _, err := os.Stat(filepath.Join(PATH_CHECKPOINTS_DIR, fmt.Sprintf("beta_%d.state", rev))); !os.IsNotExist(err) {
			t.Errorf("beta_%d.state should be removed", rev)
		}
	}
	after, err := GetLatestCheckPoint("alpha")
	if err != nil {
		t.Fatal(err)
	}
	if after.Revision != 3 || after.SiteName != "alpha" {
		t.Errorf("post-rollback latest: revision=%d SiteName=%q, want 3/alpha", after.Revision, after.SiteName)
	}
}

// DeleteFiles should clean up parent directories that become empty as a result
// of the deletion. Directories with unmanaged files left in them must be kept.
func TestDeleteFilesRemovesEmptyParents(t *testing.T) {
	setupTempVhost(t)
	tpl := writeTestTemplate(t, "blog")

	cp, err := NewCheckPoint("alpha", tpl, TemplateInput{"domain": "alpha.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	cp.Output["conf.d/alpha/ssl-options.conf"] = []byte("# ssl options for alpha\n")
	if err := cp.Output.Save(); err != nil {
		t.Fatal(err)
	}

	// An unmanaged sibling that vhost shouldn't know about — survives the delete.
	unmanaged := filepath.Join(PATH_NGINX_DIR, "conf.d/alpha/magento.conf")
	if err := os.WriteFile(unmanaged, []byte("# provisioner-created\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := cp.Output.DeleteFiles(true); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(PATH_NGINX_DIR, "conf.d/alpha/ssl-options.conf")); !os.IsNotExist(err) {
		t.Errorf("managed file should be removed")
	}
	if _, err := os.Stat(unmanaged); err != nil {
		t.Errorf("unmanaged file should remain: %v", err)
	}
	if _, err := os.Stat(filepath.Join(PATH_NGINX_DIR, "conf.d/alpha")); err != nil {
		t.Errorf("conf.d/alpha still has unmanaged content — directory must be kept: %v", err)
	}

	// Now remove the unmanaged file and re-run delete on a separate output that
	// only references conf.d/alpha — the directory should disappear this time.
	if err := os.Remove(unmanaged); err != nil {
		t.Fatal(err)
	}
	leftover := TemplateOutput{
		"conf.d/alpha/extra.conf": []byte("# extra\n"),
	}
	if err := leftover.Save(); err != nil {
		t.Fatal(err)
	}
	if err := leftover.DeleteFiles(true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(PATH_NGINX_DIR, "conf.d/alpha")); !os.IsNotExist(err) {
		t.Errorf("empty parent conf.d/alpha should be removed")
	}
	// PATH_NGINX_DIR is the stop boundary — the walk must never remove it.
	if _, err := os.Stat(PATH_NGINX_DIR); err != nil {
		t.Errorf("PATH_NGINX_DIR should never be removed: %v", err)
	}
}

// NewCheckPointWithoutProvisioner skips validation for provisioner_only
// inputs so users don't have to supply provisioner-only KEY=VALUE pairs that
// have no effect on the rendered files (those inputs are only consumed by
// provisioner.sh, which won't run on this path).
func TestNewCheckPointWithoutProvisionerSkipsProvisionerOnlyValidation(t *testing.T) {
	setupTempVhost(t)
	tplDir := filepath.Join(PATH_TEMPLATES_DIR, "magentolike")
	files := map[string][]byte{
		"schema.yaml": []byte(`domain:
  pattern: domain
  description: domain name
  value: example.com
magento_version:
  pattern: version
  description: version (provisioner-only, no default)
  provisioner_only: true
`),
		"provisioner.sh":                       []byte("#!/bin/sh\nexit 0\n"),
		"sites-available/[site_name].conf.tpl": []byte("server { server_name {{ .domain }}; }\n"),
	}
	for rel, body := range files {
		full := filepath.Join(tplDir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, body, 0644); err != nil {
			t.Fatal(err)
		}
	}
	tpl, err := LoadTemplate("magentolike")
	if err != nil {
		t.Fatal(err)
	}

	// Sanity: the normal path must reject because magento_version has no value
	// and its default is empty (doesn't match the version pattern).
	if _, err := NewCheckPoint("alpha", tpl, TemplateInput{"domain": "alpha.example.com"}); err == nil {
		t.Fatal("NewCheckPoint should fail when provisioner_only inputs lack a valid default")
	}

	cp, err := NewCheckPointWithoutProvisioner("alpha", tpl, TemplateInput{"domain": "alpha.example.com"})
	if err != nil {
		t.Fatalf("NewCheckPointWithoutProvisioner unexpected error: %v", err)
	}
	if _, ok := cp.Input["magento_version"]; ok {
		t.Errorf("provisioner_only input should not be stored on checkpoint, got Input=%v", cp.Input)
	}
	if _, ok := cp.Output["sites-available/alpha.conf"]; !ok {
		t.Errorf("expected rendered output to contain site conf, got keys=%v", cp.Output.SortedKeys())
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
