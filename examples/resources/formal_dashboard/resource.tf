resource "formal_dashboard" "ai_usage" {
  name = "AI usage"
  definition = jsonencode({
    layout = { columns = 1 }
    widgets = [{
      key    = "providers"
      layout = { span = 1 }
      logs = {
        visualization = "DASHBOARD_VISUALIZATION_BARS"
        group_by      = ["llm.provider"]
      }
    }]
  })
}
