package deep

import (
	"strings"
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/style"
)

func TestEnhanceStyleContextSkipsSceneSensitivePreferencesAcrossDomains(t *testing.T) {
	profile := &style.EnhancedProfile{
		UserProfile: style.UserProfile{
			LanguageTone:      "专业正式",
			PreferredColors:   []string{"#111111"},
			PreferredThemes:   []string{"charcoal_light"},
			LayoutPreferences: []string{"two-column"},
			SpecialNotes:      []string{"用户偏好深色背景"},
			TypicalPageCount:  18,
			UserFacts: style.UserFacts{
				DisplayName:  "李明",
				Organization: "蓝鲸智云",
				Department:   "产品研发部",
				JobTitle:     "解决方案架构师",
			},
		},
		DomainPreferences: map[string]int{"business": 5},
		SuccessPatterns: []style.SuccessPattern{{
			Domain:       "business",
			Template:     "pitch-deck",
			Theme:        "charcoal_light",
			SuccessCount: 5,
		}},
	}

	ctx := enhanceStyleContextWithProfile("当前推荐配色: ocean_soft", profile, "academic")

	mustContain(t, ctx, "【用户确定性资料】")
	mustContain(t, ctx, "姓名/称呼: 李明")
	mustContain(t, ctx, "工作单位/组织: 蓝鲸智云")
	mustContain(t, ctx, "部门/团队: 产品研发部")
	mustContain(t, ctx, "职位/身份: 解决方案架构师")
	mustContain(t, ctx, "语言风格参考: 专业正式")
	mustContain(t, ctx, "已跳过历史模板/配色/布局/备注等场景敏感偏好")
	mustNotContain(t, ctx, "#111111")
	mustNotContain(t, ctx, "two-column")
	mustNotContain(t, ctx, "pitch-deck")
	mustNotContain(t, ctx, "用户偏好深色背景")
}

func TestEnhanceStyleContextIncludesSameDomainPreferences(t *testing.T) {
	profile := &style.EnhancedProfile{
		UserProfile: style.UserProfile{
			PreferredColors:   []string{"#111111"},
			LayoutPreferences: []string{"two-column"},
			SpecialNotes:      []string{"用户偏好深色背景"},
		},
		DomainPreferences: map[string]int{"business": 2},
		SuccessPatterns: []style.SuccessPattern{{
			Domain:       "business",
			Template:     "pitch-deck",
			Theme:        "charcoal_light",
			SuccessCount: 2,
		}},
	}

	ctx := enhanceStyleContextWithProfile("", profile, "business")

	mustContain(t, ctx, "同领域历史可参考项")
	mustContain(t, ctx, "#111111")
	mustContain(t, ctx, "two-column")
	mustContain(t, ctx, "用户偏好深色背景")
	mustContain(t, ctx, "pitch-deck")
}

func mustContain(t *testing.T, s, want string) {
	t.Helper()
	if !strings.Contains(s, want) {
		t.Fatalf("expected context to contain %q, got:\n%s", want, s)
	}
}

func mustNotContain(t *testing.T, s, want string) {
	t.Helper()
	if strings.Contains(s, want) {
		t.Fatalf("expected context not to contain %q, got:\n%s", want, s)
	}
}
