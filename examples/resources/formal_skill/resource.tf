# One skill per skills/<name>/SKILL.md file.
resource "formal_skill" "repo" {
  for_each = fileset("${path.module}/skills", "*/SKILL.md")
  skill_md = file("${path.module}/skills/${each.value}")
}

# Only release managers can use the release-notes skill.
resource "formal_policy" "release_notes_skill" {
  name        = "Skill access: release-notes"
  description = "Only release managers can use the release-notes skill."
  status      = "active"
  module      = <<-EOT
    package formal.v2

    import future.keywords.if
    import future.keywords.in

    request := {
      "action": "block",
      "type": "block_with_formal_message",
      "reason": "Only release managers can use the release-notes skill.",
    } if {
      input.resource.technology == "llm"
      some skill in input.llm.loaded_skills
      skill.name == "${formal_skill.repo["release-notes/SKILL.md"].name}"
      not "release-managers" in input.user.groups
    }
  EOT
}
