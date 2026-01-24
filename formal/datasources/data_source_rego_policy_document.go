package datasources

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func RegoPolicyDocument() *schema.Resource {
	return &schema.Resource{
		Description: "Generates Rego policy code from declarative predicates and rules.",
		ReadContext: regoPolicyDocumentRead,
		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description comment at the top of the generated Rego.",
			},
			"included_connectors": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of connector names this policy applies to.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"constant": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Named constants for use in predicates. Use jsonencode() for the value.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"predicate": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Named boolean predicates. Multiple blocks with the same name create OR branches.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"condition": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     conditionSchema(),
						},
					},
				},
			},
			"rule": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"default": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"effect": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem:     effectSchema(),
						},
						"when": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"all_of": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "Predicates that must all be true. Use ! prefix for NOT.",
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"raw_rego": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Raw Rego code to include for custom functions.",
			},
			"rego": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func conditionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"test": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"equals", "not_equals",
					"in", "not_in", "any_in", "all_in", "none_in",
					"contains", "not_contains",
					"starts_with", "ends_with",
					"regex",
					"greater_than", "less_than",
					"greater_than_or_equal", "less_than_or_equal",
					"exists", "not_exists",
				}, false),
			},
			"variable": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"values": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"constant": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Reference a constant instead of inline values.",
			},
		},
	}
}

func effectSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"action": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"allow", "block", "mask", "decrypt"}, false),
			},
			"type":        {Type: schema.TypeString, Optional: true},
			"sub_type":    {Type: schema.TypeString, Optional: true},
			"typesafe":    {Type: schema.TypeString, Optional: true},
			"message":     {Type: schema.TypeString, Optional: true},
			"reason":      {Type: schema.TypeString, Optional: true},
			"data_label":  {Type: schema.TypeString, Optional: true},
			"column_name": {Type: schema.TypeString, Optional: true},
			"all_columns": {Type: schema.TypeBool, Optional: true, Default: false},
		},
	}
}

var reservedNames = map[string]bool{
	"input": true, "data": true, "true": true, "false": true, "null": true,
	"if": true, "in": true, "some": true, "every": true, "not": true,
	"with": true, "as": true, "default": true, "else": true,
	"package": true, "import": true,
	"count": true, "sum": true, "max": true, "min": true,
	"contains": true, "startswith": true, "endswith": true,
}

var inOperators = map[string]bool{
	"in": true, "not_in": true, "any_in": true, "all_in": true, "none_in": true,
}

var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func validateName(name string) error {
	if !validIdentifier.MatchString(name) {
		return fmt.Errorf("%q is not a valid identifier", name)
	}
	if reservedNames[name] {
		return fmt.Errorf("%q is a reserved name", name)
	}
	return nil
}

type condition struct {
	test        string
	variable    string
	values      []string
	constantRef string
}

type predicate struct {
	name       string
	conditions []condition
}

type effect struct {
	action     string
	effectType string
	subType    string
	typesafe   string
	message    string
	reason     string
	dataLabel  string
	columnName string
	allColumns bool
}

type rule struct {
	name      string
	isDefault bool
	effect    effect
	whenAllOf []string
	index     int
}

type generator struct {
	description        string
	includedConnectors []string
	constants          map[string]string
	predicates         map[string][]predicate
	rules              []rule
	rawRego            string
	errors             []string
}

func regoPolicyDocumentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	gen := &generator{
		constants:  make(map[string]string),
		predicates: make(map[string][]predicate),
	}

	gen.parseDescription(d)
	gen.parseIncludedConnectors(d)
	gen.parseConstants(d)
	gen.parsePredicates(d)
	gen.parseRules(d)
	gen.parseRawRego(d)
	gen.validateReferences()

	if len(gen.errors) > 0 {
		var diags diag.Diagnostics
		for _, err := range gen.errors {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Validation error",
				Detail:   err,
			})
		}
		return diags
	}

	rego, err := gen.generate()
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("rego", rego)
	hash := sha256.Sum256([]byte(rego))
	d.SetId(hex.EncodeToString(hash[:]))

	return nil
}

func (gen *generator) parseDescription(d *schema.ResourceData) {
	if v, ok := d.GetOk("description"); ok {
		gen.description = v.(string)
	}
}

func (gen *generator) parseIncludedConnectors(d *schema.ResourceData) {
	if v, ok := d.GetOk("included_connectors"); ok {
		for _, item := range v.([]interface{}) {
			gen.includedConnectors = append(gen.includedConnectors, item.(string))
		}
	}
}

func (gen *generator) parseConstants(d *schema.ResourceData) {
	v, ok := d.GetOk("constant")
	if !ok {
		return
	}

	for i, item := range v.([]interface{}) {
		data := item.(map[string]interface{})
		name := data["name"].(string)
		value := data["value"].(string)

		if err := validateName(name); err != nil {
			gen.errors = append(gen.errors, fmt.Sprintf("constant[%d] %q: %s (use snake_case identifiers)", i, name, err))
			continue
		}

		var js interface{}
		if err := json.Unmarshal([]byte(value), &js); err != nil {
			gen.errors = append(gen.errors, fmt.Sprintf("constant[%d] %q: value is not valid JSON - use jsonencode() for complex values", i, name))
			continue
		}

		gen.constants[name] = value
	}
}

func (gen *generator) parsePredicates(d *schema.ResourceData) {
	v, ok := d.GetOk("predicate")
	if !ok {
		return
	}

	for i, item := range v.([]interface{}) {
		data := item.(map[string]interface{})
		name := data["name"].(string)

		if err := validateName(name); err != nil {
			gen.errors = append(gen.errors, fmt.Sprintf("predicate[%d] %q: %s (use snake_case identifiers)", i, name, err))
			continue
		}

		if _, exists := gen.constants[name]; exists {
			gen.errors = append(gen.errors, fmt.Sprintf("predicate[%d] %q: name already used by a constant - choose a different name", i, name))
			continue
		}

		pred := predicate{name: name}

		if conds, ok := data["condition"]; ok {
			for _, c := range conds.([]interface{}) {
				pred.conditions = append(pred.conditions, parseCondition(c.(map[string]interface{})))
			}
		}

		if len(pred.conditions) == 0 {
			gen.errors = append(gen.errors, fmt.Sprintf(
				"predicate[%d] %q: requires at least one 'condition' block", i, name))
			continue
		}

		gen.predicates[name] = append(gen.predicates[name], pred)
	}
}

func (gen *generator) parseRules(d *schema.ResourceData) {
	v, ok := d.GetOk("rule")
	if !ok {
		return
	}

	for i, item := range v.([]interface{}) {
		data := item.(map[string]interface{})

		rl := rule{
			name:      data["name"].(string),
			isDefault: data["default"].(bool),
			index:     i,
		}

		if effectList, ok := data["effect"].([]interface{}); ok && len(effectList) > 0 {
			rl.effect = parseEffect(effectList[0].(map[string]interface{}))
		}

		if whenList, ok := data["when"].([]interface{}); ok && len(whenList) > 0 {
			whenData := whenList[0].(map[string]interface{})
			if refs, ok := whenData["all_of"]; ok {
				for _, ref := range refs.([]interface{}) {
					rl.whenAllOf = append(rl.whenAllOf, ref.(string))
				}
			}
		}

		if rl.isDefault && len(rl.whenAllOf) > 0 {
			gen.errors = append(gen.errors, fmt.Sprintf(
				"rule[%d] %q: default rules cannot have 'when' - they define fallback behavior", i, rl.name))
		}

		needsTarget := rl.effect.action == "mask" || rl.effect.action == "decrypt"
		hasTarget := rl.effect.dataLabel != "" || rl.effect.columnName != "" || rl.effect.allColumns
		if needsTarget && !hasTarget {
			gen.errors = append(gen.errors, fmt.Sprintf(
				"rule[%d] %q: %s effect needs a target - add 'data_label', 'column_name', or 'all_columns = true' to the effect block", i, rl.name, rl.effect.action))
		}

		gen.rules = append(gen.rules, rl)
	}
}

func (gen *generator) parseRawRego(d *schema.ResourceData) {
	if v, ok := d.GetOk("raw_rego"); ok {
		gen.rawRego = v.(string)
	}
}

func parseCondition(m map[string]interface{}) condition {
	c := condition{test: m["test"].(string)}
	if v := m["variable"]; v != nil {
		c.variable = v.(string)
	}
	if v := m["values"]; v != nil {
		for _, val := range v.([]interface{}) {
			c.values = append(c.values, val.(string))
		}
	}
	if v := m["constant"]; v != nil {
		c.constantRef = v.(string)
	}
	return c
}

func (c condition) validate(context string) []string {
	var errs []string

	if c.variable == "" {
		errs = append(errs, fmt.Sprintf("%s: missing 'variable' - specify the input path to check (e.g., \"user.groups\", \"resource.name\")", context))
	}

	needsValue := c.test != "exists" && c.test != "not_exists"
	if needsValue && len(c.values) == 0 && c.constantRef == "" {
		errs = append(errs, fmt.Sprintf("%s: %q test requires 'values' or 'constant' to compare against", context, c.test))
	}

	return errs
}

func parseEffect(m map[string]interface{}) effect {
	e := effect{action: m["action"].(string)}
	if v := m["type"]; v != nil {
		e.effectType = v.(string)
	}
	if v := m["sub_type"]; v != nil {
		e.subType = v.(string)
	}
	if v := m["typesafe"]; v != nil {
		e.typesafe = v.(string)
	}
	if v := m["message"]; v != nil {
		e.message = v.(string)
	}
	if v := m["reason"]; v != nil {
		e.reason = v.(string)
	}
	if v := m["data_label"]; v != nil {
		e.dataLabel = v.(string)
	}
	if v := m["column_name"]; v != nil {
		e.columnName = v.(string)
	}
	if v := m["all_columns"]; v != nil {
		e.allColumns = v.(bool)
	}
	return e
}

func (g *generator) validateReferences() {
	definedPredicates := make([]string, 0, len(g.predicates))
	for name := range g.predicates {
		definedPredicates = append(definedPredicates, name)
	}
	sort.Strings(definedPredicates)

	for name, preds := range g.predicates {
		for _, pred := range preds {
			for i, cond := range pred.conditions {
				context := fmt.Sprintf("predicate %q condition[%d]", name, i)
				g.errors = append(g.errors, cond.validate(context)...)
				if cond.constantRef != "" {
					if _, ok := g.constants[cond.constantRef]; !ok {
						definedConstants := make([]string, 0, len(g.constants))
						for c := range g.constants {
							definedConstants = append(definedConstants, c)
						}
						g.errors = append(g.errors, fmt.Sprintf(
							"predicate %q condition[%d]: references undefined constant %q - available constants: %v", name, i, cond.constantRef, definedConstants))
					}
				}
			}
		}
	}

	for i, r := range g.rules {
		for _, ref := range r.whenAllOf {
			cleanRef := strings.TrimPrefix(ref, "!")
			if _, ok := g.predicates[cleanRef]; !ok {
				g.errors = append(g.errors, fmt.Sprintf(
					"rule[%d] %q: references undefined predicate %q in when.all_of - available predicates: %v", i, r.name, cleanRef, definedPredicates))
			}
		}
	}
}

func (g *generator) generate() (string, error) {
	var sb strings.Builder

	sb.WriteString("package formal.v2\n\n")

	needsIf := len(g.predicates) > 0 || len(g.rules) > 0
	needsIn := g.usesInOperator()

	if needsIf {
		sb.WriteString("import future.keywords.if\n")
	}
	if needsIn {
		sb.WriteString("import future.keywords.in\n")
	}
	if needsIf || needsIn {
		sb.WriteString("\n")
	}

	if g.description != "" {
		sb.WriteString(fmt.Sprintf("# %s\n\n", g.description))
	}

	if len(g.includedConnectors) > 0 {
		connJSON, _ := json.Marshal(g.includedConnectors)
		sb.WriteString(fmt.Sprintf("included_connectors := %s\n\n", connJSON))
	}

	if len(g.constants) > 0 {
		names := make([]string, 0, len(g.constants))
		for name := range g.constants {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			sb.WriteString(fmt.Sprintf("%s := %s\n", name, g.constants[name]))
		}
		sb.WriteString("\n")
	}

	predicateNames := make([]string, 0, len(g.predicates))
	for name := range g.predicates {
		predicateNames = append(predicateNames, name)
	}
	sort.Strings(predicateNames)

	for _, name := range predicateNames {
		for _, pred := range g.predicates[name] {
			sb.WriteString(g.generatePredicate(pred))
			sb.WriteString("\n")
		}
	}

	if g.rawRego != "" {
		sb.WriteString(g.rawRego)
		sb.WriteString("\n\n")
	}

	if len(g.rules) > 0 {
		sortedRules := make([]rule, len(g.rules))
		copy(sortedRules, g.rules)
		sort.SliceStable(sortedRules, func(i, j int) bool {
			if sortedRules[i].isDefault != sortedRules[j].isDefault {
				return !sortedRules[i].isDefault
			}
			return sortedRules[i].index < sortedRules[j].index
		})

		for _, r := range sortedRules {
			sb.WriteString(g.generateRule(r))
			sb.WriteString("\n")
		}
	}

	return strings.TrimRight(sb.String(), "\n") + "\n", nil
}

func (g *generator) usesInOperator() bool {
	for _, preds := range g.predicates {
		for _, pred := range preds {
			for _, cond := range pred.conditions {
				if inOperators[cond.test] {
					return true
				}
			}
		}
	}
	return false
}

func (g *generator) generatePredicate(pred predicate) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s if {\n", pred.name))

	for _, cond := range pred.conditions {
		sb.WriteString(fmt.Sprintf("    %s\n", g.generateCondition(cond)))
	}

	sb.WriteString("}\n")
	return sb.String()
}

func (g *generator) generateCondition(cond condition) string {
	varPath := "input." + cond.variable

	var compareValue string
	if cond.constantRef != "" {
		compareValue = cond.constantRef
	} else if len(cond.values) == 1 {
		if inOperators[cond.test] {
			compareValue = "[" + formatValue(cond.values[0]) + "]"
		} else {
			compareValue = formatValue(cond.values[0])
		}
	} else if len(cond.values) > 1 {
		parts := make([]string, len(cond.values))
		for i, v := range cond.values {
			parts[i] = formatValue(v)
		}
		compareValue = "[" + strings.Join(parts, ", ") + "]"
	}

	switch cond.test {
	case "equals":
		return fmt.Sprintf("%s == %s", varPath, compareValue)
	case "not_equals":
		return fmt.Sprintf("%s != %s", varPath, compareValue)
	case "in":
		return fmt.Sprintf("%s in %s", varPath, compareValue)
	case "not_in":
		return fmt.Sprintf("not %s in %s", varPath, compareValue)
	case "any_in":
		return fmt.Sprintf("some _x in %s; _x in %s", varPath, compareValue)
	case "all_in":
		return fmt.Sprintf("every _x in %s { _x in %s }", varPath, compareValue)
	case "none_in":
		return fmt.Sprintf("not (some _x in %s; _x in %s)", varPath, compareValue)
	case "contains":
		return fmt.Sprintf("contains(%s, %s)", varPath, compareValue)
	case "not_contains":
		return fmt.Sprintf("not contains(%s, %s)", varPath, compareValue)
	case "starts_with":
		return fmt.Sprintf("startswith(%s, %s)", varPath, compareValue)
	case "ends_with":
		return fmt.Sprintf("endswith(%s, %s)", varPath, compareValue)
	case "regex":
		return fmt.Sprintf("regex.match(%s, %s)", compareValue, varPath)
	case "greater_than":
		return fmt.Sprintf("%s > %s", varPath, compareValue)
	case "less_than":
		return fmt.Sprintf("%s < %s", varPath, compareValue)
	case "greater_than_or_equal":
		return fmt.Sprintf("%s >= %s", varPath, compareValue)
	case "less_than_or_equal":
		return fmt.Sprintf("%s <= %s", varPath, compareValue)
	case "exists":
		return varPath
	case "not_exists":
		return fmt.Sprintf("not %s", varPath)
	default:
		return ""
	}
}

func (g *generator) generateRule(r rule) string {
	var sb strings.Builder

	effectObj := g.generateEffect(r.effect)

	if r.isDefault {
		sb.WriteString(fmt.Sprintf("default %s := %s\n", r.name, effectObj))
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("%s := %s if {\n", r.name, effectObj))

	for _, ref := range r.whenAllOf {
		if strings.HasPrefix(ref, "!") {
			sb.WriteString(fmt.Sprintf("    not %s\n", strings.TrimPrefix(ref, "!")))
		} else {
			sb.WriteString(fmt.Sprintf("    %s\n", ref))
		}
	}

	if r.effect.dataLabel != "" || r.effect.columnName != "" {
		sb.WriteString(fmt.Sprintf("    %s\n", g.generateTargetAssignment(r.effect)))
	}

	sb.WriteString("}\n")
	return sb.String()
}

func (g *generator) generateEffect(e effect) string {
	parts := []string{fmt.Sprintf(`"action": %q`, e.action)}

	if e.effectType != "" {
		parts = append(parts, fmt.Sprintf(`"type": %q`, e.effectType))
	}
	if e.subType != "" {
		parts = append(parts, fmt.Sprintf(`"sub_type": %q`, e.subType))
	}
	if e.typesafe != "" {
		parts = append(parts, fmt.Sprintf(`"typesafe": %q`, e.typesafe))
	}
	if e.message != "" {
		parts = append(parts, fmt.Sprintf(`"message": %q`, e.message))
	}
	if e.reason != "" {
		parts = append(parts, fmt.Sprintf(`"reason": %q`, e.reason))
	}

	if e.allColumns {
		parts = append(parts, `"columns": input.columns`)
	} else if e.dataLabel != "" || e.columnName != "" {
		parts = append(parts, `"columns": columns`)
	}

	return "{ " + strings.Join(parts, ", ") + " }"
}

func (g *generator) generateTargetAssignment(e effect) string {
	var match string
	if e.dataLabel != "" {
		match = fmt.Sprintf(`col["data_label"] == %q`, e.dataLabel)
	} else {
		match = fmt.Sprintf(`col["name"] == %q`, e.columnName)
	}
	return fmt.Sprintf(`columns := [col | col := input.columns[_]; %s]`, match)
}

func formatValue(v string) string {
	var js interface{}
	if err := json.Unmarshal([]byte(v), &js); err == nil {
		return v
	}
	return fmt.Sprintf("%q", v)
}
