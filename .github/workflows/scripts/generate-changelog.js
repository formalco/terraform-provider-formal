#!/usr/bin/env node

/**
 * Generate changelog for Terraform Provider release
 * This script is called from the GitHub Actions workflow
 */

const axios = require('axios');

async function main(params) {
  const { github, context, core, process: proc } = params;
  const semver = require('semver');

  const components = {
    'v': { label: 'provider', color: 'f29513' },
  };

  const tag = context.ref.split('/').pop();
  console.log(`Processing tag: ${tag}`);
  const tagComponent = tag.charAt(0);
  const version = tag.substring(1);

  const config = components[tagComponent];
  if (!config) {
    console.log(`Unknown component: ${tagComponent}`);
    core.setFailed('Unknown component');
    return;
  }

  // Get previous tag
  const tags = await github.rest.git.listMatchingRefs({
    owner: context.repo.owner,
    repo: context.repo.repo,
    ref: `tags/${tagComponent}-`
  });

  const sortedTags = tags.data
    .map(tag => tag.ref.split('/').pop())
    .filter(tag => tag.startsWith(`${tagComponent}-`))
    .sort((a, b) => {
      const vA = a.split('-')[1];
      const vB = b.split('-')[1];
      return -semver.compare(vA, vB);
    });

  const currentIndex = sortedTags.indexOf(tag);
  const previousTag = currentIndex < sortedTags.length - 1 ? sortedTags[currentIndex + 1] : null;
  console.log(`Previous tag: ${previousTag}`);
  let since = '';
  if (previousTag) {
    try {
      const prevTagData = await github.rest.git.getRef({
        owner: context.repo.owner,
        repo: context.repo.repo,
        ref: `tags/${previousTag}`
      });
      console.log(`Previous tag data:`, prevTagData.data);

      let commitSha = prevTagData.data.object.sha;

      // If the tag object is an annotated tag, fetch the tag details to get the commit SHA
      if (prevTagData.data.object.type === 'tag') {
        const tagData = await github.rest.git.getTag({
          owner: context.repo.owner,
          repo: context.repo.repo,
          tag_sha: commitSha
        });
        commitSha = tagData.data.object.sha; // The commit SHA from the annotated tag
      }

      const prevTagCommit = await github.rest.git.getCommit({
        owner: context.repo.owner,
        repo: context.repo.repo,
        commit_sha: commitSha
      });
      since = prevTagCommit.data.author.date;
    } catch (error) {
      console.error(`Error fetching previous tag commit: ${error.message}`);
      core.setFailed('Failed to fetch previous tag commit');
      return;
    }
  }
  console.log(`Since: ${since}`);
  // Fetch PRs for this component
  const searchQuery = `repo:${context.repo.owner}/${context.repo.repo} is:pr is:merged base:main label:${config.label} ${since ? `merged:>${since}` : ''}`;
  console.log(`Search query: ${searchQuery}`);

  const result = await github.rest.search.issuesAndPullRequests({
    q: searchQuery,
    per_page: 100
  });

  if (result.data.items.length === 0) {
    console.log(`No PRs found for ${tagComponent}`);
    core.setOutput('changelog_content', `# ${tagComponent} ${version}\n\nNo changes in this release.`);
    core.setOutput('component', tagComponent);
    core.setOutput('version', version);
    core.setOutput('branch_component', tagComponent);
    return;
  }

  // Prepare PR data for OpenAI
  const prData = result.data.items.map(pr => ({
    title: pr.title,
    number: pr.number,
    url: pr.html_url,
    labels: pr.labels.map(l => l.name.toLowerCase()),
    body: pr.body
  }));

  // Generate changelog using OpenAI
  const openaiPrompt = {
    model: "gpt-5",
    messages: [{
      role: "system",
      content: "You are a technical writer creating an external-facing changelog for our customers. Your goal is to communicate **high-level functionality changes**, not internal implementation details like code refactors, function changes, or technical restructuring.\n\n" +
        "### Changelog Format:\n" +
        "- Each release must include only the relevant sections: \n" +
        "  - **### New** (for new features)\n" +
        "  - **### Fixed** (for bug fixes)\n" +
        "  - **### Changed** (for modifications to existing functionality)\n" +
        "- Omit any section that has no changes.\n" +
        "- Each bullet point should be **short, clear, and impact-focused**.\n" +
        "- Keep bullet points between 15-120 characters.\n" +
        "- Use a single line per change; only use two lines for complex changes that cannot be simplified further.\n\n" +

        "### What to Include:\n" +
        "1. Extract only customer-facing changes of PRs.\n" +
        "2. **Choose only one category per change** (do not list the same item in multiple sections).\n" +
        "3. **Explain why the change matters** to users, avoiding technical jargon.\n" +
        "4. **Do not include PR numbers, internal function names, or implementation details.**\n" +
        "5. **Start each bullet point with a verb** in present tense (e.g., 'Add', 'Fix', 'Update').\n" +
        "6. **Group related changes** under a single bullet point to avoid fragmentation.\n\n" +

        "### Grouping Related Changes:\n" +
        "- Combine related changes into a single, comprehensive bullet point\n" +
        "- Include all relevant aspects of the feature or change\n" +
        "- Use commas or 'with' to connect related components\n\n" +

        "Example of good grouping:\n" +
        "✓ Add PDF export with custom headers, watermarks, and page numbering\n" +
        "✗ Add PDF export\n" +
        "✗ Add custom headers to PDF export\n" +
        "✗ Add watermarks to PDF export\n" +
        "✗ Add page numbering to PDF export\n\n" +

        "### Example Output:\n" +
        "```\n" +
        "## 1.12.37 (2025-02-06)\n" +
        "\n" +
        "### New\n" +
        "- Add support for masking policies on columns with shared paths in Snowflake, improving data security and compliance.\n" +
        "\n" +
        "### Changed\n" +
        "- Improve masking policy enforcement for better consistency across similar columns, reducing manual configuration.\n" +
        "```\n" +
        "\n" +
        "Follow this structure strictly. The focus should always be on how the changes affect customers, **not on how the changes were implemented.**"
    }, {
      role: "user",
      content: `Generate a changelog for our Terraform Provider version ${version} based on these Pull Requests:\n${JSON.stringify(prData, null, 2)}`
    }]
  };

  try {
    const openaiResponse = await axios.post('https://api.openai.com/v1/chat/completions', openaiPrompt, {
      headers: {
        'Authorization': `Bearer ${proc.env.OPENAI_API_KEY}`,
        'Content-Type': 'application/json'
      }
    });

    let changelogContent = openaiResponse.data.choices[0].message.content;

    // Only remove title if it starts with # but not with ###
    if (changelogContent.startsWith('#') && !changelogContent.startsWith('###')) {
      const firstSectionIndex = changelogContent.indexOf('### ');
      if (firstSectionIndex !== -1) {
        changelogContent = changelogContent.substring(firstSectionIndex);
      }
    }

    // Determine tags based on changelog content
    const tags = [];
    if (changelogContent.includes('### New')) tags.push('New Features');
    if (changelogContent.includes('### Fixed')) tags.push('Bug Fixes');
    if (changelogContent.includes('### Changed')) tags.push('Improvements');

    const tagsString = tags.length > 0 ? ` tags={${JSON.stringify(tags)}}` : '';

    // Format date as "Month Day, Year" (e.g., "October 15, 2025")
    const formatDate = (dateStr) => {
      const d = new Date(dateStr);
      const monthNames = ['January', 'February', 'March', 'April', 'May', 'June',
                         'July', 'August', 'September', 'October', 'November', 'December'];
      return `${monthNames[d.getMonth()]} ${d.getDate()}, ${d.getFullYear()}`;
    };

    const today = new Date();
    const formattedDate = formatDate(today);

    // Return with Update wrapper for Mintlify
    const formattedChangelog = `<Update label="${formattedDate}"${tagsString}>

## ${version}

${changelogContent.trim()}

</Update>`;

    // Base64 encode the changelog content to avoid bash escaping issues
    const changelogBase64 = Buffer.from(formattedChangelog).toString('base64');
    core.setOutput('changelog_content', changelogBase64);

    core.setOutput('component', "Formal Terraform Provider");
    core.setOutput('branch_component', "provider");
    core.setOutput('version', version);

    // Generate PR list for the PR description
    const githubRepo = context.repo.owner + '/' + context.repo.repo;
    let prListMarkdown = `\n### ${version}\n`;
    for (const pr of prData) {
      prListMarkdown += `- [#${pr.number}](${pr.url}): ${pr.title}\n`;
    }
    core.setOutput('pr_list', prListMarkdown);

  } catch (error) {
    console.error('OpenAI API Error:', error.response?.data || error);
    core.setFailed('Failed to generate changelog with OpenAI');
  }
}

module.exports = main;
