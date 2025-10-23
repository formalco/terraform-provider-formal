#!/usr/bin/env node

/**
 * Terraform Provider Changelog Generation Script
 *
 * This script:
 * 1. Finds the latest version in the docs changelog
 * 2. Finds ALL missing versions from git tags
 * 3. Uses git log to find commits between tags
 * 4. Generates changelog via LLM using PR descriptions
 * 5. Processes all missing versions in a single run (oldest to newest)
 *
 * Usage: node generate-changelog-v2.js [--mode auto|latest] [--docs-repo-path <path>]
 */

const { execSync } = require('child_process');
const fs = require('fs');
const https = require('https');

/**
 * Execute shell command and return output
 */
function exec(command) {
  try {
    return execSync(command, {
      encoding: 'utf8',
      stdio: ['pipe', 'pipe', 'pipe']
    }).trim();
  } catch (error) {
    return '';
  }
}

/**
 * Make GitHub API search request
 */
async function githubSearchApi(query, token) {
  return new Promise((resolve, reject) => {
    const encodedQuery = encodeURIComponent(query);
    const options = {
      hostname: 'api.github.com',
      path: `/search/issues?q=${encodedQuery}&per_page=100`,
      headers: {
        'Authorization': `token ${token}`,
        'User-Agent': 'Formal-Changelog-Bot',
        'Accept': 'application/vnd.github.v3+json'
      }
    };

    https.get(options, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          const parsed = JSON.parse(data);
          resolve(parsed.items || []);
        } else {
          reject(new Error(`GitHub API error: ${res.statusCode} - ${data}`));
        }
      });
    }).on('error', reject);
  });
}

/**
 * Make OpenAI API request
 */
async function openaiApi(messages, apiKey, model = 'gpt-5') {
  return new Promise((resolve, reject) => {
    const postData = JSON.stringify({ model, messages });

    const options = {
      hostname: 'api.openai.com',
      path: '/v1/chat/completions',
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${apiKey}`,
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(postData)
      }
    };

    const req = https.request(options, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(JSON.parse(data));
        } else {
          reject(new Error(`OpenAI API error: ${res.statusCode} - ${data}`));
        }
      });
    });

    req.on('error', reject);
    req.write(postData);
    req.end();
  });
}

/**
 * Parse latest version from Mintlify changelog
 */
function parseLatestVersionFromChangelog(changelogPath) {
  if (!fs.existsSync(changelogPath)) {
    console.log(`Changelog file does not exist: ${changelogPath}`);
    return null;
  }

  const content = fs.readFileSync(changelogPath, 'utf8');
  const versionRegex = /^##\s+(\d+\.\d+\.\d+)/m;
  const match = content.match(versionRegex);

  if (match) {
    console.log(`Latest version in changelog: ${match[1]}`);
    return match[1];
  }

  console.log('No versions found in changelog');
  return null;
}

/**
 * Get all version tags from git
 */
function getAllVersionTags() {
  const tags = exec('git tag -l "v*" --sort=-version:refname');
  if (!tags) return [];

  return tags.split('\n')
    .filter(tag => tag.match(/^v\d+\.\d+\.\d+$/))
    .map(tag => tag.substring(1)); // Remove 'v' prefix
}

/**
 * Find all missing versions that need changelogs
 */
function findAllMissingVersions(mode, docsRepoPath) {
  const changelogPath = `${docsRepoPath}/docs/changelog/terraform-provider.mdx`;
  const latestInChangelog = parseLatestVersionFromChangelog(changelogPath);
  const allVersions = getAllVersionTags();

  if (allVersions.length === 0) {
    console.log('No version tags found');
    return [];
  }

  const newestVersion = allVersions[0];
  const missingVersionPairs = [];

  // If mode is 'latest', only process the newest version
  if (mode === 'latest') {
    console.log(`Mode is 'latest', processing only version ${newestVersion}`);
    const newestIndex = 0;
    const previousVersion = allVersions[1] || null;
    return [{ version: newestVersion, previousVersion }];
  }

  // If no changelog exists yet, process ALL versions from oldest to newest
  if (!latestInChangelog) {
    console.log(`No changelog exists yet, will process all ${allVersions.length} versions`);

    for (let i = allVersions.length - 1; i >= 0; i--) {
      const version = allVersions[i];
      const previousVersion = i < allVersions.length - 1 ? allVersions[i + 1] : null;
      missingVersionPairs.push({ version, previousVersion });
    }

    return missingVersionPairs;
  }

  // If changelog already has the newest version, we're done
  if (latestInChangelog === newestVersion) {
    console.log(`Changelog is already up to date at version ${newestVersion}`);
    return [];
  }

  // Find the index of the current latest version in the changelog
  const latestChangelogIndex = allVersions.indexOf(latestInChangelog);

  if (latestChangelogIndex === -1) {
    console.error(`Version ${latestInChangelog} from changelog not found in git tags`);
    console.log('Will process all versions');

    for (let i = allVersions.length - 1; i >= 0; i--) {
      const version = allVersions[i];
      const previousVersion = i < allVersions.length - 1 ? allVersions[i + 1] : null;
      missingVersionPairs.push({ version, previousVersion });
    }

    return missingVersionPairs;
  }

  // Build list of missing versions from oldest to newest
  for (let i = latestChangelogIndex; i > 0; i--) {
    const version = allVersions[i - 1];
    const previousVersion = allVersions[i];
    missingVersionPairs.push({ version, previousVersion });
  }

  console.log(`Found ${missingVersionPairs.length} missing versions to process`);
  if (missingVersionPairs.length > 0) {
    console.log(`Range: ${missingVersionPairs[0].version} → ${missingVersionPairs[missingVersionPairs.length - 1].version}`);
  }

  return missingVersionPairs;
}

/**
 * Get commit hash for a version tag
 */
function getCommitForTag(version) {
  return exec(`git rev-list -n 1 v${version}`);
}

/**
 * Find commits between two tags
 */
function findCommitsBetweenVersions(newVersion, oldVersion) {
  const newCommit = getCommitForTag(newVersion);

  if (!newCommit) {
    console.error(`Could not find commit for version ${newVersion}`);
    return [];
  }

  let range;
  if (oldVersion) {
    const oldCommit = getCommitForTag(oldVersion);
    if (oldCommit) {
      range = `${oldCommit}..${newCommit}`;
    } else {
      range = newCommit;
    }
  } else {
    range = newCommit;
  }

  const commits = exec(`git log ${range} --format=%H`);
  return commits ? commits.split('\n').filter(Boolean) : [];
}

/**
 * Get PR details from commits
 */
async function getPRDetailsFromCommits(commits, githubRepo, githubToken) {
  if (commits.length === 0) return [];

  const prNumbers = new Set();

  // Extract PR numbers from commit messages
  for (const commit of commits) {
    const message = exec(`git log -1 --format=%s ${commit}`);
    const prMatch = message.match(/#(\d+)/);
    if (prMatch) {
      prNumbers.add(prMatch[1]);
    }
  }

  if (prNumbers.size === 0) {
    console.log('No PR references found in commits');
    return [];
  }

  console.log(`Found ${prNumbers.size} unique PRs in commits`);

  // Fetch PR details from GitHub
  const prDetails = [];
  for (const prNumber of prNumbers) {
    const query = `repo:${githubRepo} is:pr is:merged label:provider ${prNumber}`;
    const results = await githubSearchApi(query, githubToken);

    if (results.length > 0) {
      const pr = results[0];
      prDetails.push({
        number: pr.number,
        title: pr.title,
        url: pr.html_url,
        body: pr.body || '',
        labels: pr.labels.map(l => l.name)
      });
    }
  }

  console.log(`Found ${prDetails.length} PRs with 'provider' label`);
  return prDetails;
}

/**
 * Generate changelog using LLM
 */
async function generateChangelogWithLLM(version, prDetails, openaiKey, date) {
  const systemPrompt = `You are a technical writer creating an external-facing changelog for our customers. Your goal is to communicate **high-level functionality changes**, not internal implementation details like code refactors, function changes, or technical restructuring.

### Changelog Format:
- Each release must include only the relevant sections:
  - **### New** (for new features)
  - **### Fixed** (for bug fixes)
  - **### Changed** (for modifications to existing functionality)
- Omit any section that has no changes.
- Each bullet point should be **short, clear, and impact-focused**.
- Keep bullet points between 15-120 characters.
- Use a single line per change; only use two lines for complex changes that cannot be simplified further.

### What to Include:
1. Extract only customer-facing changes of PRs.
2. **Choose only one category per change** (do not list the same item in multiple sections).
3. **Explain why the change matters** to users, avoiding technical jargon.
4. **Do not include PR numbers, internal function names, or implementation details.**
5. **Start each bullet point with a verb** in present tense (e.g., 'Add', 'Fix', 'Update').
6. **Group related changes** under a single bullet point to avoid fragmentation.

### Grouping Related Changes:
- Combine related changes into a single, comprehensive bullet point
- Include all relevant aspects of the feature or change
- Use commas or 'with' to connect related components

Example of good grouping:
✓ Add PDF export with custom headers, watermarks, and page numbering
✗ Add PDF export
✗ Add custom headers to PDF export
✗ Add watermarks to PDF export
✗ Add page numbering to PDF export

Follow this structure strictly. The focus should always be on how the changes affect customers, **not on how the changes were implemented.**`;

  const userPrompt = `Generate a changelog for Terraform Provider version ${version} based on these Pull Requests:\n${JSON.stringify(prDetails, null, 2)}`;

  try {
    const response = await openaiApi([
      { role: 'system', content: systemPrompt },
      { role: 'user', content: userPrompt }
    ], openaiKey, 'gpt-5');

    let changelogContent = response.choices[0].message.content;

    // Remove title if it starts with # but not with ###
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

    // Format date
    const formatDate = (dateStr) => {
      const d = new Date(dateStr);
      const monthNames = ['January', 'February', 'March', 'April', 'May', 'June',
                         'July', 'August', 'September', 'October', 'November', 'December'];
      return `${monthNames[d.getMonth()]} ${d.getDate()}, ${d.getFullYear()}`;
    };

    const formattedDate = formatDate(date);

    return `<Update label="${formattedDate}"${tagsString}>

## ${version}

${changelogContent.trim()}

</Update>`;

  } catch (error) {
    console.error('Error generating changelog with LLM:', error.message);
    throw error;
  }
}

/**
 * Insert changelog into docs file
 */
function insertChangelogIntoFile(changelogPath, changelogContent) {
  if (!fs.existsSync(changelogPath)) {
    const dir = require('path').dirname(changelogPath);
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }

    const header = `---
title: "Terraform Provider"
description: "Release notes for Formal Terraform Provider"
---

`;
    fs.writeFileSync(changelogPath, header);
  }

  let content = fs.readFileSync(changelogPath, 'utf8');
  const frontmatterRegex = /^---\n[\s\S]*?\n---\n/;
  const match = frontmatterRegex.exec(content);

  if (!match) {
    throw new Error('Could not find frontmatter in MDX file');
  }

  const insertPosition = match.index + match[0].length;
  content = content.slice(0, insertPosition) +
            '\n' + changelogContent + '\n\n' +
            content.slice(insertPosition);

  fs.writeFileSync(changelogPath, content);
  console.log(`✓ Updated ${changelogPath}`);
}

/**
 * Main function
 */
async function main() {
  const args = process.argv.slice(2);

  // Parse arguments
  let mode = 'auto';
  let docsRepoPath = 'docs-repo';

  for (let i = 0; i < args.length; i++) {
    if (args[i] === '--mode' && i + 1 < args.length) {
      mode = args[i + 1];
      i++;
    } else if (args[i] === '--docs-repo-path' && i + 1 < args.length) {
      docsRepoPath = args[i + 1];
      i++;
    }
  }

  // Required environment variables
  const githubToken = process.env.GITHUB_TOKEN;
  const githubRepo = process.env.GITHUB_REPOSITORY || 'formalco/terraform-provider-formal';
  const openaiKey = process.env.OPENAI_API_KEY;

  if (!githubToken) {
    console.error('GITHUB_TOKEN environment variable is required');
    process.exit(1);
  }

  if (!openaiKey) {
    console.error('OPENAI_API_KEY environment variable is required');
    process.exit(1);
  }

  console.log('='.repeat(70));
  console.log('Generating Terraform Provider changelog');
  console.log(`Mode: ${mode}`);
  console.log('='.repeat(70));

  // Step 1: Find ALL missing versions needing changelogs
  console.log('\n[1/6] Finding all missing versions...');
  const missingVersionPairs = findAllMissingVersions(mode, docsRepoPath);

  if (missingVersionPairs.length === 0) {
    console.log('No versions need changelogs. Exiting.');
    process.exit(0);
  }

  console.log(`\nWill process ${missingVersionPairs.length} versions in order (oldest to newest):\n`);
  missingVersionPairs.forEach((v, i) => {
    console.log(`  ${i + 1}. ${v.version}${v.previousVersion ? ` (vs ${v.previousVersion})` : ' (first version)'}`);
  });
  console.log();

  const changelogPath = `${docsRepoPath}/docs/changelog/terraform-provider.mdx`;
  const allGeneratedChangelogs = [];
  const allPRsByVersion = [];

  // Process each version in sequence
  for (let i = 0; i < missingVersionPairs.length; i++) {
    const { version, previousVersion } = missingVersionPairs[i];

    console.log('\n' + '━'.repeat(70));
    console.log(`Processing version ${i + 1}/${missingVersionPairs.length}: ${version}`);
    console.log('━'.repeat(70));

    // Get commit date for the version
    const commitHash = getCommitForTag(version);
    const commitDate = exec(`TZ=UTC git show -s --format=%cd --date=format-local:%Y-%m-%d ${commitHash}`);
    console.log(`Version ${version} date (UTC): ${commitDate}`);

    // Find commits between versions
    console.log('\n[2/6] Finding commits between versions...');
    const commits = findCommitsBetweenVersions(version, previousVersion);
    console.log(`Found ${commits.length} commits`);

    // Get PR details
    console.log('\n[3/6] Fetching PR details...');
    const prDetails = await getPRDetailsFromCommits(commits, githubRepo, githubToken);

    // Skip versions with no changes
    if (prDetails.length === 0) {
      console.log(`⏭️  No PRs with 'provider' label found for version ${version}, skipping`);
      console.log(`✅ Version ${version} skipped (no changes)\n`);
      continue;
    }

    // Store PR information for this version
    allPRsByVersion.push({
      version,
      prs: prDetails.map(pr => ({ number: pr.number, title: pr.title, url: pr.url }))
    });

    // Generate changelog
    console.log('\n[4/6] Generating changelog with LLM...');
    const changelogContent = await generateChangelogWithLLM(
      version,
      prDetails,
      openaiKey,
      commitDate
    );

    console.log('\n' + '-'.repeat(70));
    console.log(`Generated Changelog for ${version}:`);
    console.log('-'.repeat(70));
    console.log(changelogContent);
    console.log('-'.repeat(70));

    allGeneratedChangelogs.push(changelogContent);
    console.log(`✅ Version ${version} complete!\n`);
  }

  // Check if we actually generated any changelogs
  if (allGeneratedChangelogs.length === 0) {
    console.log('\n' + '='.repeat(70));
    console.log('ℹ️  All versions were skipped (no changes with provider label)');
    console.log('='.repeat(70));
    process.exit(0);
  }

  // Step 6: Insert all changelogs at once
  console.log('\n[6/6] Inserting changelogs into docs...');
  const fullChangelog = allGeneratedChangelogs.join('\n\n');
  insertChangelogIntoFile(changelogPath, fullChangelog);

  // Output version range and PR list
  const firstVersion = missingVersionPairs[0].version;
  const lastVersion = missingVersionPairs[missingVersionPairs.length - 1].version;
  const versionRange = firstVersion === lastVersion ? firstVersion : `${firstVersion} - ${lastVersion}`;

  // Generate PR list markdown
  let prListMarkdown = '';
  for (const { version, prs } of allPRsByVersion) {
    prListMarkdown += `\n### ${version}\n`;
    for (const pr of prs) {
      prListMarkdown += `- [#${pr.number}](${pr.url}): ${pr.title}\n`;
    }
  }

  // Write outputs to file for GitHub Actions
  if (process.env.GITHUB_OUTPUT) {
    fs.appendFileSync(process.env.GITHUB_OUTPUT, `version=${versionRange}\n`);
    fs.appendFileSync(process.env.GITHUB_OUTPUT, `versions_count=${allGeneratedChangelogs.length}\n`);
    fs.appendFileSync(process.env.GITHUB_OUTPUT, `pr_list=${prListMarkdown.replace(/\n/g, '%0A')}\n`);
  }

  console.log('\n' + '='.repeat(70));
  console.log(`✅ Successfully generated ${allGeneratedChangelogs.length} changelogs!`);
  console.log(`   Version range: ${versionRange}`);
  console.log('='.repeat(70));
}

main().catch(error => {
  console.error('Fatal error:', error);
  process.exit(1);
});
