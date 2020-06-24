package config

import "testing"

func TestParseGitURL(t *testing.T) {
	tests := []struct {
		input                  string
		outputHost, outputPath string
	}{
		// yes technically this is a gitlab-cli, but we can still parse URLs with github :-)
		{"git@github.com:tommyknows/gitlab-cli.git", "github.com", "/tommyknows/gitlab-cli"},
		{"git@gitlab.company.com:our/super/project.git", "gitlab.company.com", "/our/super/project"},
		// gitea uses these as SSH URL, so let's check that we can parse those too. maybe
		// some gitlab configurations or versions use the same format.
		{"ssh://git@gitea.com/user/project.git", "gitea.com", "/user/project"},
		{"https://gitlab.com/gitlab-org/gitlab.git", "gitlab.com", "/gitlab-org/gitlab"},
	}

	for _, tt := range tests {
		u, err := parseGitURL(tt.input)
		if err != nil {
			t.Errorf("error parsing git URL: %v", err)
		}

		if u.Host != tt.outputHost || u.Path != tt.outputPath {
			t.Errorf("parsed URL not correct. expected=%v%v, got=%v%v",
				tt.outputHost, tt.outputPath,
				u.Host, u.Path)
		}
	}
}
