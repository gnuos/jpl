package pm

import (
	"testing"
)

// ============================================================================
// Version 解析测试
// ============================================================================

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
		major   uint64
		minor   uint64
		patch   uint64
	}{
		{"1.2.3", false, 1, 2, 3},
		{"v1.2.3", false, 1, 2, 3},
		{"0.1.0", false, 0, 1, 0},
		{"10.20.30", false, 10, 20, 30},
		{"1.2.3-alpha", false, 1, 2, 3},
		{"1.2.3+build", false, 1, 2, 3},
		{"1.2.3-alpha.1+build.123", false, 1, 2, 3},
		{"invalid", true, 0, 0, 0},
		{"", true, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v, err := ParseVersion(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseVersion(%q) should return error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseVersion(%q) failed: %v", tt.input, err)
			}
			if v.Major() != tt.major || v.Minor() != tt.minor || v.Patch() != tt.patch {
				t.Errorf("ParseVersion(%q) = %d.%d.%d, want %d.%d.%d",
					tt.input, v.Major(), v.Minor(), v.Patch(), tt.major, tt.minor, tt.patch)
			}
		})
	}
}

// ============================================================================
// Version 比较测试
// ============================================================================

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.2.3", "1.2.3", 0},
		{"1.2.3", "1.2.4", -1},
		{"1.2.3", "1.3.0", -1},
		{"1.2.3", "2.0.0", -1},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0", "1.0.0-alpha", 1},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			a := MustParseVersion(tt.a)
			b := MustParseVersion(tt.b)
			got := CompareVersions(a, b)
			if got != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// ============================================================================
// Constraint 测试
// ============================================================================

func TestParseConstraint(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"^1.2.3", false},
		{"~1.2.3", false},
		{">=1.2.3", false},
		{">1.2.3", false},
		{"<1.2.3", false},
		{"<=1.2.3", false},
		{"=1.2.3", false},
		{"1.2.3", false},
		{"*", false},
		{"", false},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParseConstraint(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseConstraint(%q) should return error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseConstraint(%q) failed: %v", tt.input, err)
			}
		})
	}
}

// ============================================================================
// VersionSatisfies 测试
// ============================================================================

func TestVersionSatisfies(t *testing.T) {
	tests := []struct {
		version    string
		constraint string
		want       bool
	}{
		// ^ (compatible with)
		{"1.2.3", "^1.2.3", true},
		{"1.2.4", "^1.2.3", true},
		{"1.3.0", "^1.2.3", true},
		{"2.0.0", "^1.2.3", false},
		{"1.2.2", "^1.2.3", false},

		// ~ (patch level)
		{"1.2.3", "~1.2.3", true},
		{"1.2.4", "~1.2.3", true},
		{"1.2.99", "~1.2.3", true},
		{"1.3.0", "~1.2.3", false},
		{"1.2.2", "~1.2.3", false},

		// >= (greater than or equal)
		{"1.2.3", ">=1.2.3", true},
		{"1.2.4", ">=1.2.3", true},
		{"2.0.0", ">=1.2.3", true},
		{"1.2.2", ">=1.2.3", false},

		// > (greater than)
		{"1.2.4", ">1.2.3", true},
		{"2.0.0", ">1.2.3", true},
		{"1.2.3", ">1.2.3", false},
		{"1.2.2", ">1.2.3", false},

		// < (less than)
		{"1.2.2", "<1.2.3", true},
		{"1.0.0", "<1.2.3", true},
		{"1.2.3", "<1.2.3", false},
		{"1.2.4", "<1.2.3", false},

		// <= (less than or equal)
		{"1.2.3", "<=1.2.3", true},
		{"1.2.2", "<=1.2.3", true},
		{"1.2.4", "<=1.2.3", false},

		// = (exact match)
		{"1.2.3", "=1.2.3", true},
		{"1.2.4", "=1.2.3", false},

		// * (any)
		{"1.0.0", "*", true},
		{"999.999.999", "*", true},

		// 空约束 (any)
		{"1.0.0", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.version+"_"+tt.constraint, func(t *testing.T) {
			v := MustParseVersion(tt.version)
			c, err := ParseConstraint(tt.constraint)
			if err != nil {
				t.Fatalf("ParseConstraint(%q) failed: %v", tt.constraint, err)
			}
			got := VersionSatisfies(v, c)
			if got != tt.want {
				t.Errorf("VersionSatisfies(%q, %q) = %v, want %v", tt.version, tt.constraint, got, tt.want)
			}
		})
	}
}

// ============================================================================
// ParseSourceWithConstraint 测试
// ============================================================================

func TestParseSourceWithConstraint(t *testing.T) {
	tests := []struct {
		source     string
		wantURL    string
		wantConst  string
		wantTag    string
		wantBranch string
	}{
		{
			source:    "https://github.com/user/repo.git@^1.2.0",
			wantURL:   "https://github.com/user/repo.git",
			wantConst: "^1.2.0",
		},
		{
			source:    "https://github.com/user/repo.git@~2.0.0",
			wantURL:   "https://github.com/user/repo.git",
			wantConst: "~2.0.0",
		},
		{
			source:    "https://github.com/user/repo.git@>=1.0.0",
			wantURL:   "https://github.com/user/repo.git",
			wantConst: ">=1.0.0",
		},
		{
			source:  "https://github.com/user/repo.git@v1.0.0",
			wantURL: "https://github.com/user/repo.git",
			wantTag: "v1.0.0",
		},
		{
			source:     "https://github.com/user/repo.git#main",
			wantURL:    "https://github.com/user/repo.git",
			wantBranch: "main",
		},
		{
			source:  "https://github.com/user/repo.git",
			wantURL: "https://github.com/user/repo.git",
		},
		{
			source:  "../my-lib",
			wantURL: "../my-lib",
		},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			info, err := ParseSourceWithConstraint(tt.source)
			if err != nil {
				t.Fatalf("ParseSourceWithConstraint(%q) failed: %v", tt.source, err)
			}
			if info.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", info.URL, tt.wantURL)
			}
			if info.Constraint != tt.wantConst {
				t.Errorf("Constraint = %q, want %q", info.Constraint, tt.wantConst)
			}
			if info.Tag != tt.wantTag {
				t.Errorf("Tag = %q, want %q", info.Tag, tt.wantTag)
			}
			if info.Branch != tt.wantBranch {
				t.Errorf("Branch = %q, want %q", info.Branch, tt.wantBranch)
			}
		})
	}
}

// ============================================================================
// SelectBestVersion 测试
// ============================================================================

func TestSelectBestVersion(t *testing.T) {
	versions := []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0", "2.1.0"}

	tests := []struct {
		constraint string
		want       string
		wantErr    bool
	}{
		{"^1.0.0", "v1.2.0", false},
		{"~1.0.0", "v1.0.0", false},
		{">=2.0.0", "v2.1.0", false},
		{"^2.0.0", "v2.1.0", false},
		{"^3.0.0", "", true}, // 无匹配
		{"", "2.1.0", false}, // 无约束，返回列表中最后一个
	}

	for _, tt := range tests {
		t.Run(tt.constraint, func(t *testing.T) {
			got, err := SelectBestVersion(versions, tt.constraint)
			if tt.wantErr {
				if err == nil {
					t.Errorf("SelectBestVersion(%q) should return error", tt.constraint)
				}
				return
			}
			if err != nil {
				t.Fatalf("SelectBestVersion(%q) failed: %v", tt.constraint, err)
			}
			if got != tt.want {
				t.Errorf("SelectBestVersion(%q) = %q, want %q", tt.constraint, got, tt.want)
			}
		})
	}
}

// ============================================================================
// isVersionConstraint 测试
// ============================================================================

func TestIsVersionConstraint(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"^1.2.3", true},
		{"~1.2.3", true},
		{">=1.2.3", true},
		{">1.2.3", true},
		{"<1.2.3", true},
		{"<=1.2.3", true},
		{"=1.2.3", true},
		{"1.x", true},
		{"1.2.x", true},
		{"*", true},
		{"1.0.0 - 2.0.0", true},
		{"v1.0.0", false},
		{"1.0.0", false},
		{"main", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isVersionConstraint(tt.input)
			if got != tt.want {
				t.Errorf("isVersionConstraint(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
