resource "formal_hook" "risk" {
  name        = "risk"
  description = "Score request risk for policy decisions"
  status      = "active"
  timeout_ms  = 5000
  code        = <<-JS
    export default function hook(input) {
      return { score: 1 };
    }
  JS
}

resource "formal_policy" "block_high_risk" {
  depends_on  = [formal_hook.risk]
  name        = "block-high-risk"
  description = "Block requests when the risk hook scores high"
  status      = "active"
  module      = <<-REGO
    package formal.v2

    import future.keywords.if

    pre_request := {"action": "block"} if {
      input.hooks.risk.score >= 5
    }
  REGO
}
