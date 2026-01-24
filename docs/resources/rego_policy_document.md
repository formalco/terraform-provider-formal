---
page_title: "formal_rego_policy_document Data Source - terraform-provider-formal"
subcategory: ""
description: |-
  Generates Rego policy code from declarative predicates and rules.
---

# formal_rego_policy_document (Data Source)

Generates Rego policy code from declarative predicates and rules. This data source allows you to compose Formal policies without writing raw Rego code.

## How It Works

1. **Predicates** are atomic boolean checks on input values
2. **Rules** compose predicates using `when.all_of` (AND) and negation (`!`)
3. **OR logic** is achieved via multiple rule blocks with the same name

## Example Usage

### Mask Columns by Data Label

```hcl
data "formal_rego_policy_document" "mask_emails" {
  rule {
    name = "post_request"
    effect {
      action     = "mask"
      type       = "redact.partial"
      sub_type   = "email_mask_username"
      typesafe   = "fallback_to_default"
      data_label = "email_address"
    }
  }
}

resource "formal_policy" "mask_emails" {
  name   = "mask-email"
  module = data.formal_rego_policy_document.mask_emails.rego
}
```

### Block by Default, Allow Specific Users

```hcl
data "formal_rego_policy_document" "block_db" {
  predicate {
    name = "is_main_db"
    condition {
      test     = "equals"
      variable = "db_name"
      values   = ["main"]
    }
  }

  predicate {
    name = "is_analyst"
    condition {
      test     = "any_in"
      variable = "user.groups"
      values   = ["analyst"]
    }
  }

  # Default: block everyone
  rule {
    name    = "session"
    default = true
    effect {
      action = "block"
      type   = "block_with_formal_message"
    }
  }

  # Allow analysts on main db
  rule {
    name = "session"
    effect {
      action = "allow"
      reason = "User is authorized"
    }
    when {
      all_of = ["is_main_db", "is_analyst"]
    }
  }
}
```

### OR Logic via Multiple Rules

```hcl
data "formal_rego_policy_document" "allow_privileged" {
  predicate {
    name = "is_admin"
    condition {
      test     = "any_in"
      variable = "user.groups"
      values   = ["admin"]
    }
  }

  predicate {
    name = "is_owner"
    condition {
      test     = "any_in"
      variable = "user.groups"
      values   = ["owner"]
    }
  }

  # Allow if admin OR owner (two rules with same name)
  rule {
    name = "session"
    effect { action = "allow" }
    when { all_of = ["is_admin"] }
  }

  rule {
    name = "session"
    effect { action = "allow" }
    when { all_of = ["is_owner"] }
  }
}
```

### Using Constants

```hcl
data "formal_rego_policy_document" "with_constants" {
  constant {
    name  = "allowed_groups"
    value = jsonencode(["analysts", "engineers", "admins"])
  }

  predicate {
    name = "has_allowed_group"
    condition {
      test     = "any_in"
      variable = "user.groups"
      constant = "allowed_groups"
    }
  }

  rule {
    name = "session"
    effect { action = "allow" }
    when { all_of = ["has_allowed_group"] }
  }
}
```

### Using raw_rego for Complex Logic

```hcl
data "formal_rego_policy_document" "complex" {
  raw_rego = <<-EOF
# Row-level check for EU data
is_eu_row if {
    some col in input.row
    col.name == "region"
    col.value == "eu"
}
EOF

  rule {
    name = "post_request"
    effect {
      action      = "mask"
      type        = "hash.with_salt"
      all_columns = true
    }
    when {
      all_of = ["is_eu_row"]
    }
  }
}
```

## Schema

### Optional

- `description` (String) - Description comment at the top of the generated Rego.
- `included_connectors` (List of String) - Connector names this policy applies to.
- `constant` (Block List) - Named constants for use in predicates.
- `predicate` (Block List) - Named boolean predicates.
- `rule` (Block List) - Policy rules that produce effects.
- `raw_rego` (String) - Raw Rego code for custom logic.

### Read-Only

- `rego` (String) - The generated Rego policy code.
- `id` (String) - SHA256 hash of the generated Rego.

### Nested Schema for `constant`

- `name` (String, Required) - Constant name.
- `value` (String, Required) - JSON value. Use `jsonencode()`.

### Nested Schema for `predicate`

- `name` (String, Required) - Predicate name.
- `condition` (Block List, Required) - Conditions that must all be true.

### Nested Schema for `condition`

- `test` (String, Required) - Operator: `equals`, `not_equals`, `in`, `not_in`, `any_in`, `all_in`, `none_in`, `contains`, `not_contains`, `starts_with`, `ends_with`, `regex`, `greater_than`, `less_than`, `greater_than_or_equal`, `less_than_or_equal`, `exists`, `not_exists`.
- `variable` (String, Required) - Input path (e.g., `user.groups`).
- `values` (List of String) - Values to compare.
- `constant` (String) - Reference a constant instead.

### Nested Schema for `rule`

- `name` (String, Required) - Rule name (`session`, `pre_request`, `post_request`).
- `default` (Boolean) - If true, this is the fallback rule.
- `effect` (Block, Required) - The effect when rule matches.
- `when` (Block) - Predicate references.

### Nested Schema for `effect`

- `action` (String, Required) - `allow`, `block`, `mask`, `decrypt`.
- `type` (String) - Action subtype.
- `sub_type` (String) - Further subtype.
- `typesafe` (String) - Typesafe behavior.
- `message` (String) - Custom message for block.
- `reason` (String) - Reason for allow.
- `data_label` (String) - Target columns by label.
- `column_name` (String) - Target columns by name.
- `all_columns` (Boolean) - Target all columns.

### Nested Schema for `when`

- `all_of` (List of String) - Predicates that must all be true. Use `!` prefix for NOT.
