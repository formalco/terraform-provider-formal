package datasources

import (
	"strings"
	"testing"
)

func TestRegoPolicyDocument_BasicPredicate(t *testing.T) {
	g := &generator{
		constants:  make(map[string]string),
		predicates: make(map[string][]predicate),
	}

	g.predicates["is_admin"] = []predicate{{
		name: "is_admin",
		conditions: []condition{{
			test:     "any_in",
			variable: "user.groups",
			values:   []string{"admin"},
		}},
	}}

	rego, err := g.generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(rego, "package formal.v2") {
		t.Error("missing package declaration")
	}
	if !strings.Contains(rego, "import future.keywords.if") {
		t.Error("missing if import")
	}
	if !strings.Contains(rego, "import future.keywords.in") {
		t.Error("missing in import")
	}
	if !strings.Contains(rego, "is_admin if {") {
		t.Error("missing predicate definition")
	}
	if !strings.Contains(rego, `some _x in input.user.groups; _x in ["admin"]`) {
		t.Errorf("incorrect condition generation, got:\n%s", rego)
	}
}

func TestRegoPolicyDocument_MultipleConditions(t *testing.T) {
	gen := &generator{
		constants:  make(map[string]string),
		predicates: make(map[string][]predicate),
	}

	gen.predicates["is_human_admin"] = []predicate{{
		name: "is_human_admin",
		conditions: []condition{
			{test: "equals", variable: "user.type", values: []string{"human"}},
			{test: "any_in", variable: "user.groups", values: []string{"admin"}},
		},
	}}

	rego, err := gen.generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(rego, `input.user.type == "human"`) {
		t.Errorf("missing first condition, got:\n%s", rego)
	}
	if !strings.Contains(rego, `some _x in input.user.groups; _x in ["admin"]`) {
		t.Errorf("missing second condition, got:\n%s", rego)
	}
}

func TestRegoPolicyDocument_Constants(t *testing.T) {
	gen := &generator{
		constants:  make(map[string]string),
		predicates: make(map[string][]predicate),
	}

	gen.constants["allowed_groups"] = `["admin", "user"]`
	gen.constants["rate_limits"] = `{"admin": -1, "user": 100}`

	rego, err := gen.generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(rego, `allowed_groups := ["admin", "user"]`) {
		t.Errorf("missing array constant, got:\n%s", rego)
	}
	if !strings.Contains(rego, `rate_limits := {"admin": -1, "user": 100}`) {
		t.Errorf("missing map constant, got:\n%s", rego)
	}
}

func TestRegoPolicyDocument_BasicRule(t *testing.T) {
	gen := &generator{
		constants:  make(map[string]string),
		predicates: make(map[string][]predicate),
	}

	gen.predicates["should_block"] = []predicate{{
		name: "should_block",
		conditions: []condition{{
			test:     "equals",
			variable: "resource.env",
			values:   []string{"production"},
		}},
	}}

	gen.rules = []rule{{
		name: "pre_request",
		effect: effect{
			action:     "block",
			effectType: "block_with_formal_message",
		},
		whenAllOf: []string{"should_block"},
	}}

	rego, err := gen.generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(rego, `pre_request := { "action": "block", "type": "block_with_formal_message" } if {`) {
		t.Errorf("incorrect rule generation, got:\n%s", rego)
	}
	if !strings.Contains(rego, "    should_block\n") {
		t.Errorf("missing predicate reference in rule, got:\n%s", rego)
	}
}

func TestRegoPolicyDocument_RuleWithNegation(t *testing.T) {
	gen := &generator{
		constants:  make(map[string]string),
		predicates: make(map[string][]predicate),
	}

	gen.predicates["is_admin"] = []predicate{{
		name: "is_admin",
		conditions: []condition{{
			test:     "equals",
			variable: "user.role",
			values:   []string{"admin"},
		}},
	}}

	gen.rules = []rule{{
		name: "pre_request",
		effect: effect{
			action:  "block",
			message: "Admin only",
		},
		whenAllOf: []string{"!is_admin"},
	}}

	rego, err := gen.generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(rego, "not is_admin") {
		t.Errorf("missing negated predicate reference, got:\n%s", rego)
	}
}

func TestRegoPolicyDocument_DefaultRule(t *testing.T) {
	gen := &generator{
		constants:  make(map[string]string),
		predicates: make(map[string][]predicate),
	}

	gen.rules = []rule{{
		name:      "session",
		isDefault: true,
		effect: effect{
			action:     "block",
			effectType: "block_with_formal_message",
		},
	}}

	rego, err := gen.generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(rego, `default session := { "action": "block", "type": "block_with_formal_message" }`) {
		t.Errorf("incorrect default rule generation, got:\n%s", rego)
	}
}

func TestRegoPolicyDocument_MaskRule(t *testing.T) {
	gen := &generator{
		constants:  make(map[string]string),
		predicates: make(map[string][]predicate),
	}

	gen.rules = []rule{{
		name: "post_request",
		effect: effect{
			action:     "mask",
			effectType: "redact.full",
			typesafe:   "fallback_to_default",
			dataLabel:  "email_address",
		},
	}}

	rego, err := gen.generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(rego, `"action": "mask"`) {
		t.Error("missing mask action")
	}
	if !strings.Contains(rego, `"columns": columns`) {
		t.Error("missing columns reference in effect")
	}
	if !strings.Contains(rego, `columns := [col | col := input.columns[_]; col["data_label"] == "email_address"]`) {
		t.Errorf("incorrect column filtering, got:\n%s", rego)
	}
}

func TestRegoPolicyDocument_AllColumnsTarget(t *testing.T) {
	gen := &generator{
		constants:  make(map[string]string),
		predicates: make(map[string][]predicate),
	}

	gen.rules = []rule{{
		name: "post_request",
		effect: effect{
			action:     "mask",
			effectType: "hash.with_salt",
			allColumns: true,
		},
	}}

	rego, err := gen.generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(rego, `"columns": input.columns`) {
		t.Errorf("incorrect all_columns reference, got:\n%s", rego)
	}
}

func TestRegoPolicyDocument_ColumnNameTarget(t *testing.T) {
	gen := &generator{
		constants:  make(map[string]string),
		predicates: make(map[string][]predicate),
	}

	gen.rules = []rule{{
		name: "post_request",
		effect: effect{
			action:     "decrypt",
			columnName: "ssn",
		},
	}}

	rego, err := gen.generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(rego, `col["name"] == "ssn"`) {
		t.Errorf("column_name target incorrect: %s", rego)
	}
}

func TestRegoPolicyDocument_MultipleRulesOR(t *testing.T) {
	gen := &generator{
		constants:  make(map[string]string),
		predicates: make(map[string][]predicate),
	}

	gen.predicates["is_admin"] = []predicate{{
		name: "is_admin",
		conditions: []condition{{
			test:     "equals",
			variable: "user.role",
			values:   []string{"admin"},
		}},
	}}

	gen.predicates["is_owner"] = []predicate{{
		name: "is_owner",
		conditions: []condition{{
			test:     "equals",
			variable: "user.role",
			values:   []string{"owner"},
		}},
	}}

	// OR via multiple rules with same name
	gen.rules = []rule{
		{name: "session", effect: effect{action: "allow"}, whenAllOf: []string{"is_admin"}, index: 0},
		{name: "session", effect: effect{action: "allow"}, whenAllOf: []string{"is_owner"}, index: 1},
	}

	rego, err := gen.generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count := strings.Count(rego, `session := { "action": "allow" } if {`)
	if count != 2 {
		t.Errorf("expected 2 session rules for OR, got %d:\n%s", count, rego)
	}
}

func TestRegoPolicyDocument_MultiplePredicateBlocksOR(t *testing.T) {
	gen := &generator{
		constants:  make(map[string]string),
		predicates: make(map[string][]predicate),
	}

	// OR via multiple predicate blocks with same name
	gen.predicates["is_privileged"] = []predicate{
		{name: "is_privileged", conditions: []condition{{test: "equals", variable: "user.role", values: []string{"admin"}}}},
		{name: "is_privileged", conditions: []condition{{test: "equals", variable: "user.role", values: []string{"owner"}}}},
	}

	rego, err := gen.generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count := strings.Count(rego, "is_privileged if {")
	if count != 2 {
		t.Errorf("expected 2 is_privileged rules for OR, got %d:\n%s", count, rego)
	}
}

func TestRegoPolicyDocument_UndefinedPredicateInRule(t *testing.T) {
	gen := &generator{
		constants:  make(map[string]string),
		predicates: make(map[string][]predicate),
		errors:     []string{},
	}

	gen.rules = []rule{{
		name:      "session",
		whenAllOf: []string{"nonexistent_predicate"},
		effect:    effect{action: "block"},
	}}

	gen.validateReferences()

	if len(gen.errors) == 0 {
		t.Error("expected undefined predicate error in rule")
	}
	if !strings.Contains(gen.errors[0], "undefined predicate") {
		t.Errorf("expected undefined predicate error, got: %s", gen.errors[0])
	}
}

func TestRegoPolicyDocument_UndefinedConstant(t *testing.T) {
	gen := &generator{
		constants:  make(map[string]string),
		predicates: make(map[string][]predicate),
		errors:     []string{},
	}

	gen.predicates["test_pred"] = []predicate{{
		name: "test_pred",
		conditions: []condition{{
			test:        "any_in",
			variable:    "user.groups",
			constantRef: "nonexistent_constant",
		}},
	}}

	gen.validateReferences()

	if len(gen.errors) == 0 {
		t.Error("expected undefined constant error")
	}
}

func TestRegoPolicyDocument_ReservedName(t *testing.T) {
	reserved := []string{"input", "data", "true", "false", "if", "in", "not", "count"}

	for _, name := range reserved {
		err := validateName(name)
		if err == nil {
			t.Errorf("expected error for reserved name %q", name)
		}
	}
}

func TestRegoPolicyDocument_ValidIdentifier(t *testing.T) {
	valid := []string{"is_admin", "hasGroup", "check_123", "_private"}
	invalid := []string{"123start", "has-dash", "has space", "has.dot"}

	for _, name := range valid {
		err := validateName(name)
		if err != nil {
			t.Errorf("expected %q to be valid, got error: %v", name, err)
		}
	}

	for _, name := range invalid {
		err := validateName(name)
		if err == nil {
			t.Errorf("expected %q to be invalid", name)
		}
	}
}

func TestRegoPolicyDocument_ConditionOperators(t *testing.T) {
	tests := []struct {
		test     string
		variable string
		values   []string
		expected string
	}{
		{"equals", "user.name", []string{"alice"}, `input.user.name == "alice"`},
		{"not_equals", "user.name", []string{"bob"}, `input.user.name != "bob"`},
		{"in", "user.role", []string{"admin", "user"}, `input.user.role in ["admin", "user"]`},
		{"not_in", "user.role", []string{"guest"}, `not input.user.role in ["guest"]`},
		{"any_in", "user.groups", []string{"admin"}, `some _x in input.user.groups; _x in ["admin"]`},
		{"all_in", "user.groups", []string{"a", "b"}, `every _x in input.user.groups { _x in ["a", "b"] }`},
		{"none_in", "user.groups", []string{"banned"}, `not (some _x in input.user.groups; _x in ["banned"])`},
		{"greater_than", "user.age", []string{"18"}, `input.user.age > 18`},
		{"less_than", "user.score", []string{"100"}, `input.user.score < 100`},
		{"contains", "user.email", []string{"@company.com"}, `contains(input.user.email, "@company.com")`},
		{"not_contains", "user.email", []string{"spam"}, `not contains(input.user.email, "spam")`},
		{"starts_with", "resource.name", []string{"prod-"}, `startswith(input.resource.name, "prod-")`},
		{"ends_with", "resource.name", []string{"-db"}, `endswith(input.resource.name, "-db")`},
		{"regex", "user.email", []string{".*@company.com"}, `regex.match(".*@company.com", input.user.email)`},
		{"exists", "user.verified", nil, `input.user.verified`},
		{"not_exists", "user.banned", nil, `not input.user.banned`},
	}

	gen := &generator{}

	for _, tc := range tests {
		cond := condition{test: tc.test, variable: tc.variable, values: tc.values}
		result := gen.generateCondition(cond)
		if result != tc.expected {
			t.Errorf("test=%q: expected %q, got %q", tc.test, tc.expected, result)
		}
	}
}

func TestRegoPolicyDocument_ConditionValidation(t *testing.T) {
	// Missing variable
	cond := condition{test: "equals", values: []string{"foo"}}
	errs := cond.validate("test")
	if len(errs) == 0 {
		t.Error("expected error for missing variable")
	}

	// Missing values for non-exists operator
	cond = condition{test: "equals", variable: "user.name"}
	errs = cond.validate("test")
	if len(errs) == 0 {
		t.Error("expected error for missing values")
	}

	// exists doesn't need values
	cond = condition{test: "exists", variable: "user.name"}
	errs = cond.validate("test")
	if len(errs) != 0 {
		t.Errorf("exists should not require values, got errors: %v", errs)
	}

	// Valid condition with constant reference
	cond = condition{test: "any_in", variable: "user.groups", constantRef: "allowed"}
	errs = cond.validate("test")
	if len(errs) != 0 {
		t.Errorf("valid condition should have no errors, got: %v", errs)
	}
}

func TestRegoPolicyDocument_IncludedConnectors(t *testing.T) {
	gen := &generator{
		constants:          make(map[string]string),
		predicates:         make(map[string][]predicate),
		includedConnectors: []string{"db-proxy", "api-gateway"},
	}

	rego, err := gen.generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(rego, `included_connectors := ["db-proxy","api-gateway"]`) {
		t.Errorf("incorrect included_connectors, got:\n%s", rego)
	}
}

func TestRegoPolicyDocument_Description(t *testing.T) {
	gen := &generator{
		constants:   make(map[string]string),
		predicates:  make(map[string][]predicate),
		description: "Access control policy for production database",
	}

	rego, err := gen.generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(rego, "# Access control policy for production database") {
		t.Errorf("missing description comment, got:\n%s", rego)
	}
}

func TestRegoPolicyDocument_RawRego(t *testing.T) {
	gen := &generator{
		constants:  make(map[string]string),
		predicates: make(map[string][]predicate),
		rawRego: `# Custom helper function
get_table_name(path) := name if {
    parts := split(".", path)
    name := parts[count(parts) - 1]
}`,
	}

	rego, err := gen.generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(rego, "# Custom helper function") {
		t.Errorf("missing raw rego, got:\n%s", rego)
	}
}

func TestRegoPolicyDocument_ConstantReference(t *testing.T) {
	gen := &generator{
		constants:  make(map[string]string),
		predicates: make(map[string][]predicate),
	}

	gen.constants["allowed_groups"] = `["admin", "user"]`
	gen.predicates["has_group"] = []predicate{{
		name: "has_group",
		conditions: []condition{{
			test:        "any_in",
			variable:    "user.groups",
			constantRef: "allowed_groups",
		}},
	}}

	rego, err := gen.generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(rego, "_x in allowed_groups") {
		t.Errorf("should reference constant by name, got:\n%s", rego)
	}
}

func TestRegoPolicyDocument_FullIntegration(t *testing.T) {
	gen := &generator{
		description:        "Test policy",
		includedConnectors: []string{"test-connector"},
		constants:          map[string]string{"allowed_groups": `["admin", "analyst"]`},
		predicates: map[string][]predicate{
			"is_authorized": {{
				name: "is_authorized",
				conditions: []condition{{
					test:        "any_in",
					variable:    "user.groups",
					constantRef: "allowed_groups",
				}},
			}},
		},
		rules: []rule{
			{
				name:      "session",
				isDefault: true,
				effect:    effect{action: "block", effectType: "block_with_formal_message"},
				index:     0,
			},
			{
				name:      "session",
				isDefault: false,
				effect:    effect{action: "allow", reason: "User is authorized"},
				whenAllOf: []string{"is_authorized"},
				index:     1,
			},
		},
	}

	rego, err := gen.generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedParts := []string{
		"package formal.v2",
		"import future.keywords.if",
		"import future.keywords.in",
		"# Test policy",
		`included_connectors := ["test-connector"]`,
		`allowed_groups := ["admin", "analyst"]`,
		"is_authorized if {",
		"some _x in input.user.groups; _x in allowed_groups",
		`default session := { "action": "block", "type": "block_with_formal_message" }`,
		`session := { "action": "allow", "reason": "User is authorized" } if {`,
		"is_authorized",
	}

	for _, part := range expectedParts {
		if !strings.Contains(rego, part) {
			t.Errorf("missing expected part %q in:\n%s", part, rego)
		}
	}
}
